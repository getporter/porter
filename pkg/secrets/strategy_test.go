package secrets

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSet_Merge(t *testing.T) {
	set := Set{
		"first":  "first",
		"second": "second",
		"third":  "third",
	}

	is := assert.New(t)

	err := set.Merge(Set{})
	is.NoError(err)
	is.Len(set, 3)
	is.NotContains(set, "fourth")

	err = set.Merge(Set{"fourth": "fourth"})
	is.NoError(err)
	is.Len(set, 4)
	is.Contains(set, "fourth")

	err = set.Merge(Set{"second": "bis"})
	is.EqualError(err, `ambiguous value resolution: "second" is already present in base sets, cannot merge`)
}
