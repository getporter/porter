package mixinprovider

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/deislabs/porter/pkg/context"
	"github.com/deislabs/porter/pkg/mixin"
	"github.com/pkg/errors"
)

type Runner struct {
	*context.Context
	// mixinDir is the absolute path to the directory containing the mixin
	mixinDir string

	mixin   string
	runtime bool
}

func NewRunner(mixin, mixinDir string, runtime bool) *Runner {
	return &Runner{
		Context:  context.New(),
		mixin:    mixin,
		runtime:  runtime,
		mixinDir: mixinDir,
	}
}

func (r *Runner) Validate() error {
	if r.mixin == "" {
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

func (r *Runner) Run(commandOpts mixin.CommandOptions) error {
	if r.Debug {
		fmt.Fprintf(r.Err, "DEBUG mixin:    %s\n", r.mixin)
		fmt.Fprintf(r.Err, "DEBUG mixinDir: %s\n", r.mixinDir)
		fmt.Fprintf(r.Err, "DEBUG file:     %s\n", commandOpts.File)
		fmt.Fprintf(r.Err, "DEBUG stdin:\n%s\n", commandOpts.Input)
	}

	mixinPath := r.getMixinPath()
	cmdArgs := strings.Split(commandOpts.Command, " ")
	command := cmdArgs[0]
	cmd := r.NewCommand(mixinPath, cmdArgs...)

	// Pipe the output from the mixin to porter
	cmd.Stdout = r.Context.Out
	cmd.Stderr = r.Context.Err

	if !mixin.IsCoreMixinCommand(command) {
		// For custom commands, don't call the mixin as "mixin CUSTOM" but as "mixin invoke --action CUSTOM"
		for i := range cmd.Args {
			if cmd.Args[i] == command {
				cmd.Args[i] = "invoke"
				break
			}
		}
		cmd.Args = append(cmd.Args, "--action", command)
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
		return errors.Wrapf(err, "could not run mixin command %s", prettyCmd)
	}

	return cmd.Wait()
}

func (r *Runner) getMixinPath() string {
	path := filepath.Join(r.mixinDir, r.mixin)
	if r.runtime {
		return path + "-runtime"
	}
	return path + mixin.FileExt
}
