//go:build windows
// +build windows

package pluggable

import (
	"os"
	"os/exec"
)

func isDelveInstalled() bool {
	cmd := exec.Command("where", "dlv")
	cmd.Env = os.Environ()
	if err := cmd.Run(); err != nil {
		return false
	}

	return true
}
