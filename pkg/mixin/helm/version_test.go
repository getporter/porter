package helm

import (
	"bytes"
	"strings"
	"testing"

	"github.com/deislabs/porter/pkg"
)

func TestPrintVersion(t *testing.T) {
	pkg.Commit = "abc123"
	pkg.Version = "v1.2.3"

	output := &bytes.Buffer{}
	m := Mixin{Out: output}
	m.PrintVersion()

	gotOutput := string(output.Bytes())
	wantOutput := "helm mixin v1.2.3 (abc123)"
	if !strings.Contains(gotOutput, wantOutput) {
		t.Fatalf("invalid output:\nWANT:\t%q\nGOT:\t%q\n", wantOutput, gotOutput)
	}
}
