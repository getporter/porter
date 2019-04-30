package godirwalk

import (
	"os"
)

// symlinkDirHelper accepts a pathname and Dirent for a symlink, and determine
// what the ModeType bits are for what the symlink ultimately points to (such
// as whether it is a directory or not)
func symlinkDirHelper(pathname string, de *Dirent, o *Options) (bool, error) {
	// Only need to Stat entry if platform did not already have os.ModeDir
	// set, such as would be the case for unix like operating systems.
	// (This guard eliminates extra os.Stat check on Windows.)
	if !de.IsDir() {
		// Remove symlinks from the pathname (replaced with a more direct path)
		pathname, err := evalSymlinksHelper(pathname)
		if err != nil {
			if action := o.ErrorCallback(pathname, err); action == SkipNode {
				err = nil
				return true, nil
			}
			return false, err
		}

		fi, err := os.Stat(pathname)
		if err != nil {
			if action := o.ErrorCallback(pathname, err); action == SkipNode {
				return true, nil
			}
			return false, err
		}
		de.modeType = fi.Mode() & os.ModeType
	}
	return false, nil
}
