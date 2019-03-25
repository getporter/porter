package cnabprovider

import (
	_ "github.com/deislabs/duffle/pkg/action"
	"github.com/deislabs/porter/pkg/context"
)

type Duffle struct {
	*context.Context
}

func NewDuffle(c *context.Context) *Duffle {
	return &Duffle{
		Context: c,
	}
}

func (d *Duffle) Install() error {
	return nil
}
