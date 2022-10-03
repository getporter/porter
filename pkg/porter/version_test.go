package porter

import (
	"context"
	"runtime"
	"strings"
	"testing"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/printer"
	"get.porter.sh/porter/pkg/test"
	"github.com/stretchr/testify/require"
)

func TestPrintVersion(t *testing.T) {
	pkg.Commit = "abc123"
	pkg.Version = "v1.2.3"
	defer func() {
		pkg.Commit = ""
		pkg.Version = ""
	}()

	ctx := context.Background()
	p := NewTestPorter(t)
	defer p.Close()

	opts := VersionOpts{}
	err := opts.Validate()
	require.NoError(t, err)
	p.PrintVersion(ctx, opts)

	gotOutput := p.TestConfig.TestContext.GetOutput()
	wantOutput := "porter v1.2.3 (abc123)"
	if !strings.Contains(gotOutput, wantOutput) {
		t.Fatalf("invalid output:\nWANT:\t%q\nGOT:\t%q\n", wantOutput, gotOutput)
	}
}

func TestPrintJsonVersion(t *testing.T) {
	pkg.Commit = "abc123"
	pkg.Version = "v1.2.3"
	defer func() {
		pkg.Commit = ""
		pkg.Version = ""
	}()

	ctx := context.Background()
	p := NewTestPorter(t)
	defer p.Close()

	opts := VersionOpts{}
	opts.RawFormat = string(printer.FormatJson)
	err := opts.Validate()
	require.NoError(t, err)
	p.PrintVersion(ctx, opts)

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
	// Only run this on linux + amd64 machines to simplify the test (it has different output based on the os/arch)
	if runtime.GOOS != "linux" || runtime.GOARCH != "amd64" {
		t.Skip("skipping test because it is only for linux/amd64")
	}

	pkg.Commit = "abc123"
	pkg.Version = "v1.2.3"
	defer func() {
		pkg.Commit = ""
		pkg.Version = ""
	}()

	ctx := context.Background()
	p := NewTestPorter(t)
	defer p.Close()

	opts := VersionOpts{System: true}
	opts.RawFormat = string(printer.FormatJson)
	err := opts.Validate()
	require.Nil(t, err)
	p.PrintVersion(ctx, opts)

	gotOutput := p.TestConfig.TestContext.GetOutput()
	test.CompareGoldenFile(t, "testdata/version/version-output.json", gotOutput)
}

func TestPrintDebugInfoPlainTextVersion(t *testing.T) {
	// Only run this on linux + amd64 machines to simplify the test (it has different output based on the os/arch)
	if runtime.GOOS != "linux" || runtime.GOARCH != "amd64" {
		t.Skip("skipping test because it is only for linux/amd64")
	}

	pkg.Commit = "abc123"
	pkg.Version = "v1.2.3"
	defer func() {
		pkg.Commit = ""
		pkg.Version = ""
	}()

	ctx := context.Background()
	p := NewTestPorter(t)
	defer p.Close()

	opts := VersionOpts{System: true}
	err := opts.Validate()
	require.Nil(t, err)
	p.PrintVersion(ctx, opts)

	gotOutput := p.TestConfig.TestContext.GetOutput()
	test.CompareGoldenFile(t, "testdata/version/version-output.txt", gotOutput)
}
