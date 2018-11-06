package porter

import (
	"io"
)

// Porter is the logic behind the porter client
type Porter struct {
	Out io.Writer
}
