package porter

import (
	"fmt"
	"runtime"
	"strings"
	"testing"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/printer"
	"get.porter.sh/porter/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrintVersion(t *testing.T) {
	pkg.Commit = "abc123"
	pkg.Version = "v1.2.3"

	p := NewTestPorter(t)
	defer p.Teardown()

	opts := VersionOpts{}
	err := opts.Validate()
	require.NoError(t, err)
	p.PrintVersion(opts)

	gotOutput := p.TestConfig.TestContext.GetOutput()
	wantOutput := "porter v1.2.3 (abc123)"
	if !strings.Contains(gotOutput, wantOutput) {
		t.Fatalf("invalid output:\nWANT:\t%q\nGOT:\t%q\n", wantOutput, gotOutput)
	}
}

func TestPrintJsonVersion(t *testing.T) {
	pkg.Commit = "abc123"
	pkg.Version = "v1.2.3"

	p := NewTestPorter(t)
	defer p.Teardown()

	opts := VersionOpts{}
	opts.RawFormat = string(printer.FormatJson)
	err := opts.Validate()
	require.NoError(t, err)
	p.PrintVersion(opts)

	gotOutput := p.TestConfig.TestContext.GetOutput()
	wantOutput := `{
  "name": "porter",
  "version": "v1.2.3",
  "commit": "abc123"
}
`
	if !strings.Contains(gotOutput, wantOutput) {
		t.Fatalf("invalid output:\nWANT:\t%q\nGOT:\t%q\n", wantOutput, gotOutput)
	}
}

func TestPrintDebugInfoJsonVersion(t *testing.T) {
	pkg.Commit = "abc123"
	pkg.Version = "v1.2.3"

	p := NewTestPorter(t)
	defer p.Teardown()

	opts := VersionOpts{System: true}
	opts.RawFormat = string(printer.FormatJson)
	err := opts.Validate()
	require.Nil(t, err)
	p.PrintVersion(opts)

	gotOutput := p.TestConfig.TestContext.GetOutput()
	wantOutput := fmt.Sprintf(`{
  "version": {
    "name": "porter",
    "version": "v1.2.3",
    "commit": "abc123"
  },
  "system": {
    "OS": "%s",
    "Arch": "%s"
  },
  "mixins": [
    {
      "name": "exec",
      "version": "v1.0",
      "commit": "abc123",
      "author": "Porter Authors"
    }
  ]
}
`, runtime.GOOS, runtime.GOARCH)
	assert.Equal(t, wantOutput, gotOutput)
}

func TestPrintDebugInfoPlainTextVersion(t *testing.T) {
	pkg.Commit = "abc123"
	pkg.Version = "v1.2.3"

	p := NewTestPorter(t)
	defer p.Teardown()

	opts := VersionOpts{System: true}
	err := opts.Validate()
	require.Nil(t, err)
	p.PrintVersion(opts)

	gotOutput := p.TestConfig.TestContext.GetOutput()
	test.CompareGoldenFile(t, "testdata/version/version-output.txt", gotOutput)
}
