package mixin

import (
	"fmt"
	"io"
	"os/exec"
	"path/filepath"

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
	Data    string
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
		fmt.Fprintf(r.Out, "DEBUG mixin:    %s\n", r.Mixin)
		fmt.Fprintf(r.Out, "DEBUG mixinDir: %s\n", r.mixinDir)
		fmt.Fprintf(r.Out, "DEBUG file:     %s\n", r.File)
		fmt.Fprintf(r.Out, "DEBUG stdin:\n%s\n", r.Data)
	}

	mixinPath := r.getMixinPath()
	cmd := exec.Command(mixinPath, r.Command)

	// Pipe the output from the mixin to porter
	cmd.Stdout = r.Context.Out
	cmd.Stderr = r.Context.Err

	if r.File != "" {
		cmd.Args = append(cmd.Args, "-f", r.File)
	}

	if r.Data != "" {
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return err
		}
		go func() {
			defer stdin.Close()
			io.WriteString(stdin, r.Data)
		}()
	}

	err := cmd.Start()
	if err != nil {
		return errors.Wrapf(err, "could not run mixin command %v", cmd)
	}

	return cmd.Wait()
}

func (r *Runner) getMixinPath() string {
	path := filepath.Join(r.mixinDir, r.Mixin)
	if r.Runtime {
		return path + "-runtime"
	}
	return path
}
