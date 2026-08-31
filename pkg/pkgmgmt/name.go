package pkgmgmt

import (
	"fmt"
	"regexp"
)

// packageNameRe is the set of characters allowed in a mixin or plugin name.
// Names are used directly in install URLs and filesystem paths, so they are
// restricted to lowercase alphanumerics, dashes, and underscores.
var packageNameRe = regexp.MustCompile(`^[a-z0-9_-]+$`)

// ValidateName checks that a mixin or plugin name follows Porter's naming
// rules: lowercase letters, numbers, dashes, and underscores only.
func ValidateName(name string) error {
	if !packageNameRe.MatchString(name) {
		return fmt.Errorf("invalid name %q: must contain only lowercase letters, numbers, dashes, and underscores", name)
	}
	return nil
}
