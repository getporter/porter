package porter

import (
	"fmt"
	"io"

	"github.com/deislabs/porter/pkg"
)

func PrintVersion(w io.Writer) {
	fmt.Fprintf(w, "porter %s (%s)\n", pkg.Version, pkg.Commit)
}
