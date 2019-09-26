package builder

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/deislabs/porter/pkg/context"
	"github.com/pkg/errors"
)

type ExecutableAction interface {
	GetSteps() []ExecutableStep
}

type ExecutableStep interface {
	GetCommand() string
	GetArguments() []string
	GetFlags() Flags
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
	args = append(args, flags.ToSlice()...)

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
