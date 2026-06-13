//go:build !windows

package client

import (
	"os/exec"
	"syscall"
	"time"
)

// configureGracefulShutdown sets up the command to send SIGTERM when the
// context is cancelled, waiting up to 30 seconds before falling back to SIGKILL.
// This allows bundle steps (e.g. terraform) to perform cleanup before exiting.
func configureGracefulShutdown(cmd *exec.Cmd) {
	cmd.WaitDelay = 30 * time.Second
	cmd.Cancel = func() error {
		if cmd.Process != nil {
			_ = cmd.Process.Signal(syscall.SIGTERM) // best-effort; process may have already exited
		}
		return nil
	}
}
