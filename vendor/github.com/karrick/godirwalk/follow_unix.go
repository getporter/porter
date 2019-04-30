// +build darwin freebsd linux netbsd openbsd

package godirwalk

// On unix we don't need to fixup pathnames with symlinks when doing a Stat().
func evalSymlinksHelper(pathname string) (string, error) {
	return pathname, nil
}
