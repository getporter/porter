package mixin

import (
	"strings"

	"github.com/pkg/errors"
)

type UninstallOptions struct {
	Name string
}

func (o *UninstallOptions) Validate(args []string) error {
	switch len(args) {
	case 0:
		return errors.Errorf("no mixin name was specified")
	case 1:
		o.Name = strings.ToLower(args[0])
		return nil
	default:
		return errors.Errorf("only one positional argument may be specified, the mixin name, but multiple were received: %s", args)

	}
}
