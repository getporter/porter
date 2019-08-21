package exec

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlags_Sort(t *testing.T) {
	flags := Flags{
		NewFlag("b", "1"),
		NewFlag("a", "2"),
		NewFlag("c", "3"),
	}

	sort.Sort(flags)

	assert.Equal(t, "a", flags[0].Name)
	assert.Equal(t, "b", flags[1].Name)
	assert.Equal(t, "c", flags[2].Name)
}

func TestFlag_ToSlice(t *testing.T) {
	t.Run("short flag", func(t *testing.T) {
		f := NewFlag("f", "abc")
		args := f.ToSlice()
		assert.Equal(t, []string{"-f", "abc"}, args)
	})

	t.Run("long flag", func(t *testing.T) {
		f := NewFlag("full", "abc")
		args := f.ToSlice()
		assert.Equal(t, []string{"--full", "abc"}, args)
	})

	t.Run("valueless flag", func(t *testing.T) {
		f := NewFlag("l")
		args := f.ToSlice()
		assert.Equal(t, []string{"-l"}, args)
	})
}

func TestFlags_ToSlice(t *testing.T) {
	flags := Flags{
		NewFlag("bull", "2"),
		NewFlag("a", "1"),
	}

	args := flags.ToSlice()

	// Flags should be sorted and sliced up on a platter
	assert.Equal(t, []string{"-a", "1", "--bull", "2"}, args)
}
