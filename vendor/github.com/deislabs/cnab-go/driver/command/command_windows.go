// +build windows

package command

import (
	"os"
	"os/exec"
)

// CheckDriverExists checks to see if the named driver exists
func (d *Driver) CheckDriverExists() bool {
	cmd := exec.Command("where", d.cliName())
	cmd.Env = os.Environ()
	if err := cmd.Run(); err != nil {
		return false
	}

	return true
}
