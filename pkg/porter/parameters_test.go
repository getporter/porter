package porter

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDisplayValuesSort(t *testing.T) {
	v := DisplayValues{
		{Name: "b"},
		{Name: "c"},
		{Name: "a"},
	}

	sort.Sort(v)

	assert.Equal(t, "a", v[0].Name)
	assert.Equal(t, "b", v[1].Name)
	assert.Equal(t, "c", v[2].Name)
}
