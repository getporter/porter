package builder

import (
	"io/ioutil"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v2"
)

var testStep = TestStep{}

func TestFlags_UnmarshalYAML(t *testing.T) {
	b, err := ioutil.ReadFile("testdata/flags-input.yaml")
	require.NoError(t, err, "could not read the input file")

	var flags Flags
	err = yaml.Unmarshal(b, &flags)
	require.NoError(t, err, "could not unmarshal the flags")

	assert.Contains(t, flags, NewFlag("int", "1"))
	assert.Contains(t, flags, NewFlag("bool", "true"))
	assert.Contains(t, flags, NewFlag("string", "abc"))
	assert.Contains(t, flags, NewFlag("empty"))
	assert.Contains(t, flags, NewFlag("repeated", "FOO=BAR", "STUFF=THINGS"))
}

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
		args := f.ToSlice(testStep.GetDashes())
		assert.Equal(t, []string{"-f", "abc"}, args)
	})

	t.Run("long flag", func(t *testing.T) {
		f := NewFlag("full", "abc")
		args := f.ToSlice(testStep.GetDashes())
		assert.Equal(t, []string{"--full", "abc"}, args)
	})

	t.Run("valueless flag", func(t *testing.T) {
		f := NewFlag("l")
		args := f.ToSlice(testStep.GetDashes())
		assert.Equal(t, []string{"-l"}, args)
	})

	dashes := Dashes{
		Long:  "---",
		Short: "-",
	}
	t.Run("flag with non-default dashes", func(t *testing.T) {
		f := NewFlag("full", "abc")
		args := f.ToSlice(dashes)
		assert.Equal(t, []string{"---full", "abc"}, args)
	})
}

func TestFlags_ToSlice(t *testing.T) {
	flags := Flags{
		NewFlag("bull", "2"),
		NewFlag("a", "1"),
	}

	args := flags.ToSlice(testStep.GetDashes())

	// Flags should be sorted and sliced up on a platter
	assert.Equal(t, []string{"-a", "1", "--bull", "2"}, args)
}

func TestFlags_NonStringKeys(t *testing.T) {
	flags := Flags{}
	err := yaml.Unmarshal([]byte(`
yes: ["y", "true"]
true: ["y", "yes"]
nil: ["no", "nah"]
1.0: ["true", "yes"]
1: ["yes"]
`), &flags)
	assert.NoError(t, err)
	assert.Equal(t, 5, len(flags))
}
