package mixin

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/deislabs/porter/pkg/context"
	"github.com/pkg/errors"
)

type Runner struct {
	*context.Context
	// mixinDir is the absolute path to the directory containing the mixin
	mixinDir string

	Mixin   string
	Runtime bool
	Command string
	Input   string
	File    string
}

func NewRunner(mixin, mixinDir string, runtime bool) *Runner {
	return &Runner{
		Context:  context.New(),
		Mixin:    mixin,
		Runtime:  runtime,
		mixinDir: mixinDir,
	}
}

func (r *Runner) Validate() error {
	if r.Mixin == "" {
		return errors.New("mixin not specified")
	}

	mixinPath := r.getMixinPath()
	exists, err := r.FileSystem.Exists(mixinPath)
	if err != nil {
		return errors.Wrapf(err, "failed to stat mixin (%s)", mixinPath)
	}
	if !exists {
		return errors.Errorf("mixin not found (%s)", mixinPath)
	}

	return nil
}

func (r *Runner) Run() error {
	if r.Debug {
		fmt.Fprintf(r.Err, "DEBUG mixin:    %s\n", r.Mixin)
		fmt.Fprintf(r.Err, "DEBUG mixinDir: %s\n", r.mixinDir)
		fmt.Fprintf(r.Err, "DEBUG file:     %s\n", r.File)
		fmt.Fprintf(r.Err, "DEBUG stdin:\n%s\n", r.Input)
	}

	mixinPath := r.getMixinPath()
	cmdArgs := strings.Split(r.Command, " ")
	command := cmdArgs[0]
	cmd := r.NewCommand(mixinPath, cmdArgs...)

	// Pipe the output from the mixin to porter
	cmd.Stdout = r.Context.Out
	cmd.Stderr = r.Context.Err

	if !IsCoreMixinCommand(command) {
		// For custom commands, don't call the mixin as "mixin CUSTOM" but as "mixin invoke --action CUSTOM"
		for i := range cmd.Args {
			if cmd.Args[i] == command {
				cmd.Args[i] = "invoke"
				break
			}
		}
		cmd.Args = append(cmd.Args, "--action", command)
	}

	if r.File != "" {
		cmd.Args = append(cmd.Args, "-f", r.File)
	}

	if r.Debug {
		cmd.Args = append(cmd.Args, "--debug")
	}

	if r.Input != "" {
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return err
		}
		go func() {
			defer stdin.Close()
			io.WriteString(stdin, r.Input)
		}()
	}

	prettyCmd := fmt.Sprintf("%s %s", cmd.Path, strings.Join(cmd.Args, " "))
	if r.Debug {
		fmt.Fprintln(r.Err, prettyCmd)
	}

	err := cmd.Start()
	if err != nil {
		return errors.Wrapf(err, "could not run mixin command %s", prettyCmd)
	}

	return cmd.Wait()
}

func (r *Runner) getMixinPath() string {
	path := filepath.Join(r.mixinDir, r.Mixin)
	if r.Runtime {
		return path + "-runtime"
	}
	return path + FileExt
}
