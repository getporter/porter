//go:build windows

package portercontext

func (c *TestContext) DisableUmask() {
	// Windows doesn't have umask
}
