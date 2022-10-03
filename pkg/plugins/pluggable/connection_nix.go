//go:build !windows

package pluggable

import (
	"os"
	"syscall"
)

// IsPluginRunning indicates is the plugin process is still running.
// Used for testing only.
func (c *PluginConnection) IsPluginRunning() bool {
	if c.pluginCmd == nil || c.pluginCmd.Process == nil {
		return false
	}

	// We are only defining this function for linux/darwin because windows doesn't have SIGCONT
	err := c.pluginCmd.Process.Signal(os.Signal(syscall.SIGCONT))
	return err == nil
}
