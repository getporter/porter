package mixin

import (
	"bytes"
	"context"
	"io/ioutil"
	"os/exec"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/pkgmgmt"
	"get.porter.sh/porter/pkg/pkgmgmt/client"
	"get.porter.sh/porter/pkg/tracing"
	"go.uber.org/zap/zapcore"
)

const (
	Directory = "mixins"
)

var _ MixinProvider = &PackageManager{}

// PackageManager handles package management for mixins.
type PackageManager struct {
	*client.FileSystem
}

func NewPackageManager(c *config.Config) *PackageManager {
	client := &PackageManager{
		FileSystem: client.NewFileSystem(c, Directory),
	}
	client.PreRun = client.PreRunMixinCommandHandler
	client.BuildMetadata = func() pkgmgmt.PackageMetadata {
		return &Metadata{}
	}
	return client
}

func (c *PackageManager) PreRunMixinCommandHandler(command string, cmd *exec.Cmd) {
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
}

func (c *PackageManager) GetSchema(ctx context.Context, name string) (string, error) {
	log := tracing.LoggerFromContext(ctx)

	mixinDir, err := c.GetPackageDir(name)
	if err != nil {
		return "", err
	}

	r := client.NewRunner(name, mixinDir, false)

	// Copy the existing context and tweak to pipe the output differently
	mixinSchema := &bytes.Buffer{}
	mixinContext := *c.Context
	mixinContext.Out = mixinSchema
	if !log.ShouldLog(zapcore.DebugLevel) {
		mixinContext.Err = ioutil.Discard
	}
	r.Context = &mixinContext

	cmd := pkgmgmt.CommandOptions{Command: "schema", PreRun: c.PreRun}
	err = r.Run(ctx, cmd)
	if err != nil {
		return "", err
	}

	return mixinSchema.String(), nil
}

var _ pkgmgmt.PackageMetadata = Metadata{}

// Metadata about an installed mixin.
type Metadata = pkgmgmt.Metadata
