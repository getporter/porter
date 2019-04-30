package godirwalk

import (
	"path/filepath"
)

// The Windows version uses EvalSymlinks to remove the symlinks, since there
// are problems with Stat()ing symlinks on Windows (apparently it can fail to
// recognize invalid links in the path, for example).
func evalSymlinksHelper(pathname string) (string, error) {
	return filepath.EvalSymlinks(pathname)
}
