package builder

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"get.porter.sh/porter/pkg/portercontext"
	"github.com/pkg/errors"
)

var DefaultFlagDashes = Dashes{
	Long:  "--",
	Short: "-",
}

// BuildableAction is an Action that can be marshaled and unmarshaled "generically"
type BuildableAction interface {
	// MakeSteps returns a Steps struct to unmarshal into.
	MakeSteps() interface{}
}

type ExecutableAction interface {
	GetSteps() []ExecutableStep
}

type ExecutableStep interface {
	GetCommand() string
	//GetArguments() puts the arguments at the beginning of the command
	GetArguments() []string
	GetFlags() Flags
	GetWorkingDir() string
}

type HasEnvironmentVars interface {
	GetEnvironmentVars() map[string]string
}

type HasOrderedArguments interface {
	GetSuffixArguments() []string
}

type HasCustomDashes interface {
	GetDashes() Dashes
}

type SuppressesOutput interface {
	SuppressesOutput() bool
}

// HasErrorHandling is implemented by mixin commands that want to handle errors
// themselves, and possibly allow failed commands to either pass, or to improve
// the displayed error message
type HasErrorHandling interface {
	HandleError(cxt *portercontext.Context, err ExitError, stdout string, stderr string) error
}

type ExitError interface {
	error
	ExitCode() int
}

// ExecuteSingleStepAction runs the command represented by an ExecutableAction, where only
// a single step is allowed to be defined in the Action (which is what happens when Porter
// executes steps one at a time).
func ExecuteSingleStepAction(cxt *portercontext.Context, action ExecutableAction) (string, error) {
	steps := action.GetSteps()
	if len(steps) != 1 {
		return "", errors.Errorf("expected a single step, but got %d", len(steps))
	}
	step := steps[0]

	output, err := ExecuteStep(cxt, step)
	if err != nil {
		return output, err
	}

	swo, ok := step.(StepWithOutputs)
	if !ok {
		return output, nil
	}

	err = ProcessJsonPathOutputs(cxt, swo, output)
	if err != nil {
		return output, err
	}

	err = ProcessRegexOutputs(cxt, swo, output)
	if err != nil {
		return output, err
	}

	err = ProcessFileOutputs(cxt, swo)
	return output, err
}

// ExecuteStep runs the command represented by an ExecutableStep, piping stdout/stderr
// back to the context and returns the buffered output for subsequent processing.
func ExecuteStep(pctx *portercontext.Context, step ExecutableStep) (string, error) {
	ctx := context.TODO()

	// Identify if any suffix arguments are defined
	var suffixArgs []string
	orderedArgs, ok := step.(HasOrderedArguments)
	if ok {
		suffixArgs = orderedArgs.GetSuffixArguments()
	}

	// Preallocate an array big enough to hold all arguments
	arguments := step.GetArguments()
	flags := step.GetFlags()
	args := make([]string, len(arguments), 1+len(arguments)+len(flags)*2+len(suffixArgs))

	// Copy all prefix arguments
	copy(args, arguments)

	// Copy all flags
	dashes := DefaultFlagDashes
	if dashing, ok := step.(HasCustomDashes); ok {
		dashes = dashing.GetDashes()
	}

	// Split up flags that have spaces so that we pass them as separate array elements
	// It doesn't show up any differently in the printed command, but it matters to how the command
	// it executed against the system.
	flagsSlice := splitCommand(flags.ToSlice(dashes))

	args = append(args, flagsSlice...)

	// Append any final suffix arguments
	args = append(args, suffixArgs...)

	// Add env vars if defined
	if stepWithEnvVars, ok := step.(HasEnvironmentVars); ok {
		for k, v := range stepWithEnvVars.GetEnvironmentVars() {
			pctx.Setenv(k, v)
		}
	}

	cmd := pctx.NewCommand(ctx, step.GetCommand(), args...)

	// ensure command is executed in the correct directory
	wd := step.GetWorkingDir()
	if len(wd) > 0 && wd != "." {
		cmd.Dir = wd
	}

	prettyCmd := fmt.Sprintf("%s %s", cmd.Dir, strings.Join(cmd.Args, " "))

	// Setup output streams for command
	// If Step suppresses output, update streams accordingly
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	suppressOutput := false
	if suppressable, ok := step.(SuppressesOutput); ok {
		suppressOutput = suppressable.SuppressesOutput()
	}

	if suppressOutput {
		// We still capture the output, but we won't print it
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		if pctx.Debug {
			fmt.Fprintf(pctx.Err, "DEBUG: output suppressed for command %s\n", prettyCmd)
		}
	} else {
		cmd.Stdout = io.MultiWriter(pctx.Out, stdout)
		cmd.Stderr = io.MultiWriter(pctx.Err, stderr)
		if pctx.Debug {
			fmt.Fprintln(pctx.Err, prettyCmd)
		}
	}

	err := cmd.Start()
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("couldn't run command %s", prettyCmd))
	}

	err = cmd.Wait()

	// Check if the command knows how to handle and recover from its own errors
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if handler, ok := step.(HasErrorHandling); ok {
				err = handler.HandleError(pctx, exitErr, stdout.String(), stderr.String())
			}
		}
	}

	// Ok, now check if we still have a problem
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("error running command %s", prettyCmd))
	}

	return stdout.String(), nil
}

var whitespace = string([]rune{space, newline, tab})

const (
	space       = rune(' ')
	newline     = rune('\n')
	tab         = rune('\t')
	backslash   = rune('\\')
	doubleQuote = rune('"')
	singleQuote = rune('\'')
)

// expandOnWhitespace finds elements with multiple words that are not "glued" together with quotes
// and splits them into separate elements in the slice
func splitCommand(slice []string) []string {
	expandedSlice := make([]string, 0, len(slice))
	for _, chunk := range slice {
		chunkettes := findWords(chunk)
		expandedSlice = append(expandedSlice, chunkettes...)
	}

	return expandedSlice
}

func findWords(input string) []string {
	words := make([]string, 0, 1)
	next := input
	for len(next) > 0 {
		word, remainder, err := findNextWord(next)
		if err != nil {
			return []string{input}
		}
		next = remainder
		words = append(words, word)
	}

	return words
}

func findNextWord(input string) (string, string, error) {
	var buf bytes.Buffer

	// Remove leading whitespace before starting
	input = strings.TrimLeft(input, whitespace)

	var escaped bool
	var wordStart, wordStop int
	var closingQuote rune

	for i, r := range input {
		// Prevent escaped characters from matching below
		if escaped {
			r = -1
			escaped = false
		}

		switch r {
		case backslash:
			// Escape the next character
			escaped = true
			continue
		case closingQuote:
			wordStop = i
			closingQuote = 0 // Reset looking for a closing quote
		case singleQuote, doubleQuote:
			// Seek to the closing quote only
			if closingQuote != 0 {
				continue
			}

			wordStart = 1    // Skip opening quote
			closingQuote = r // Seek to the same closing quote
		case space, tab, newline:
			// Seek to the closing quote only
			if closingQuote != 0 {
				continue
			}

			wordStart = 0
			wordStop = i
		}

		// Found the end of a word
		if wordStop > 0 {
			_, err := buf.WriteString(input[wordStart:wordStop])
			if err != nil {
				return "", input, errors.New("error writing to buffer")
			}
			return buf.String(), input[wordStop+1:], nil
		}
	}

	if closingQuote != 0 {
		return "", "", errors.New("unmatched quote found")
	}

	// Hit the end of input, flush the remainder
	_, err := buf.WriteString(input)
	if err != nil {
		return "", input, errors.New("error writing to buffer")
	}

	return buf.String(), "", nil
}
