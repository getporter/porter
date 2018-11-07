package porter

import (
	"strings"
	"testing"

	"github.com/deislabs/porter/pkg"
)

func TestPrintVersion(t *testing.T) {
	pkg.Commit = "abc123"
	pkg.Version = "v1.2.3"

	p, output := NewTestPorter(t)

	p.PrintVersion()

	gotOutput := string(output.Bytes())
	wantOutput := "porter v1.2.3 (abc123)"
	if !strings.Contains(gotOutput, wantOutput) {
		t.Fatalf("invalid output:\nWANT:\t%q\nGOT:\t%q\n", wantOutput, gotOutput)
	}
}
