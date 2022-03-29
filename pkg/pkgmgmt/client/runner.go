package client

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"get.porter.sh/porter/pkg/pkgmgmt"
	"get.porter.sh/porter/pkg/portercontext"
	"github.com/pkg/errors"
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
		return errors.Wrapf(err, "failed to stat package (%s)", pkgPath)
	}
	if !exists {
		return errors.Errorf("package not found (%s)", pkgPath)
	}

	return nil
}

func (r *Runner) Run(commandOpts pkgmgmt.CommandOptions) error {
	if r.Debug {
		fmt.Fprintf(r.Err, "DEBUG name:    %s\n", r.pkgName)
		fmt.Fprintf(r.Err, "DEBUG pkgDir: %s\n", r.pkgDir)
		fmt.Fprintf(r.Err, "DEBUG file:     %s\n", commandOpts.File)
		fmt.Fprintf(r.Err, "DEBUG stdin:\n%s\n", commandOpts.Input)
	}

	pkgPath := r.getExecutablePath()
	cmdArgs := strings.Split(commandOpts.Command, " ")
	command := cmdArgs[0]
	cmd := r.NewCommand(pkgPath, cmdArgs...)

	// Pipe the output to porter
	cmd.Stdout = r.Context.Out
	cmd.Stderr = r.Context.Err

	if commandOpts.PreRun != nil {
		commandOpts.PreRun(command, cmd)
	}

	if commandOpts.File != "" {
		cmd.Args = append(cmd.Args, "-f", commandOpts.File)
	}

	if r.Debug {
		cmd.Args = append(cmd.Args, "--debug")
	}

	if commandOpts.Input != "" {
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return err
		}
		go func() {
			defer stdin.Close()
			io.WriteString(stdin, commandOpts.Input)
		}()
	}

	prettyCmd := fmt.Sprintf("%s%s", cmd.Dir, strings.Join(cmd.Args, " "))
	if r.Debug {
		fmt.Fprintln(r.Err, prettyCmd)
	}

	err := cmd.Start()
	if err != nil {
		return errors.Wrapf(err, "could not run package command %s", prettyCmd)
	}

	return cmd.Wait()
}

func (r *Runner) getExecutablePath() string {
	path := filepath.Join(r.pkgDir, r.pkgName)
	if r.runtime {
		return filepath.Join(r.pkgDir, "runtimes", r.pkgName+"-runtime")
	}
	return path + pkgmgmt.FileExt
}
