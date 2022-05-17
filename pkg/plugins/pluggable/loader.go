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
	"time"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/plugins"
	"get.porter.sh/porter/pkg/tracing"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/pkg/errors"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc"
)

const (
	// PluginStartTimeoutDefault is the default amount of time to wait for a plugin
	// to start. Override with PluginStartTimeoutEnvVar.
	PluginStartTimeoutDefault = 1 * time.Second

	// PluginStopTimeoutDefault is the default amount of time to wait for a plugin
	// to stop (kill). Override with PluginStopTimeoutEnvVar.
	PluginStopTimeoutDefault = 100 * time.Millisecond

	// PluginStartTimeoutEnvVar is the environment variable used to override
	// PluginStartTimeoutDefault.
	PluginStartTimeoutEnvVar = "PORTER_PLUGIN_START_TIMEOUT"

	// PluginStopTimeoutEnvVar is the environment variable used to override
	// PluginStopTimeoutDefault.
	PluginStopTimeoutEnvVar = "PORTER_PLUGIN_STOP_TIMEOUT"
)

// PluginLoader handles finding, configuring and loading porter plugins.
type PluginLoader struct {
	*config.Config

	SelectedPluginKey    *plugins.PluginKey
	SelectedPluginConfig interface{}
}

func NewPluginLoader(c *config.Config) *PluginLoader {
	return &PluginLoader{
		Config: c,
	}
}

// PluginConnection represents a connection to a running plugin.
type PluginConnection struct {
	// Key is the fully-qualified plugin key.
	// For example, porter.storage.mongodb
	Key string

	// Client should be cast to the plugin protocol interface.
	Client interface{}

	// cleanup is called when the plugin connection is closed
	cleanup func()
}

// Close releases the resources held by the plugin connection.
func (c PluginConnection) Close() {
	// TODO: Move plugin connection into its own and move lots of load into it
	if c.cleanup != nil {
		c.cleanup()
		c.cleanup = nil
	}
}

