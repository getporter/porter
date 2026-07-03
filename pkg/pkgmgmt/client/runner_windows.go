//go:build windows

package client

import "os/exec"

// configureGracefulShutdown is a no-op on Windows; the default
// force-kill/TerminateProcess behaviour from exec.CommandContext is kept.
func configureGracefulShutdown(cmd *exec.Cmd) {}
