package mixin

import (
	"fmt"
	"io"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/deislabs/porter/pkg/context"
)

type Runner struct {
	context.Context
	// dir is the absolute path to the directory containing the mixin
	dir string

	Mixin   string
	Command string
	Data    string
}

func NewRunner(mixin, command, data, dir string) *Runner {
	return &Runner{
		Context: context.New(),
		dir:     dir,
		Mixin:   mixin,
		Command: command,
		Data:    data,
	}
}

func (r *Runner) Validate() error {
	mixinPath := r.getMixinPath()
	exists, err := r.FileSystem.Exists(mixinPath)
	if !exists {
		return errors.Wrapf(err, "mixin not found (%s)", mixinPath)
	}
	return nil
}

func (r *Runner) Run() error {
	if r.Debug {
		fmt.Fprintf(r.Out, "DEBUG cwd:   %s\n", r.dir)
		fmt.Fprintf(r.Out, "DEBUG mixin: %s\n", r.Mixin)
		fmt.Fprintf(r.Out, "DEBUG stdin:\n%s\n", r.Data)
	}

	mixinPath := r.getMixinPath()
	cmd := exec.Command(mixinPath, r.Command)
	cmd.Dir = r.dir
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, r.Data)
	}()

	err = cmd.Start()
	if err != nil {
		return errors.Wrapf(err, "could not run mixin command %v", cmd)
	}

	return cmd.Wait()
}

func (r *Runner) getMixinPath() string {
	return filepath.Join(r.dir, r.Mixin)
}
