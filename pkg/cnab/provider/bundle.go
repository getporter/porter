package cnabprovider

import (
	"github.com/deislabs/cnab-go/bundle"
	"github.com/deislabs/duffle/pkg/loader"
	"github.com/pkg/errors"
)

// TODO: Export everything in this file from duffle cmd/duffle/pull.go
var ErrNotSigned = errors.New("bundle is not signed")

func (d *Duffle) LoadBundle(bundleFile string, insecure bool) (*bundle.Bundle, error) {
	// TODO: once we support secure bundles we need more logic here (it's in duffle but I didn't copy it)
	// I'm hoping we've gotten this code exported from duffle by then though
	if !insecure {
		return nil, errors.New("secure bundles not implemented")
	}

	l := loader.New()

	// Issue #439: Errors that come back from the loader can be
	// pretty opaque.
	bun, err := l.Load(bundleFile)

	if err != nil {
		if err.Error() == "no signature block in data" {
			return bun, ErrNotSigned
		}
		// Dear Go, Y U NO TERNARY, kthxbye
		secflag := "secure"
		if insecure {
			secflag = "insecure"
		}
		return bun, errors.Wrapf(err, "cannot load %s bundle", secflag)
	}

	return bun, nil
}
