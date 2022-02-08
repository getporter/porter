package docker

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"get.porter.sh/porter/pkg/build"
	portercontext "get.porter.sh/porter/pkg/context"
	"get.porter.sh/porter/pkg/manifest"
	"github.com/docker/cli/cli/command"
	clibuild "github.com/docker/cli/cli/command/image/build"
	cliflags "github.com/docker/cli/cli/flags"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/moby/term"
	"github.com/pkg/errors"
)

type Builder struct {
	*portercontext.Context
}

func NewBuilder(cxt *portercontext.Context) *Builder {
	return &Builder{
		Context: cxt,
	}
}

func (b *Builder) BuildInvocationImage(manifest *manifest.Manifest) error {
	buildArgs := make(map[string]*string)
	buildArgs["BUNDLE_DIR"] = &build.BUNDLE_DIR

	convertedCustomInput := make(map[string]string)
	convertedCustomInput, err := convertMap(manifest.Custom)
	if err != nil {
		return err
	}

	for k, v := range convertedCustomInput {
		v := v
		buildArgs[strings.ToUpper(strings.Replace(k, ".", "_", -1))] = &v
	}

	fmt.Fprintf(b.Out, "\nStarting Invocation Image Build (%s) =======> \n", manifest.Image)
	buildOptions := types.ImageBuildOptions{
		SuppressOutput: false,
		PullParent:     false,
		Remove:         true,
		Tags:           []string{manifest.Image},
		Dockerfile:     filepath.ToSlash(build.DOCKER_FILE),
		BuildArgs:      buildArgs,
	}

	excludes, err := clibuild.ReadDockerignore(b.Getwd())
	if err != nil {
		return err
	}
	excludes = clibuild.TrimBuildFilesFromExcludes(excludes, buildOptions.Dockerfile, false)

	tar, err := archive.TarWithOptions(b.Getwd(), &archive.TarOptions{ExcludePatterns: excludes})
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

func (b *Builder) TagInvocationImage(origTag, newTag string) error {
	cli, err := command.NewDockerCli()
	if err != nil {
		return errors.Wrap(err, "could not create new docker client")
	}
	if err := cli.Initialize(cliflags.NewClientOptions()); err != nil {
		return err
	}

	if err := cli.Client().ImageTag(context.Background(), origTag, newTag); err != nil {
		return errors.Wrapf(err, "could not tag image %s with value %s", origTag, newTag)
	}
	return nil
}

func convertMap(mapInput map[string]interface{}) (map[string]string, error) {
	out := make(map[string]string)

	for key, value := range mapInput {
		switch v := value.(type) {
		case string:
			out[key] = v
		case map[string]interface{}:
			tmp, err := convertMap(v)
			if err != nil {
				return nil, err
			}
			for innerKey, innerValue := range tmp {
				out[key+"."+innerKey] = innerValue
			}
		case map[string]string:
			for innerKey, innerValue := range v {
				out[key+"."+innerKey] = innerValue
			}
		default:
			return nil, errors.Errorf("Unknown type %#v: %t", v, v)
		}
	}
	return out, nil
}
