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

	"get.porter.sh/porter/pkg/tracing"
	"go.opentelemetry.io/otel/attribute"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/plugins"
	hclog "github.com/hashicorp/go-hclog"
	plugin "github.com/hashicorp/go-plugin"
	"github.com/pkg/errors"
)

// PluginLoader handles finding, configuring and loading porter plugins.
type PluginLoader struct {
	*config.Config

	SelectedPluginKey    *plugins.PluginKey
	SelectedPluginConfig interface{}

	// called for internal plugins (starts with porter.*) to create an instance of
	// the plugin without going through a separate binary.
	createInternalPlugin InternalPluginHandler
}

type InternalPluginHandler func(ctx context.Context, key string, config interface{}) (protocol plugins.Plugin, err error)

func NewPluginLoader(c *config.Config, createInternalPlugin InternalPluginHandler) *PluginLoader {
	return &PluginLoader{
		Config:               c,
		createInternalPlugin: createInternalPlugin,
	}
}

// Load a plugin, returning the plugin's interface which the caller must then cast to
// the typed interface, a cleanup function to stop the plugin when finished communicating with it,
// and an error if the plugin could not be loaded.
func (l *PluginLoader) Load(ctx context.Context, pluginType PluginTypeConfig) (interface{}, func(), error) {
	ctx, span := tracing.StartSpan(ctx,
		attribute.String("plugin-interface", pluginType.Interface),
		attribute.String("plugin-protocol-version", fmt.Sprintf("%v", pluginType.ProtocolVersion)))
	defer span.EndSpan()

	err := l.selectPlugin(ctx, pluginType)
	if err != nil {
		return nil, nil, err
	}

	l.SelectedPluginKey.Interface = pluginType.Interface

	var pluginCommand *exec.Cmd
	if l.SelectedPluginKey.IsInternal {
		span.Debug("Selected plugin is internal")

		intPlugin, err := l.createInternalPlugin(ctx, l.SelectedPluginKey.String(), l.SelectedPluginConfig)
		if err != nil {
			return nil, func() {}, span.Error(err)
		}

		return intPlugin, func() { intPlugin.Close(ctx) }, intPlugin.Connect(ctx)
	}

	pluginPath, err := l.GetPluginPath(l.SelectedPluginKey.Binary)
	if err != nil {
		return nil, nil, span.Error(err)
	}
	span.SetAttributes(attribute.String("plugin-path", pluginPath))

	pluginCommand = l.NewCommand(pluginPath, "run", l.SelectedPluginKey.String())

	configReader, err := l.readPluginConfig()
	if err != nil {
		return nil, nil, span.Error(err)
	}

	pluginCommand.Stdin = configReader

	pluginOutput := bytes.NewBufferString("")
	logger := hclog.New(&hclog.LoggerOptions{
		Name:       "porter",
		Output:     pluginOutput,
		Level:      hclog.Debug,
		JSONFormat: true,
	})

	pluginCtx, pluginLogger := l.NewRootLogger(ctx, l.SelectedPluginKey.String(), "RunPlugin")
	go l.logPluginMessages(pluginCtx, pluginOutput)

	pluginTypes := map[string]plugin.Plugin{
		pluginType.Interface: pluginType.Plugin,
	}

	var errbuf bytes.Buffer
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: plugin.HandshakeConfig{
			ProtocolVersion:  pluginType.ProtocolVersion,
			MagicCookieKey:   plugins.HandshakeConfig.MagicCookieKey,
			MagicCookieValue: plugins.HandshakeConfig.MagicCookieValue,
		},
		Plugins: pluginTypes,
		Cmd:     pluginCommand,
		Logger:  logger,
		Stderr:  &errbuf,
	})
	cleanup := func() {
		client.Kill()
		pluginLogger.EndSpan()
		pluginLogger.Close()
	}

	// Connect via RPC
	span.Debug("Connecting to plugin", attribute.String("plugin-command", strings.Join(pluginCommand.Args, " ")))
	rpcClient, err := client.Client()
	if err != nil {
		cleanup()
		if stderr := errbuf.String(); stderr != "" {
			err = errors.Wrap(errors.New(stderr), err.Error())
		}
		return nil, nil, span.Error(errors.Wrapf(err, "could not connect to the %s plugin", l.SelectedPluginKey))
	}

	cleanup, err = l.setUpDebugger(ctx, client, cleanup)
	if err != nil {
		return nil, nil, span.Error(errors.Wrap(err, "could not set up debugger for plugin"))
	}

	// Request the plugin
	raw, err := rpcClient.Dispense(pluginType.Interface)
	if err != nil {
		cleanup()
		return nil, nil, span.Error(errors.Wrapf(err, "could not connect to the %s plugin", l.SelectedPluginKey))
	}

	return raw, cleanup, nil
}

func (l *PluginLoader) setUpDebugger(ctx context.Context, client *plugin.Client, cleanup func()) (func(), error) {
	log := tracing.LoggerFromContext(ctx)

	debugContext := l.Context.PlugInDebugContext
	if len(debugContext.RunPlugInInDebugger) > 0 && strings.ToLower(l.SelectedPluginKey.String()) == strings.TrimSpace(strings.ToLower(debugContext.RunPlugInInDebugger)) {
		if !isDelveInstalled() {
			return cleanup, log.Errorf("Delve needs to be installed to debug plugins")
		}
		listen := fmt.Sprintf("--listen=127.0.0.1:%s", debugContext.DebuggerPort)
		if len(debugContext.PlugInWorkingDirectory) == 0 {
			return cleanup, log.Errorf("Plugin Working Directory is required for debugging")
		}
		wd := fmt.Sprintf("--wd=%s", debugContext.PlugInWorkingDirectory)
		pid := client.ReattachConfig().Pid
		dlvCmd := exec.Command("dlv", "attach", strconv.Itoa(pid), "--headless=true", "--api-version=2", "--log", listen, "--accept-multiclient", wd)
		dlvCmd.Stderr = os.Stderr
		dlvCmd.Stdout = os.Stdout

		err := dlvCmd.Start()
		if err != nil {
			return cleanup, log.Errorf("Error starting dlv: %w", err)
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
			return cleanup, log.Errorf("dlv exited unexpectedly: %w", err)
		default:
		}

		newcleanup := func() {
			if dlvCmd.Process != nil {
				select {
				case err = <-dlvCmdTerminated:
				default:
					_ = dlvCmd.Process.Kill()
				}
			}
			cleanup()
		}
		return newcleanup, nil

	}
	return cleanup, nil
}

func (l *PluginLoader) logPluginMessages(ctx context.Context, pluginOutput io.Reader) {
	log := tracing.LoggerFromContext(ctx)

	r := bufio.NewReader(pluginOutput)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			continue
		}
		if line == "" {
			return
		}

		var pluginLog map[string]interface{}
		err = json.Unmarshal([]byte(line), &pluginLog)
		if err != nil {
			log.Error(err)
			continue
		}

		msg, ok := pluginLog["@message"].(string)
		if !ok {
			continue
		}

		switch pluginLog["@level"] {
		case hclog.Error:
			log.Errorf(msg)
		case hclog.Warn:
			log.Warn(msg)
		case hclog.Info:
			log.Info(msg)
		default:
			log.Debug(msg)
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
		span.Debug("Selected default plugin", attribute.String("plugin-key", pluginKey))
		pluginKey = cfg.GetDefaultPlugin(l.Config)
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
