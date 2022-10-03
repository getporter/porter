package cnabtooci

import (
	"errors"
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	"github.com/stretchr/testify/assert"
)

func TestErrNotFound_Is(t *testing.T) {
	err := ErrNotFound{Reference: cnab.MustParseOCIReference("example/bundle:v1.0.0")}

	assert.True(t, errors.Is(err, ErrNotFound{}))
}
