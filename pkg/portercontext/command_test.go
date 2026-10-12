package portercontext

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatCommand(t *testing.T) {
	testcases := []struct {
		name string
		goos string
		args []string
		want string
	}{
		{name: "linux", goos: "linux",
			args: []string{"/home/me/.porter/mixins/exec/exec", "version", "--output", "json"},
			want: "/home/me/.porter/mixins/exec/exec version --output json"},
		{name: "linux with spaces", goos: "linux",
			args: []string{"/home/my user/.porter/mixins/exec/exec", "install", "-f", "my file.yaml"},
			want: "'/home/my user/.porter/mixins/exec/exec' install -f 'my file.yaml'"},
		{name: "darwin with spaces", goos: "darwin",
			args: []string{"/Users/me/.porter/mixins/exec/exec", "install", "-f", "my file.yaml"},
			want: "/Users/me/.porter/mixins/exec/exec install -f 'my file.yaml'"},
		{name: "windows", goos: "windows",
			args: []string{`C:\Users\me\.porter\mixins\exec\exec.exe`, "version", "--output", "json"},
			want: `C:\Users\me\.porter\mixins\exec\exec.exe version --output json`},
		{name: "windows with spaces", goos: "windows",
			args: []string{`C:\Users\my user\.porter\mixins\exec\exec.exe`, "install", "-f", "my file.yaml"},
			want: `"C:\Users\my user\.porter\mixins\exec\exec.exe" install -f "my file.yaml"`},
		{name: "windows with quotes", goos: "windows",
			args: []string{"exec.exe", `say "hi"`},
			want: `exec.exe "say \"hi\""`},
		{name: "windows with trailing backslash", goos: "windows",
			args: []string{"exec.exe", `C:\my dir\`},
			want: `exec.exe "C:\my dir\\"`},
		{name: "windows with cmd operators", goos: "windows",
			args: []string{"exec.exe", "-f", "a&b.yaml", "x|y", "<in>", "a^b", "(c)"},
			want: `exec.exe -f "a&b.yaml" "x|y" "<in>" "a^b" "(c)"`},
		{name: "windows empty argument", goos: "windows",
			args: []string{"exec.exe", ""},
			want: `exec.exe ""`},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			got := formatCommand(tc.goos, tc.args)
			assert.Equal(t, tc.want, got)
		})
	}
}
