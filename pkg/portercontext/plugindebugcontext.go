package portercontext

import (
	"strconv"
)

type PluginDebugContext struct {
	RunPlugInInDebugger    string
	PlugInWorkingDirectory string
	DebuggerPort           string
}

func NewPluginDebugContext(c *Context) *PluginDebugContext {

	runPlugInInDebugger := c.Getenv("PORTER_RUN_PLUGIN_IN_DEBUGGER")
	debugWorkingDirectory := c.Getenv("PORTER_PLUGIN_WORKING_DIRECTORY")
	debuggerPort := c.Getenv("PORTER_DEBUGGER_PORT")

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
