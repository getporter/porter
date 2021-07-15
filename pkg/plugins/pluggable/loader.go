package pluggable

import (
	"bufio"
	"bytes"
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

type InternalPluginHandler func(key string, config interface{}) (protocol plugins.Plugin, err error)

func NewPluginLoader(c *config.Config, createInternalPlugin InternalPluginHandler) *PluginLoader {
	return &PluginLoader{
		Config:               c,
		createInternalPlugin: createInternalPlugin,
	}
}

// Load a plugin, returning the plugin's interface which the caller must then cast to
// the typed interface, a cleanup function to stop the plugin when finished communicating with it,
// and an error if the plugin could not be loaded.
func (l *PluginLoader) Load(pluginType PluginTypeConfig) (interface{}, func(), error) {
	err := l.selectPlugin(pluginType)
	if err != nil {
		return nil, nil, err
	}

	l.SelectedPluginKey.Interface = pluginType.Interface

	if l.DebugPlugins {
		fmt.Fprintf(l.Err, "Resolved %s plugin to %s\n", pluginType.Interface, l.SelectedPluginKey)
		if l.SelectedPluginConfig != nil {
			fmt.Fprintf(l.Err, "Resolved plugin config: \n %#v\n", l.SelectedPluginConfig)
		}
	}

	var pluginCommand *exec.Cmd
	if l.SelectedPluginKey.IsInternal {
		plugin, err := l.createInternalPlugin(l.SelectedPluginKey.String(), l.SelectedPluginConfig)
		if err != nil {
			return nil, func() {}, err
		}

		return plugin, func() { plugin.Close() }, plugin.Connect()
	}

	pluginPath, err := l.GetPluginPath(l.SelectedPluginKey.Binary)
	if err != nil {
		return nil, nil, err
	}

	pluginCommand = l.NewCommand(pluginPath, "run", l.SelectedPluginKey.String())

	configReader, err := l.readPluginConfig()
	if err != nil {
		return nil, nil, err
	}

	pluginCommand.Stdin = configReader

	// Explicitly set PORTER_HOME for the plugin
	pluginCommand.Env = l.Environ()

	if l.DebugPlugins {
		fmt.Fprintln(l.Err, strings.Join(pluginCommand.Args, " "))
	}

	pluginOutput := bytes.NewBufferString("")
	logger := hclog.New(&hclog.LoggerOptions{
		Name:       "porter",
		Output:     pluginOutput,
		Level:      hclog.Debug,
		JSONFormat: true,
	})

	if l.DebugPlugins {
		logger.SetLevel(hclog.Info)

		go l.logPluginMessages(pluginOutput)
	}

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
	}

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		cleanup()
		if stderr := errbuf.String(); stderr != "" {
			err = errors.Wrap(errors.New(stderr), err.Error())
		}
		return nil, nil, errors.Wrapf(err, "could not connect to the %s plugin", l.SelectedPluginKey)
	}

	cleanup, err = l.setUpDebugger(client, cleanup)
	if err != nil {
		return nil, nil, errors.Wrap(err, "could not set up debugger for plugin")
	}

	// Request the plugin
	raw, err := rpcClient.Dispense(pluginType.Interface)
	if err != nil {
		cleanup()
		return nil, nil, errors.Wrapf(err, "could not connect to the %s plugin", l.SelectedPluginKey)
	}

	return raw, cleanup, nil
}

func (l *PluginLoader) setUpDebugger(client *plugin.Client, cleanup func()) (func(), error) {
	debugContext := l.Context.PlugInDebugContext
	if len(debugContext.RunPlugInInDebugger) > 0 && strings.ToLower(l.SelectedPluginKey.String()) == strings.TrimSpace(strings.ToLower(debugContext.RunPlugInInDebugger)) {
		if !isDelveInstalled() {
			return cleanup, errors.New("Delve needs to be installed to debug plugins")
		}
		listen := fmt.Sprintf("--listen=127.0.0.1:%s", debugContext.DebuggerPort)
		if len(debugContext.PlugInWorkingDirectory) == 0 {
			return cleanup, errors.New("Plugin Working Directory is required for debugging")
		}
		wd := fmt.Sprintf("--wd=%s", debugContext.PlugInWorkingDirectory)
		pid := client.ReattachConfig().Pid
		dlvCmd := exec.Command("dlv", "attach", strconv.Itoa(pid), "--headless=true", "--api-version=2", "--log", listen, "--accept-multiclient", wd)
		dlvCmd.Stderr = os.Stderr
		dlvCmd.Stdout = os.Stdout

		err := dlvCmd.Start()
		if err != nil {
			return cleanup, errors.Wrap(err, "Error starting dlv")
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
			return cleanup, errors.Wrap(err, "dlv exited unexpectedly")
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

func (l *PluginLoader) logPluginMessages(pluginOutput io.Reader) {
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
			// plaintext log
			fmt.Fprintln(l.Err, line)
		} else {
			// print just the message from the json log
			fmt.Fprintln(l.Err, pluginLog["@message"])
		}
	}
}

// selectPlugin picks the plugin to use and loads its configuration.
func (l *PluginLoader) selectPlugin(cfg PluginTypeConfig) error {
	l.SelectedPluginKey = nil
	l.SelectedPluginConfig = nil

	var pluginKey string

	defaultStore := cfg.GetDefaultPluggable(l.Config)
	if defaultStore != "" {
		is, err := cfg.GetPluggable(l.Config, defaultStore)
		if err != nil {
			return err
		}
		pluginKey = is.GetPluginSubKey()
		l.SelectedPluginConfig = is.GetConfig()
	}

	// If there isn't a specific plugin configured for this plugin type, fall back to the default plugin for this type
	if pluginKey == "" {
		pluginKey = cfg.GetDefaultPlugin(l.Config)
	}

	key, err := plugins.ParsePluginKey(pluginKey)
	if err != nil {
		return err
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