// Load a plugin, returning the plugin's interface which the caller must then cast to
// the typed interface, a cleanup function to stop the plugin when finished communicating with it,
// and an error if the plugin could not be loaded.
func (l *PluginLoader) Load(ctx context.Context, pluginType PluginTypeConfig) (PluginConnection, error) {
	ctx, span := tracing.StartSpan(ctx,
		attribute.String("plugin-interface", pluginType.Interface),
		attribute.String("requested-protocol-version", fmt.Sprintf("%v", pluginType.ProtocolVersion)))
	defer span.EndSpan()

	err := l.selectPlugin(ctx, pluginType)
	if err != nil {
		return PluginConnection{}, err
	}

	// quick check to detect that we are running as porter, and not a plugin already
	if l.Context.IsInternalPlugin {
		err := fmt.Errorf("the internal plugin %s tried to load the %s plugin. Report this error to https://github.com/getporter/porter", l.Context.InternalPluginKey, l.SelectedPluginKey)
		return PluginConnection{}, span.Error(err)
	}

	l.SelectedPluginKey.Interface = pluginType.Interface
	span.SetAttributes(attribute.String("plugin-key", l.SelectedPluginKey.String()))

	var pluginCommand *exec.Cmd
	if l.SelectedPluginKey.IsInternal {
		porterPath, err := l.GetPorterPath()
		if err != nil {
			return PluginConnection{}, errors.Wrap(err, "could not determine the path to the porter client")
		}

		pluginCommand = l.NewCommand(ctx, porterPath, "plugin", "run", l.SelectedPluginKey.String())
	} else {
		pluginPath, err := l.GetPluginPath(l.SelectedPluginKey.Binary)
		if err != nil {
			return PluginConnection{}, span.Error(err)
		}
		span.SetAttributes(attribute.String("plugin-path", pluginPath))

		pluginCommand = l.NewCommand(ctx, pluginPath, "run", l.SelectedPluginKey.String())
	}
	span.SetAttributes(attribute.String("plugin-path", pluginCommand.Path))

	configReader, err := l.readPluginConfig()
	if err != nil {
		return PluginConnection{}, span.Error(err)
	}

	pluginCommand.Stdin = configReader

	// Pipe logs from the plugin and capture them
	logsReader, logsWriter := io.Pipe()
	logsCtx, cancelLogCtx := context.WithCancel(ctx)
	go l.logPluginMessages(logsCtx, l.SelectedPluginKey.String(), logsReader)

	var errbuf bytes.Buffer
	logger := hclog.New(&hclog.LoggerOptions{
		Name:       "porter",
		Output:     logsWriter,
		Level:      hclog.Debug,
		JSONFormat: true,
	})
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: plugin.HandshakeConfig{
			ProtocolVersion:  pluginType.ProtocolVersion,
			MagicCookieKey:   plugins.HandshakeConfig.MagicCookieKey,
			MagicCookieValue: plugins.HandshakeConfig.MagicCookieValue,
		},
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		// Specify which plugin we want to connect to
		Plugins: map[string]plugin.Plugin{
			pluginType.Interface: pluginType.Plugin,
		},
		Cmd:          pluginCommand,
		Logger:       logger,
		Stderr:       &errbuf,
		StartTimeout: getPluginStartTimeout(),
		// Configure gRPC to propagate the span context so the plugin's traces
		// show up under the current span
		GRPCDialOptions: []grpc.DialOption{
			grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		},
	})

	conn := PluginConnection{
		Key: l.SelectedPluginKey.String(),
		cleanup: func() {
			// Stop the plugin process
			done := make(chan bool)
			go func() {
				// beware, this can block or deadlock
				client.Kill(ctx)
				done <- true
			}()
			select {
			case <-done:
				// plugin stopped as requested
				break
			case <-time.After(getPluginStopTimeout()):
				// Stop being nice, cleanup the plugin process without any waiting or blocking
				span.Debug("plugin stop timeout was exceeded, killing the plugin process")
				client.HardKill()
			}

			// Stop processing logs from the plugin
			cancelLogCtx()

			// Close the pipe between the plugin and porter
			logsWriter.Close()
			logsReader.Close()
		},
	}

	// Start the plugin
	span.Debug("Connecting to plugin", attribute.String("plugin-command", strings.Join(pluginCommand.Args, " ")))
	rpcClient, err := client.Client(ctx)
	if err != nil {
		conn.Close()
		if stderr := errbuf.String(); stderr != "" {
			err = errors.Wrap(errors.New(stderr), err.Error())
		}
		return PluginConnection{}, span.Error(errors.Wrapf(err, "could not connect to the %s plugin", l.SelectedPluginKey))
	}

	err = l.setUpDebugger(ctx, client, &conn)
	if err != nil {
		conn.Close()
		return PluginConnection{}, span.Error(errors.Wrap(err, "could not set up debugger for plugin"))
	}

	// Get a connection to the plugin
	conn.Client, err = rpcClient.Dispense(pluginType.Interface)
	if err != nil {
		conn.Close()
		return PluginConnection{}, span.Error(errors.Wrapf(err, "could not connect to the %s plugin", l.SelectedPluginKey))
	}
	span.SetAttributes(attribute.Int("negotiated-protocol-version", client.NegotiatedVersion()))

	return conn, nil
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

