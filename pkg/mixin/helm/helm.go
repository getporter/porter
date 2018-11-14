package helm

import (
	"bufio"
	"io"
	"io/ioutil"

	"github.com/pkg/errors"
)

// Helm is the logic behind the helm mixin
type Mixin struct {
	In  io.Reader
	Out io.Writer
	Err io.Writer
}

func (m *Mixin) getPayloadData() ([]byte, error) {
	reader := bufio.NewReader(m.In)
	data, err := ioutil.ReadAll(reader)
	return data, errors.Wrap(err, "could not read the payload from STDIN")
}
