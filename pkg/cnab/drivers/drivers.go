package drivers

import (
	"get.porter.sh/porter/pkg/portercontext"
	"github.com/cnabio/cnab-go/driver"
	"github.com/cnabio/cnab-go/driver/command"
	"github.com/cnabio/cnab-go/driver/debug"
	"github.com/cnabio/cnab-go/driver/docker"
	"github.com/cnabio/cnab-go/driver/kubernetes"
	"github.com/pkg/errors"
)

// LookupDriver creates a driver by name.
//
// This replaces cnab-go's lookup function because cnab-go uses global process
// values, such as $PATH, instead of our context.
func LookupDriver(cxt *portercontext.Context, name string) (driver.Driver, error) {
	switch name {
	case "docker":
		return &docker.Driver{}, nil
	case "kubernetes", "k8s":
		return &kubernetes.Driver{}, nil
	case "debug":
		return &debug.Driver{}, nil
	default:
		// Drivers must be named "cnab-NAME" and be on the PATH
		if driverPath, ok := cxt.LookPath("cnab-" + name); ok {
			d := &command.Driver{Name: name}
			d.Path = driverPath
			return d, nil
		}

		return nil, errors.Errorf("unsupported driver or driver not found in PATH: %s", name)
	}
}
