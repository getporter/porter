package cnabtooci

import (
	"fmt"

	"get.porter.sh/porter/pkg/cnab"
)

// ErrNotFound represents when a bundle or image is not found.
type ErrNotFound struct {
	Reference cnab.OCIReference
}

func (e ErrNotFound) Is(target error) bool {
	_, ok := target.(ErrNotFound)
	return ok
}

func (e ErrNotFound) String() string {
	return e.Error()
}

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("reference not found (%s)", e.Reference)
}
