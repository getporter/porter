package builder

import (
	"fmt"
	"regexp"
	"strings"

	portercontext "get.porter.sh/porter/pkg/portercontext"
)

var _ HasErrorHandling = IgnoreErrorHandler{}

// IgnoreErrorHandler implements HasErrorHandling for the exec mixin
// and can be used by any other mixin to get the same error handling behavior.
type IgnoreErrorHandler struct {
	// All ignores any error that happens when the command is run.
	All bool `yaml:"all,omitempty"`

	// ExitCodes ignores any exit codes in the list.
	ExitCodes []int `yaml:"exitCodes,omitempty"`

	// Output determines if the error should be ignored based on the command
	// output.
	Output IgnoreErrorWithOutput `yaml:"output,omitempty"`
}

type IgnoreErrorWithOutput struct {
	// Contains specifies that the error is ignored when stderr contains the
	// specified substring.
	Contains []string `yaml:"contains,omitempty"`

	// Regex specifies that the error is ignored when stderr matches the
	// specified regular expression.
	Regex []string `yaml:"regex,omitempty"`
}

func (h IgnoreErrorHandler) HandleError(cxt *portercontext.Context, err ExitError, stdout string, stderr string) error {
	// We shouldn't be called when there is no error but just in case, let's check
	if err == nil || err.ExitCode() == 0 {
		return nil
	}

	if cxt.Debug {
		fmt.Fprintf(cxt.Err, "Evaluating mixin command error %s with the mixin's error handler\n", err.Error())
	}

	// Check if the command should always be allowed to "pass"
	if h.All {
		if cxt.Debug {
			fmt.Fprintln(cxt.Err, "Ignoring mixin command error because All was specified in the mixin step definition")
		}
		return nil
	}

	// Check if the exit code was allowed
	exitCode := err.ExitCode()
	for _, code := range h.ExitCodes {
		if exitCode == code {
			if cxt.Debug {
				fmt.Fprintf(cxt.Err, "Ignoring mixin command error (exit code: %d) because it was included in the allowed ExitCodes list defined in the mixin step definition\n", exitCode)
			}
			return nil
		}
	}

	// Check if the output contains a hint that it should be allowed to pass
	for _, allowError := range h.Output.Contains {
		if strings.Contains(stderr, allowError) {
			if cxt.Debug {
				fmt.Fprintf(cxt.Err, "Ignoring mixin command error because the error contained the substring %q defined in the mixin step definition\n", allowError)
			}
			return nil
		}
	}

	// Check if the output matches an allowed regular expression
	for _, allowMatch := range h.Output.Regex {
		expression, regexErr := regexp.Compile(allowMatch)
		if regexErr != nil {
			fmt.Fprintf(cxt.Err, "Could not ignore failed command because the Regex specified by the mixin step definition (%q) is invalid:%s\n", allowMatch, regexErr.Error())
			return err
		}

		if expression.MatchString(stderr) {
			if cxt.Debug {
				fmt.Fprintf(cxt.Err, "Ignoring mixin command error because the error matched the Regex %q defined in the mixin step definition\n", allowMatch)
			}
			return nil
		}
	}

	return err
}
