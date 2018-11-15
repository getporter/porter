package helm

import (
	"bufio"
	"io/ioutil"

	"github.com/deislabs/porter/pkg/context"

	"github.com/pkg/errors"
)

// Helm is the logic behind the helm mixin
type Mixin struct {
	*context.Context
}

// New helm mixin client, initialized with useful defaults.
func New() *Mixin {
	return &Mixin{
		Context: context.New(),
	}
}

func (m *Mixin) getPayloadData() ([]byte, error) {
	reader := bufio.NewReader(m.In)
	data, err := ioutil.ReadAll(reader)
	return data, errors.Wrap(err, "could not read the payload from STDIN")
}
