package pkgmgmt

import (
	"errors"
	"fmt"
)

type UninstallOptions struct {
	Name string
}

func (o *UninstallOptions) Validate(args []string) error {
	switch len(args) {
	case 0:
		return errors.New("no name was specified")
	case 1:
		if err := ValidateName(args[0]); err != nil {
			return err
		}
		o.Name = args[0]
		return nil
	default:
		return fmt.Errorf("only one positional argument may be specified, the name, but multiple were received: %s", args)

	}
}
