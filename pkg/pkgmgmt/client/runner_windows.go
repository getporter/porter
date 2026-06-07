//go:build windows

package client

import "os/exec"

// configureGracefulShutdown is a no-op on Windows where SIGTERM is not a real
// process signal; the default SIGKILL behaviour from exec.CommandContext is kept.
func configureGracefulShutdown(cmd *exec.Cmd) {}
