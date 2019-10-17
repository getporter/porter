package porter

import (
	"fmt"
	"testing"

	"github.com/docker/distribution/reference"
	"github.com/stretchr/testify/assert"
)

func makeNamed(ref string) reference.Named {
	n, _ := reference.ParseNormalizedNamed(ref)
	return n
}
func TestCopyCheckDigestedTest(t *testing.T) {
	tests := []struct {
		Name     string
		Ref      reference.Named
		Expected bool
	}{
		{
			"valid digested reference",
			makeNamed("jeremyrickard/porter-do-bundle@sha256:a808aa4e3508d7129742eefda938249574447cce5403dc12d4cbbfe7f4f31e58"),
			true,
		},
		{
			"tagged reference",
			makeNamed("jeremyrickard/porter-do-bundle:v0.1.0"),
			false,
		},
		{
			"no tag",
			makeNamed("jeremyrickard/porter-do-bundle"),
			false,
		},
	}

	for _, test := range tests {
		ref := isCopyDigestReference(test.Ref)
		assert.Equal(t, test.Expected, ref, fmt.Sprintf("%s, expected %t, got %t", test.Name, test.Expected, ref))
	}
}

func TestCopyReferenceOnly(t *testing.T) {
	tests := []struct {
		Name     string
		Ref      string
		Expected bool
	}{
		{
			"valid digested reference",
			"jeremyrickard/porter-do-bundle@sha256:a808aa4e3508d7129742eefda938249574447cce5403dc12d4cbbfe7f4f31e58",
			false,
		},
		{
			"tagged reference",
			"jeremyrickard/porter-do-bundle:v0.1.0",
			false,
		},
		{
			"no tag or digest",
			"porter-do-bundle",
			true,
		},
		{
			"no tag or digest",
			"jeremy/rickard/porter-do-bundle",
			true,
		},
	}

	for _, test := range tests {
		ref := isCopyReferenceOnly(test.Ref)
		assert.Equal(t, test.Expected, ref, fmt.Sprintf("%s, expected %t, got %t", test.Name, test.Expected, ref))
	}
}

func TestValidateCopyArgs(t *testing.T) {

	tests := []struct {
		Name          string
		Opts          CopyOpts
		ExpectError   bool
		ExpectedError string
	}{
		{
			"valid source and dest",
			CopyOpts{
				Source:      "deislabs/mybuns:v0.1.0",
				Destination: "blah.acr.io",
			},
			false,
			"",
		},
		{
			"valid source digest and tagged destination",
			CopyOpts{
				Source:      "deislabs/mybuns@sha256:bb9b47bb07e8c2f62ea1f617351739b35264f8a6121d79e989cd4e81743afe0a",
				Destination: "blah.acr.io:v0.1.0",
			},
			false,
			"",
		},
		{
			"valid source, empty dest",
			CopyOpts{
				Source: "deislabs/mybuns:v0.1.0",
			},
			true,
			"--destination is required",
		},
		{
			"missing source",
			CopyOpts{
				Source:      "",
				Destination: "blah.acr.io",
			},
			true,
			"invalid value for --source, specified value should be of the form REGISTRY/bundle:tag or REGISTRY/bundle@sha: invalid reference format",
		},
		{
			"bad source, invalid digest ref",
			CopyOpts{
				Source:      "deislabs/mybuns@v0.1.0",
				Destination: "blah.acr.io",
			},
			true,
			"invalid value for --source, specified value should be of the form REGISTRY/bundle:tag or REGISTRY/bundle@sha: invalid reference format",
		},
		{
			"digest to reference only should fail",
			CopyOpts{
				Source:      "jeremyrickard/porter-do@sha256:bb9b47bb07e8c2f62ea1f617351739b35264f8a6121d79e989cd4e81743afe0a",
				Destination: "blah.acr.io",
			},
			true,
			"--destination must be tagged reference when --source is digested reference",
		},
	}

	for _, test := range tests {
		err := test.Opts.Validate()
		if test.ExpectError {
			assert.Error(t, err)
			assert.EqualError(t, err, test.ExpectedError)
		} else {
			assert.NoError(t, err)
		}
	}
}
