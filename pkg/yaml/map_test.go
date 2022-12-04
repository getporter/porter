package yaml

import (
	"os"
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

	testYaml, err := os.ReadFile("testdata/custom.yaml")
	require.NoError(t, err, "could not read testdata")

	err = Unmarshal([]byte(testYaml), &m)
	require.NoError(t, err, "Unmarshal failed")

	wantM := mymap{
		"root": map[string]interface{}{
			"array": []interface{}{
				map[string]interface{}{
					"map": map[string]interface{}{
						"a":          "a",
						"1":          1,
						"true":       true,
						"yes":        "yes",
						"null":       nil,
						"typesArray": []interface{}{"a", 1, true, "yes", nil, []interface{}{1, 2}},
						"map": map[string]interface{}{
							"a":          "a",
							"1":          1,
							"true":       true,
							"yes":        "yes",
							"null":       nil,
							"typesArray": []interface{}{"a", 1, true, "yes", nil, []interface{}{1, 2}},
						},
					},
				},
			},
		},
	}
	assert.Equal(t, wantM, m)
}
