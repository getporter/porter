//go:build !windows

package portercontext

import "syscall"

func (c *TestContext) DisableUmask() {
	// HACK: When running tests that check file permissions, if umask is set we can't validate the permissions we set
	// Turn off umask when running integration tests against the OS filesystem
	syscall.Umask(0000)
}