func (l *PluginLoader) setUpDebugger(ctx context.Context, client *plugin.Client, conn *PluginConnection) error {
	log := tracing.LoggerFromContext(ctx)

	debugContext := l.Context.PlugInDebugContext
	if len(debugContext.RunPlugInInDebugger) > 0 && strings.ToLower(l.SelectedPluginKey.String()) == strings.TrimSpace(strings.ToLower(debugContext.RunPlugInInDebugger)) {
		if !isDelveInstalled() {
			return log.Error(errors.New("Delve needs to be installed to debug plugins"))
		}
		listen := fmt.Sprintf("--listen=127.0.0.1:%s", debugContext.DebuggerPort)
		if len(debugContext.PlugInWorkingDirectory) == 0 {
			return log.Error(errors.New("Plugin Working Directory is required for debugging"))
		}
		wd := fmt.Sprintf("--wd=%s", debugContext.PlugInWorkingDirectory)
		pid := client.ReattachConfig().Pid
		dlvCmd := exec.CommandContext(ctx, "dlv", "attach", strconv.Itoa(pid), "--headless=true", "--api-version=2", "--log", listen, "--accept-multiclient", wd)
		dlvCmd.Stderr = os.Stderr
		dlvCmd.Stdout = os.Stdout

		err := dlvCmd.Start()
		if err != nil {
			return log.Error(fmt.Errorf("Error starting dlv: %w", err))
		}
		dlvCmdTerminated := make(chan error)
		go func() {
			dlvCmdTerminated <- dlvCmd.Wait()
		}()

		// dlv attach does not fail immediately but is common (e.g. if plugin is compiled with ldflags that prevent debugging)
		// so pause here to make sure that it has attached correctly and not failed

		time.Sleep(2 * time.Second)

		select {
		case err = <-dlvCmdTerminated:
			return log.Error(fmt.Errorf("dlv exited unexpectedly: %w", err))
		default:
		}

		parentCleanup := conn.cleanup
		conn.cleanup = func() {
			if dlvCmd.Process != nil {
				select {
				case <-ctx.Done():
				case err = <-dlvCmdTerminated:
				default:
					_ = dlvCmd.Process.Kill()
				}
			}
			parentCleanup()
		}
		return nil

	}
	return nil
}

// Watch the pipe between porter and the plugin for messages, and log them in a span.
// We don't have a good way to associate them with the actual action in porter that triggered the plugin response
// The best way to get that information is to instrument the plugin itself. This is mainly a fallback mechanism to
// collect logs from an uninstrumented plugin.
func (l *PluginLoader) logPluginMessages(ctx context.Context, pluginKey string, pluginOutput io.Reader) {
	ctx, span := tracing.StartSpanWithName(ctx, "CollectPluginLogs", attribute.String(pluginKey, pluginKey))
	defer span.EndSpan()

	r := bufio.NewReader(pluginOutput)
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
				continue
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

// selectPlugin picks the plugin to use and loads its configuration.
func (l *PluginLoader) selectPlugin(ctx context.Context, cfg PluginTypeConfig) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	l.SelectedPluginKey = nil
	l.SelectedPluginConfig = nil

	var pluginKey string

	defaultStore := cfg.GetDefaultPluggable(l.Config)
	if defaultStore != "" {
		span.SetAttributes(attribute.String("default-plugin", defaultStore))

		is, err := cfg.GetPluggable(l.Config, defaultStore)
		if err != nil {
			return span.Error(err)
		}

		pluginKey = is.GetPluginSubKey()
		l.SelectedPluginConfig = is.GetConfig()
		if l.SelectedPluginConfig == nil {
			span.Debug("No plugin config defined")
		}
	}

	// If there isn't a specific plugin configured for this plugin type, fall back to the default plugin for this type
	if pluginKey == "" {
		pluginKey = cfg.GetDefaultPlugin(l.Config)
		span.Debug("Selected default plugin", attribute.String("plugin-key", pluginKey))
	} else {
		span.Debug("Selected configured plugin", attribute.String("plugin-key", pluginKey))
	}

	key, err := plugins.ParsePluginKey(pluginKey)
	if err != nil {
		return span.Error(err)
	}
	l.SelectedPluginKey = &key

	return nil
}

func (l *PluginLoader) readPluginConfig() (io.Reader, error) {
	if l.SelectedPluginConfig == nil {
		return &bytes.Buffer{}, nil
	}

	b, err := json.Marshal(l.SelectedPluginConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "could not marshal plugin config %#v", l.SelectedPluginConfig)
	}

	return bytes.NewBuffer(b), nil
}
