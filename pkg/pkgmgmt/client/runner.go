package client

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"get.porter.sh/porter/pkg/pkgmgmt"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/tracing"
	"go.opentelemetry.io/otel/attribute"
)

type Runner struct {
	*portercontext.Context
	// pkgDir is the absolute path to where the package is installed
	pkgDir string

	pkgName string
	runtime bool
}

func NewRunner(pkgName, pkgDir string, runtime bool) *Runner {
	return &Runner{
		Context: portercontext.New(),
		pkgName: pkgName,
		pkgDir:  pkgDir,
		runtime: runtime,
	}
}

func (r *Runner) Validate() error {
	if r.pkgName == "" {
		return errors.New("package name to execute not specified")
	}

	pkgPath := r.getExecutablePath()
	exists, err := r.FileSystem.Exists(pkgPath)
	if err != nil {
		return fmt.Errorf("failed to stat package (%s: %w)", pkgPath, err)
	}
	if !exists {
		return fmt.Errorf("package not found (%s)", pkgPath)
	}

	return nil
}

func (r *Runner) Run(ctx context.Context, commandOpts pkgmgmt.CommandOptions) error {
	ctx, span := tracing.StartSpan(ctx,
		attribute.String("name", r.pkgName),
		attribute.String("pkgDir", r.pkgDir),
		attribute.String("file", commandOpts.File),
		attribute.String("stdin", commandOpts.Input),
	)
	defer span.EndSpan()

	pkgPath := r.getExecutablePath()
	cmdArgs := strings.Split(commandOpts.Command, " ")
	command := cmdArgs[0]
	cmd := r.NewCommand(ctx, pkgPath, cmdArgs...)

	// Pipe the output to porter and capture the error in case it fails
	cmdStderr := &bytes.Buffer{}
	cmd.Stdout = r.Context.Out
	cmd.Stderr = io.MultiWriter(cmdStderr, r.Context.Err)

	if commandOpts.PreRun != nil {
		commandOpts.PreRun(command, cmd)
	}

	if commandOpts.File != "" {
		cmd.Args = append(cmd.Args, "-f", commandOpts.File)
	}

	if commandOpts.Input != "" {
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return span.Error(err)
		}
		go func() {
			defer stdin.Close()
			io.WriteString(stdin, commandOpts.Input)
		}()
	}

	prettyCmd := fmt.Sprintf("%s%s", cmd.Dir, strings.Join(cmd.Args, " "))
	span.SetAttributes(attribute.String("command", prettyCmd))

	err := cmd.Start()
	if err != nil {
		return span.Error(fmt.Errorf("could not start package command %s: %w", prettyCmd, err))
	}

	err = cmd.Wait()
	if err != nil {
		// Include stderr in the error, otherwise it just includes the exit code
		err = fmt.Errorf("package command failed %s\n%s", prettyCmd, cmdStderr)
		// Do not flag this as an error in the logs because we often call mixins to see if they support a command
		// and if they don't it's not an error, e.g. not all mixins support lint or schema
		span.Debugf(err.Error())
		return err
	}

	return nil
}

func (r *Runner) getExecutablePath() string {
	path := filepath.Join(r.pkgDir, r.pkgName)
	if r.runtime {
		return filepath.Join(r.pkgDir, "runtimes", r.pkgName+"-runtime")
	}
	return path + pkgmgmt.FileExt
}
