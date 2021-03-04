package yaml

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mymap map[string]interface{}

func (m *mymap) UnmarshalYAML(unmarshal func(interface{}) error) error {
	raw, err := UnmarshalMap(unmarshal)
	if err != nil {
		return err
	}

	*m = raw
	return nil
}

func TestUnmarshalMap(t *testing.T) {
	var m mymap

	testYaml := `a:
  1: one
  two:
    three: 3
`
	err := Unmarshal([]byte(testYaml), &m)
	require.NoError(t, err)

	wantM := mymap{
		"a": map[string]interface{}{
			"1": "one",
			"two": map[string]interface{}{
				"three": 3,
			},
		},
	}
	assert.Equal(t, wantM, m)
}
