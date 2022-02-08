package buildkit

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/containerd/console"

	"get.porter.sh/porter/pkg/build"
	portercontext "get.porter.sh/porter/pkg/context"
	"get.porter.sh/porter/pkg/manifest"
	buildx "github.com/docker/buildx/build"
	"github.com/docker/buildx/driver"
	_ "github.com/docker/buildx/driver/docker" // Register the docker driver with buildkit
	"github.com/docker/buildx/util/progress"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/context/docker"
	cliflags "github.com/docker/cli/cli/flags"
	dockerclient "github.com/docker/docker/client"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/session/auth/authprovider"
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
	fmt.Fprintf(b.Out, "\nStarting Invocation Image Build (%s) =======> \n", manifest.Image)
	cli, err := command.NewDockerCli()
	if err != nil {
		return errors.Wrap(err, "could not create new docker client")
	}
	if err := cli.Initialize(cliflags.NewClientOptions()); err != nil {
		return errors.Wrapf(err, "error initializing docker client")
	}

	ctx := context.Background()
	d, err := driver.GetDriver(ctx, "porter-driver", nil, cli.Client(), cli.ConfigFile(), nil, nil, "", nil, nil, b.Getwd())
	if err != nil {
		return errors.Wrapf(err, "error loading buildx driver")
	}
	drivers := []buildx.DriverInfo{
		{
			Name:   "default",
			Driver: d,
		},
	}

	buildArgs := make(map[string]string)
	buildArgs["BUNDLE_DIR"] = build.BUNDLE_DIR

	convertedCustomInput := make(map[string]string)
	convertedCustomInput, err = convertMap(manifest.Custom)
	if err != nil {
		return err
	}

	for k, v := range convertedCustomInput {
		buildArgs[strings.ToUpper(strings.Replace(k, ".", "_", -1))] = v
	}

	opts := map[string]buildx.Options{
		"default": {
			Tags: []string{manifest.Image},
			Inputs: buildx.Inputs{
				ContextPath:    b.Getwd(),
				DockerfilePath: filepath.Join(b.Getwd(), build.DOCKER_FILE),
				InStream:       b.In,
			},
			BuildArgs: buildArgs,
			Session:   []session.Attachable{authprovider.NewDockerAuthProvider(b.Err)},
		},
	}

	out := ioutil.Discard
	if b.IsVerbose() {
		out = b.Out
	}
	dockerOut, err := getConsole(out)
	if err != nil {
		return errors.Wrap(err, "could not retrieve docker build output")
	}
	defer dockerOut.Close()

	printer := progress.NewPrinter(ctx, dockerOut, "auto")
	_, buildErr := buildx.Build(ctx, drivers, opts, dockerToBuildx{cli}, cli.ConfigFile(), printer)
	printErr := printer.Wait()

	if buildErr == nil {
		return errors.Wrapf(printErr, "error with docker printer")
	}
	return errors.Wrapf(buildErr, "error building docker image")
}

var _ console.File = dockerConsole{}

// Wraps io.Writer since docker wants to send output to a *os.File
type dockerConsole struct {
	out io.Writer
	f   *os.File
}

func getConsole(out io.Writer) (dockerConsole, error) {
	f, ok := out.(*os.File)
	if ok {
		return dockerConsole{
			f:   f,
			out: out,
		}, nil
	}

	f, err := ioutil.TempFile("", "porter-output-dockerConsole")
	if err != nil {
		return dockerConsole{}, err
	}

	return dockerConsole{
		f:   f,
		out: io.MultiWriter(out, f),
	}, nil
}

func (r dockerConsole) Read(p []byte) (n int, err error) {
	return r.f.Read(p)
}

func (r dockerConsole) Write(p []byte) (n int, err error) {
	return r.out.Write(p)
}

func (r dockerConsole) Close() error {
	r.f.Close()
	os.Remove(r.f.Name())
	return nil
}

func (r dockerConsole) Fd() uintptr {
	return r.f.Fd()
}

func (r dockerConsole) Name() string {
	return r.f.Name()
}

// Adapts between Docker CLI and Buildx
type dockerToBuildx struct {
	cli command.Cli
}

func (d dockerToBuildx) DockerAPI(_ string) (dockerclient.APIClient, error) {
	endpoint := docker.Endpoint{}
	endpoint.Host = d.cli.CurrentContext()

	clientOpts, err := endpoint.ClientOpts()
	if err != nil {
		return nil, err
	}

	return dockerclient.NewClientWithOpts(clientOpts...)
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
