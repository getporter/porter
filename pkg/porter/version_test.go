package porter

import (
	"fmt"
	"runtime"
	"strings"
	"testing"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/printer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrintVersion(t *testing.T) {
	pkg.Commit = "abc123"
	pkg.Version = "v1.2.3"

	p := NewTestPorter(t)

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

	opts := VersionOpts{System: true}
	p.TestConfig.SetupPorterHome()
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
      "author": "Deis Labs"
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

	opts := VersionOpts{System: true}
	p.TestConfig.SetupPorterHome()
	err := opts.Validate()
	require.Nil(t, err)
	p.PrintVersion(opts)

	versionOutput := "porter v1.2.3 (abc123)"
	mixinsOutput := "exec   v1.0      Deis Labs"
	systemOutput := fmt.Sprintf("os: %s\narch: %s", runtime.GOOS, runtime.GOARCH)

	gotOutput := p.TestConfig.TestContext.GetOutput()
	assert.Contains(t, gotOutput, versionOutput)
	assert.Contains(t, gotOutput, mixinsOutput)
	assert.Contains(t, gotOutput, systemOutput)
}
