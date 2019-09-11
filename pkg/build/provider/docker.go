package provider

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/deislabs/porter/pkg/build"
	"github.com/deislabs/porter/pkg/config"
	"github.com/docker/cli/cli/command"
	cliflags "github.com/docker/cli/cli/flags"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	"github.com/pkg/errors"
)

type DockerBuilder struct {
	*config.Config
}

func NewDockerBuilder(cfg *config.Config) *DockerBuilder {
	return &DockerBuilder{
		Config: cfg,
	}
}

func (b *DockerBuilder) BuildInvocationImage() error {
	fmt.Fprintf(b.Out, "\nStarting Invocation Image Build =======> \n")
	path, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "could not get current working directory")
	}
	buildOptions := types.ImageBuildOptions{
		SuppressOutput: false,
		PullParent:     false,
		Tags:           []string{b.Manifest.Image},
		Dockerfile:     "Dockerfile",
		BuildArgs: map[string]*string{
			"BUNDLE_DIR": &build.BUNDLE_DIR,
		},
	}
	tar, err := archive.TarWithOptions(path, &archive.TarOptions{})
	if err != nil {
		return err
	}

	cli, err := command.NewDockerCli()
	if err != nil {
		return errors.Wrap(err, "could not create new docker client")
	}
	if err := cli.Initialize(cliflags.NewClientOptions()); err != nil {
		return err
	}

	response, err := cli.Client().ImageBuild(context.Background(), tar, buildOptions)
	if err != nil {
		return err
	}

	dockerOutput := ioutil.Discard
	if b.IsVerbose() {
		dockerOutput = b.Out
	}

	termFd, _ := term.GetFdInfo(dockerOutput)
	// Setting this to false here because Moby os.Exit(1) all over the place and this fails on WSL (only)
	// when Term is true.
	isTerm := false
	err = jsonmessage.DisplayJSONMessagesStream(response.Body, dockerOutput, termFd, isTerm, nil)
	if err != nil {
		return errors.Wrap(err, "failed to stream docker build output")
	}
	return nil
}
