package storage

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOutputs_Sort(t *testing.T) {
	o := NewOutputs([]Output{
		{Name: "a"},
		{Name: "c"},
		{Name: "b"},
	})

	sort.Sort(o)

	wantNames := []string{"a", "b", "c"}
	gotNames := make([]string, 0, 3)
	for i := 0; i < o.Len(); i++ {
		output, ok := o.GetByIndex(i)
		require.True(t, ok, "GetByIndex failed")
		gotNames = append(gotNames, output.Name)
	}

	assert.Equal(t, wantNames, gotNames)
}
