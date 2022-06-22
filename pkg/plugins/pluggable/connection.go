package pluggable

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/plugins"
	"get.porter.sh/porter/pkg/tracing"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/go-plugin"
	"github.com/pkg/errors"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc"
)

// PluginConnection represents a connection to a plugin.
// It wraps the hashicorp/go-plugin library.
type PluginConnection struct {
	// config is the porter configuration
	config *config.Config

	// key is the fully-qualified plugin key.
	// For example, porter.storage.mongodb
	key plugins.PluginKey

	// pluginType is the type of plugin we want to connect to.
	pluginType PluginTypeConfig

	// client is the plugin framework client used to manage the connection to the
	// plugin.
	client *plugin.Client

	// pluginCmd is command that manages the plugin process
	pluginCmd *exec.Cmd

	// pluginProtocol is a connection that supports the plugin protocol, such as
	// plugins.SecretsProtocol or plugins.StorageProtocol
	pluginProtocol interface{}

	// debugger is the optionally attached go debugger command.
	debugger *exec.Cmd

	// cancelLogCtx is the cancellation function for our go-routine that collects the plugin logs
	cancelLogCtx context.CancelFunc

	// logsWaitGroup is used to ensure that any go routines spawned by the plugin connection
	// complete when Close is called. Otherwise we can get into a race between us and when the logger is closed.
	logsWaitGroup sync.WaitGroup

	// logsWriter receives logs from the plugin's stdout.
	logsWriter *io.PipeWriter

	// logsReader reads the logs from the plugin.
	logsReader *io.PipeReader
}

func NewPluginConnection(c *config.Config, pluginType PluginTypeConfig, pluginKey plugins.PluginKey) *PluginConnection {
	return &PluginConnection{
		config:     c,
		pluginType: pluginType,
		key:        pluginKey,
	}
}

// String returns the plugin key name
func (c *PluginConnection) String() string {
	return c.key.String()
}

// Start establishes a connection to the plugin.
// * pluginCfg is the resolved plugin configuration section from the Porter config file
func (c *PluginConnection) Start(ctx context.Context, pluginCfg io.Reader) error {
	ctx, span := tracing.StartSpan(ctx,
		attribute.String("plugin-key", c.key.String()))
	defer span.EndSpan()

	// Create a command to run the plugin
	if c.key.IsInternal {
		porterPath, err := c.config.GetPorterPath()
		if err != nil {
			return errors.Wrap(err, "could not determine the path to the porter pluginProtocol")
		}

		c.pluginCmd = c.config.NewCommand(ctx, porterPath, "plugin", "run", c.key.String())
	} else {
		pluginPath, err := c.config.GetPluginPath(c.key.Binary)
		if err != nil {
			return span.Error(err)
		}
		span.SetAttributes(attribute.String("plugin-path", pluginPath))

		c.pluginCmd = c.config.NewCommand(ctx, pluginPath, "run", c.key.String())
	}
	span.SetAttributes(attribute.String("plugin-path", c.pluginCmd.Path))

	// Configure the command
	c.pluginCmd.Stdin = pluginCfg
	// The plugin doesn't read the config file, we pass in relevant plugin config to them directly
	// The remaining relevant config (e.g. logging, tracing) is set via env vars
	// Config files require using the plugins to resolve templated values, so we resolve once in Porter
	// and pass relevant resolved values to the plugins explicitly
	pluginConfigVars := c.config.ExportRemoteConfigAsEnvironmentVariables()
	c.pluginCmd.Env = append(c.pluginCmd.Env, pluginConfigVars...)

	// Pipe logs from the plugin and capture them
	c.setupLogCollector(ctx)

	var errbuf bytes.Buffer
	logger := hclog.New(&hclog.LoggerOptions{
		Name:       "porter",
		Output:     c.logsWriter,
		Level:      hclog.Debug,
		JSONFormat: true,
	})
	c.client = plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: plugin.HandshakeConfig{
			ProtocolVersion:  c.pluginType.ProtocolVersion,
			MagicCookieKey:   plugins.HandshakeConfig.MagicCookieKey,
			MagicCookieValue: plugins.HandshakeConfig.MagicCookieValue,
		},
		AllowedProtocols: []plugin.Protocol{
			// All v1 plugins use gRPC
			plugin.ProtocolGRPC,
			// Enable net/rpc so that we can talk to older plugins from before v1
			plugin.ProtocolNetRPC,
		},
		// Specify which plugin we want to connect to
		Plugins: map[string]plugin.Plugin{
			c.pluginType.Interface: c.pluginType.Plugin,
		},
		Cmd:          c.pluginCmd,
		Logger:       logger,
		Stderr:       &errbuf,
		StartTimeout: getPluginStartTimeout(),
		// Configure gRPC to propagate the span context so the plugin's traces
		// show up under the current span
		GRPCDialOptions: []grpc.DialOption{
			grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		},
	})

	// Start the plugin
	span.Debug("Connecting to plugin", attribute.String("plugin-command", strings.Join(c.pluginCmd.Args, " ")))
	rpcClient, err := c.client.Client(ctx)
	if err != nil {
		if stderr := errbuf.String(); stderr != "" {
			err = fmt.Errorf("could not connect to the %s plugin: %w: %s", c.key, err, stderr)
		}
		span.Error(err) // Emit the error before trying to close the connection
		c.Close(ctx)
		return err
	}

	err = c.setUpDebugger(ctx, c.client)
	if err != nil {
		err = span.Error(fmt.Errorf("could not set up debugger for plugin: %w", err))
		c.Close(ctx) // Emit the error before trying to close the connection
		return err
	}

	// Get a connection to the plugin
	c.pluginProtocol, err = rpcClient.Dispense(c.key.Interface)
	if err != nil {
		err = span.Error(fmt.Errorf("could not connect to the %s plugin: %w", c.key, err))
		c.Close(ctx) // Emit the error before trying to close the connection
		return err
	}

	span.SetAttributes(attribute.Int("negotiated-protocol-version", c.client.NegotiatedVersion()))

	return nil
}

