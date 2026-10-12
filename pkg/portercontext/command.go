package portercontext

import (
	"os/exec"
	"runtime"
	"strings"

	"github.com/kballard/go-shellquote"
)

// FormatCommand returns the command and its arguments as a single string that
// can be copied and run as-is, quoted for the current operating system:
// for /bin/sh, or for cmd.exe on Windows.
func FormatCommand(cmd *exec.Cmd) string {
	return formatCommand(runtime.GOOS, cmd.Args)
}

func formatCommand(goos string, args []string) string {
	if goos != "windows" {
		return shellquote.Join(args...)
	}

	quoted := make([]string, len(args))
	for i, arg := range args {
		quoted[i] = quoteWindowsArg(arg)
	}
	return strings.Join(quoted, " ")
}

// windowsSpecialChars are the characters that require an argument to be quoted
// on Windows: those that separate or quote arguments, and those that cmd.exe
// treats as operators when they are not quoted.
const windowsSpecialChars = " \t\"&|<>^()"

// quoteWindowsArg quotes an argument following the rules that Windows programs
// use to parse their command line, which are the same as syscall.EscapeArg.
// Backslashes are only special when they are followed by a double quote.
//
// Arguments with cmd.exe operators are quoted as well, so that cmd.exe passes
// them on as-is. Environment variable references, such as %PATH%, cannot be
// escaped on the cmd.exe command line and are still expanded.
func quoteWindowsArg(arg string) string {
	if arg == "" {
		return `""`
	}
	if !strings.ContainsAny(arg, windowsSpecialChars) {
		return arg
	}

	var b strings.Builder
	b.WriteByte('"')
	backslashes := 0
	for i := 0; i < len(arg); i++ {
		switch arg[i] {
		case '\\':
			backslashes++
		case '"':
			// Double the backslashes before the quote, and escape the quote
			b.WriteString(strings.Repeat(`\`, backslashes+1))
			backslashes = 0
		default:
			backslashes = 0
		}
		b.WriteByte(arg[i])
	}
	// Double the trailing backslashes so that they do not escape the closing quote
	b.WriteString(strings.Repeat(`\`, backslashes))
	b.WriteByte('"')
	return b.String()
}
