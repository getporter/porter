package cnabprovider

import (
	"github.com/deislabs/cnab-go/bundle"
	"github.com/deislabs/cnab-go/bundle/loader"
	"github.com/pkg/errors"
)

var ErrNotSigned = errors.New("bundle is not signed")

func (d *Runtime) LoadBundle(bundleFile string, insecure bool) (*bundle.Bundle, error) {
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