// GetClient returns the raw connection to the pluginProtocol.
// This value should be cast to the plugin protocol interface,
// such as plugins.StorageProtocol or plugins.SecretsProtocol.
func (c *PluginConnection) GetClient() interface{} {
	return c.pluginProtocol
}

// Close releases the resources held by the plugin connection. Blocks until the
// plugin process closes. Pass a context to control the graceful shutdown of the
// plugin.
func (c *PluginConnection) Close(ctx context.Context) error {
	ctx, span := tracing.StartSpan(ctx,
		attribute.String("plugin-key", c.key.String()))
	defer span.EndSpan()

	var bigErr *multierror.Error
	if c.client != nil {
		ctx, cancel := context.WithTimeout(ctx, getPluginStopTimeout())
		defer cancel()

		// Stop the plugin process
		done := make(chan bool)
		go func() {
			// beware, this can block or deadlock
			c.client.Kill(ctx)
			done <- true
		}()
		select {
		case <-done:
			// plugin stopped as requested
			break
		case <-ctx.Done():
			// Stop being nice, cleanup the plugin process without any waiting or blocking
			span.Debugf("killing the plugin process: %s", ctx.Err())
			c.client.HardKill()
		}

		// Stop processing logs from the plugin and wait for the log collection routine to complete
		// This avoids a race where the log collector picks up a message but doesn't print it until
		// after we close the logfile. This ensures that everything releated to the plugins is released
		// when Close exits.
		c.cancelLogCtx()
		c.logsWriter.Close()
		c.logsReader.Close()
		c.logsWaitGroup.Wait()

		c.client = nil
	}

	if c.debugger != nil {
		if c.debugger.Process != nil {
			err := c.debugger.Process.Kill()
			bigErr = multierror.Append(bigErr, err)
		}
		c.debugger = nil
	}

	return bigErr.ErrorOrNil()
}

func getPluginStartTimeout() time.Duration {
	timeoutS := os.Getenv(PluginStartTimeoutEnvVar)
	if timeoutD, err := time.ParseDuration(timeoutS); err == nil {
		return timeoutD
	}
	return PluginStartTimeoutDefault
}

func getPluginStopTimeout() time.Duration {
	timeoutS := os.Getenv(PluginStopTimeoutEnvVar)
	if timeoutD, err := time.ParseDuration(timeoutS); err == nil {
		return timeoutD
	}
	return PluginStopTimeoutDefault
}

func (c *PluginConnection) setUpDebugger(ctx context.Context, client *plugin.Client) error {
	log := tracing.LoggerFromContext(ctx)

	debugContext := c.config.PlugInDebugContext
	if !(len(debugContext.RunPlugInInDebugger) > 0 && strings.ToLower(c.key.String()) == strings.TrimSpace(strings.ToLower(debugContext.RunPlugInInDebugger))) {
		return nil
	}

	if !isDelveInstalled() {
		return log.Error(errors.New("Delve needs to be installed to debug plugins"))
	}

	listen := fmt.Sprintf("--listen=127.0.0.1:%s", debugContext.DebuggerPort)
	if len(debugContext.PlugInWorkingDirectory) == 0 {
		return log.Error(errors.New("Plugin Working Directory is required for debugging"))
	}
	wd := fmt.Sprintf("--wd=%s", debugContext.PlugInWorkingDirectory)
	pid := client.ReattachConfig().Pid
	c.debugger = exec.CommandContext(ctx, "dlv", "attach", strconv.Itoa(pid), "--headless=true", "--api-version=2", "--log", listen, "--accept-multiclient", wd)
	c.debugger.Stderr = os.Stderr
	c.debugger.Stdout = os.Stdout
	err := c.debugger.Start()
	if err != nil {
		return log.Error(fmt.Errorf("Error starting dlv: %w", err))
	}
	return nil
}

// setupLogCollector kicks off a go routine to collect the plugin logs.
func (c *PluginConnection) setupLogCollector(ctx context.Context) {
	c.logsReader, c.logsWriter = io.Pipe()
	ctx, c.cancelLogCtx = context.WithCancel(ctx)

	c.logsWaitGroup.Add(1)
	go c.collectPluginLogs(ctx)
}

// Watch the pipe between porter and the plugin for messages, and log them in a span.
// We don't have a good way to associate them with the actual action in porter that triggered the plugin response
// The best way to get that information is to instrument the plugin itself. This is mainly a fallback mechanism to
// collect logs from an uninstrumented plugin.
func (c *PluginConnection) collectPluginLogs(ctx context.Context) {
	defer c.logsWaitGroup.Done()

	ctx, span := tracing.StartSpan(ctx, attribute.String("plugin-key", c.key.String()))
	defer span.EndSpan()

	r := bufio.NewReader(c.logsReader)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			line, err := r.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					return
				}
				return
			}
			if line == "" {
				return
			}

			var pluginLog map[string]interface{}
			err = json.Unmarshal([]byte(line), &pluginLog)
			if err != nil {
				continue
			}

			msg, ok := pluginLog["@message"].(string)
			if !ok {
				continue
			}

			switch pluginLog["@level"] {
			case hclog.Error:
				span.Error(fmt.Errorf(msg))
			case hclog.Warn:
				span.Warn(msg)
			case hclog.Info:
				span.Infof(msg)
			default:
				span.Debug(msg)
			}
		}
	}
}
