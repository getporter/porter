package exec

import (
	"bytes"
	"strings"
	"testing"

	"github.com/deislabs/porter/pkg"
)

func TestPrintVersion(t *testing.T) {
	output := &bytes.Buffer{}
	pkg.Commit = "abc123"
	pkg.Version = "v1.2.3"

	p := &Porter{Out: output}

	p.PrintVersion()

	gotOutput := string(output.Bytes())
	wantOutput := "exec v1.2.3 (abc123)"
	if !strings.Contains(gotOutput, wantOutput) {
		t.Fatalf("invalid output:\nWANT:\t%q\nGOT:\t%q\n", wantOutput, gotOutput)
	}
}
