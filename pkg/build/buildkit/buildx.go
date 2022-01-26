package buildkit

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/experimental"
	"get.porter.sh/porter/pkg/tracing"
	"go.opentelemetry.io/otel/attribute"

	"get.porter.sh/porter/pkg/build"
	"get.porter.sh/porter/pkg/manifest"
	"github.com/containerd/console"
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
	*config.Config
}

func NewBuilder(cfg *config.Config) *Builder {
	return &Builder{
		Config: cfg,
	}
}

var _ io.Writer = unstructuredLogger{}

// take lines from the docker output, and write them as info messages
// This allows the docker library to use our logger like an io.Writer
type unstructuredLogger struct {
	logger tracing.ScopedLogger
}

var newline = []byte("\n")

func (l unstructuredLogger) Write(p []byte) (n int, err error) {
	if l.logger == nil {
		return 0, nil
	}

	msg := string(bytes.TrimSuffix(p, newline))
	l.logger.Info(msg)
	return len(p), nil
}

func (b *Builder) BuildInvocationImage(ctx context.Context, manifest *manifest.Manifest) error {
	log := tracing.LoggerFromContext(ctx)
	ctx, log = log.StartSpan(attribute.String("image", manifest.Image))
	defer log.EndSpan()

	log.Info("Building invocation image")

	cli, err := command.NewDockerCli()
	if err != nil {
		return log.Error(errors.Wrap(err, "could not create new docker client"))
	}
	if err := cli.Initialize(cliflags.NewClientOptions()); err != nil {
		return log.Error(errors.Wrapf(err, "error initializing docker client"))
	}

	d, err := driver.GetDriver(ctx, "porter-driver", nil, cli.Client(), cli.ConfigFile(), nil, nil, "", nil, nil, b.Getwd())
	if err != nil {
		return log.Error(errors.Wrapf(err, "error loading buildx driver"))
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
	if b.IsVerbose() || b.Config.IsFeatureEnabled(experimental.FlagStructuredLogs) {
		ctx, log = log.StartSpanWithName("buildkit", attribute.String("source", "porter.build.buildkit"))
		defer log.EndSpan()
		out = unstructuredLogger{log}
	}

	dockerOut, err := getConsole(out)
	if err != nil {
		return log.Error(errors.Wrap(err, "could not retrieve docker build output"))
	}
	defer dockerOut.Close()

	printer := progress.NewPrinter(ctx, dockerOut, "auto")
	_, buildErr := buildx.Build(ctx, drivers, opts, dockerToBuildx{cli}, cli.ConfigFile(), printer)
	printErr := printer.Wait()

	if buildErr == nil {
		return log.Error(errors.Wrapf(printErr, "error with docker printer"))
	}
	return log.Error(errors.Wrapf(buildErr, "error building docker image"))
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

func (b *Builder) TagInvocationImage(ctx context.Context, origTag, newTag string) error {
	log := tracing.LoggerFromContext(ctx)
	cli, err := command.NewDockerCli()
	if err != nil {
		return log.Error(errors.Wrap(err, "could not create new docker client"))
	}
	if err := cli.Initialize(cliflags.NewClientOptions()); err != nil {
		return log.Error(errors.Wrap(err, "could not initialize a new docker client"))
	}

	if err := cli.Client().ImageTag(ctx, origTag, newTag); err != nil {
		return log.Error(errors.Wrapf(err, "could not tag image %s with value %s", origTag, newTag))
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
