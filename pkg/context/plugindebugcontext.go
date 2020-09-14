package context

import (
	"os"
	"strconv"
)

type PluginDebugContext struct {
	DebugPlugins           bool
	RunPlugInInDebugger    string
	PlugInWorkingDirectory string
	DebuggerPort           string
}

func NewPluginDebugContext() *PluginDebugContext {

	runPlugInInDebugger := os.Getenv("PORTER_RUN_PLUGIN_IN_DEBUGGER")
	debugWorkingDirectory := os.Getenv("PORTER_PLUGIN_WORKING_DIRECTORY")
	debuggerPort := os.Getenv("PORTER_DEBUGGER_PORT")

	port := "2345"
	if _, err := strconv.ParseInt(debuggerPort, 10, 16); err != nil {
		port = debuggerPort
	}

	return &PluginDebugContext{
		RunPlugInInDebugger:    runPlugInInDebugger,
		DebuggerPort:           port,
		PlugInWorkingDirectory: debugWorkingDirectory,
	}
}
