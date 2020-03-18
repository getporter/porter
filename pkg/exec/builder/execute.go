package builder

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"get.porter.sh/porter/pkg/context"
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
	GetArguments() []string
	GetFlags() Flags
}

type HasCustomDashes interface {
	GetDashes() Dashes
}

// ExecuteSingleStepAction runs the command represented by an ExecutableAction, where only
// a single step is allowed to be defined in the Action (which is what happens when Porter
// executes steps one at a time).
func ExecuteSingleStepAction(cxt *context.Context, action ExecutableAction) (string, error) {
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
func ExecuteStep(cxt *context.Context, step ExecutableStep) (string, error) {
	arguments := step.GetArguments()
	flags := step.GetFlags()
	args := make([]string, len(arguments), 1+len(arguments)+len(flags)*2)

	copy(args, arguments)
	dashes := DefaultFlagDashes
	if dashing, ok := step.(HasCustomDashes); ok {
		dashes = dashing.GetDashes()
	}
	args = append(args, flags.ToSlice(dashes)...)

	// Split up any arguments or flags that have spaces so that we pass them as separate array elements
	// It doesn't show up any differently in the printed command, but it matters to how the command
	// it executed against the system.
	args = splitCommand(args)

	cmd := cxt.NewCommand(step.GetCommand(), args...)
	output := &bytes.Buffer{}
	cmd.Stdout = io.MultiWriter(cxt.Out, output)
	cmd.Stderr = cxt.Err

	prettyCmd := fmt.Sprintf("%s%s", cmd.Dir, strings.Join(cmd.Args, " "))
	if cxt.Debug {
		fmt.Fprintln(cxt.Out, prettyCmd)
	}

	err := cmd.Start()
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("couldn't run command %s", prettyCmd))
	}

	err = cmd.Wait()
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("error running command %s", prettyCmd))
	}

	return output.String(), nil
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
