package porter

import (
	"fmt"
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
			"invalid value for --source",
		},
		{
			"bad source, invalid digest ref",
			CopyOpts{
				Source:      "deislabs/mybuns@v0.1.0",
				Destination: "blah.acr.io",
			},
			true,
			"invalid value for --source",
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
			require.Error(t, err)
			assert.Contains(t, err.Error(), test.ExpectedError)
		} else {
			assert.NoError(t, err)
		}
	}
}

func TestCopyGenerateBundleRef(t *testing.T) {
	tests := []struct {
		Name     string
		Opts     CopyOpts
		Expected string
		WantErr  string
	}{
		{
			Name: "tag source and dest repo",
			Opts: CopyOpts{
				Source:      "deislabs/mybuns:v0.1.0",
				Destination: "blah.acr.io",
			},
			Expected: "blah.acr.io/mybuns:v0.1.0",
		},
		{
			Name: "tag source and dest tag",
			Opts: CopyOpts{
				Source:      "deislabs/mybuns:v0.1.0",
				Destination: "blah.acr.io/blah:v0.10",
			},
			Expected: "blah.acr.io/blah:v0.10",
		},
		{
			Name: "valid source digest and tagged destination",
			Opts: CopyOpts{
				Source:      "deislabs/mybuns@sha256:bb9b47bb07e8c2f62ea1f617351739b35264f8a6121d79e989cd4e81743afe0a",
				Destination: "blah.acr.io/moreblah:v0.1.0",
			},
			Expected: "blah.acr.io/moreblah:v0.1.0",
		},
		{
			Name: "invalid destination",
			Opts: CopyOpts{
				Source:      "deislabs/mybuns:v0.1.0",
				Destination: "oops/",
			},
			WantErr: "invalid reference format oops",
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			src := cnab.MustParseOCIReference(test.Opts.Source)
			newRef, err := generateNewBundleRef(src, test.Opts.Destination)
			if test.WantErr == "" {
				assert.Equal(t, test.Expected, newRef.String(), fmt.Sprintf("%s: expected %s got %s", test.Name, test.Expected, newRef))
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.WantErr)
			}
		})
	}
}
