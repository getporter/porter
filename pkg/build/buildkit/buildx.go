package buildkit

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"get.porter.sh/porter/pkg/build"
	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/tracing"
	"github.com/cnabio/cnab-go/driver/docker"
	buildx "github.com/docker/buildx/build"
	"github.com/docker/buildx/driver"
	_ "github.com/docker/buildx/driver/docker" // Register the docker driver with buildkit
	"github.com/docker/buildx/store/storeutil"
	"github.com/docker/buildx/util/buildflags"
	"github.com/docker/buildx/util/confutil"
	"github.com/docker/buildx/util/progress"
	"github.com/docker/cli/cli/command"
	dockercontext "github.com/docker/cli/cli/context/docker"
	dockerclient "github.com/docker/docker/client"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/session/auth/authprovider"
	"go.opentelemetry.io/otel/attribute"
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
	logger tracing.TraceLogger
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

func (b *Builder) BuildInvocationImage(ctx context.Context, manifest *manifest.Manifest, opts build.BuildImageOptions) error {
	ctx, log := tracing.StartSpan(ctx, attribute.String("image", manifest.Image))
	defer log.EndSpan()

	log.Info("Building invocation image")

	cli, err := docker.GetDockerClient()
	if err != nil {
		return log.Error(err)
	}

	imageopt, err := storeutil.GetImageConfig(cli, nil)
	if err != nil {
		return log.Error(err)
	}

	d, err := driver.GetDriver(ctx, "porter-driver", nil, cli.Client(), imageopt.Auth, nil, nil, nil, nil, nil, b.Getwd())
	if err != nil {
		return log.Error(fmt.Errorf("error loading buildx driver: %w", err))
	}

	drivers := []buildx.DriverInfo{
		{
			Name:   "default",
			Driver: d,
			// Use any proxies specified in the docker config file
			ProxyConfig: storeutil.GetProxyConfig(cli),
			// Use stored logins from the docker config to pull from private repositories
			ImageOpt: imageopt,
		},
	}

	session := []session.Attachable{authprovider.NewDockerAuthProvider(b.Err)}
	ssh, err := buildflags.ParseSSHSpecs(opts.SSH)
	if err != nil {
		return fmt.Errorf("error parsing the --ssh flags: %w", err)
	}
	session = append(session, ssh)

	secrets, err := buildflags.ParseSecretSpecs(opts.Secrets)
	if err != nil {
		return fmt.Errorf("error parsing the --secret flags: %w", err)
	}
	session = append(session, secrets)

	args := make(map[string]string, len(opts.BuildArgs)+1)
	parseBuildArgs(opts.BuildArgs, args)
	args["BUNDLE_DIR"] = build.BUNDLE_DIR

	convertedCustomInput, err := flattenMap(manifest.Custom)
	if err != nil {
		return log.Error(err)
	}

	for k, v := range convertedCustomInput {
		args[strings.ToUpper(strings.Replace(k, ".", "_", -1))] = v
	}

	buildxOpts := map[string]buildx.Options{
		"default": {
			Tags: []string{manifest.Image},
			Inputs: buildx.Inputs{
				ContextPath:    b.Getwd(),
				DockerfilePath: filepath.Join(b.Getwd(), build.DOCKER_FILE),
				InStream:       b.In,
			},
			BuildArgs: args,
			Session:   session,
			NoCache:   opts.NoCache,
		},
	}

	mode := progress.PrinterModeAuto // Auto writes to stderr regardless of what you pass in
	out := unstructuredLogger{log}
	printer := progress.NewPrinter(ctx, out, os.Stderr, mode)
	_, buildErr := buildx.Build(ctx, drivers, buildxOpts, dockerToBuildx{cli}, confutil.ConfigDir(cli), printer)
	printErr := printer.Wait()

	if buildErr == nil && printErr != nil {
		return log.Error(fmt.Errorf("error with docker printer: %w", printErr))
	}

	if buildErr != nil {
		return log.Error(fmt.Errorf("error building docker image: %w", buildErr))
	}

	return nil
}

func parseBuildArgs(unparsed []string, parsed map[string]string) {
	for _, arg := range unparsed {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) < 2 {
			// docker ignores --build-arg with only one part, so we will too
			continue
		}

		name := parts[0]
		value := parts[1]
		parsed[name] = value
	}
}

// Adapts between Docker CLI and Buildx
type dockerToBuildx struct {
	cli command.Cli
}

func (d dockerToBuildx) DockerAPI(_ string) (dockerclient.APIClient, error) {
	endpoint := dockercontext.Endpoint{}
	endpoint.Host = d.cli.CurrentContext()

	clientOpts, err := endpoint.ClientOpts()
	if err != nil {
		return nil, err
	}

	return dockerclient.NewClientWithOpts(clientOpts...)
}

func (b *Builder) TagInvocationImage(ctx context.Context, origTag, newTag string) error {
	ctx, log := tracing.StartSpan(ctx, attribute.String("source-tag", origTag), attribute.String("destination-tag", newTag))
	defer log.EndSpan()

	cli, err := docker.GetDockerClient()
	if err != nil {
		return log.Error(err)
	}

	if err := cli.Client().ImageTag(ctx, origTag, newTag); err != nil {
		return log.Error(fmt.Errorf("could not tag image %s with value %s: %w", origTag, newTag, err))
	}
	return nil
}

// flattenMap recursively walks through nested map and flattens it
// to one-level map of key-value with string type.
func flattenMap(mapInput map[string]interface{}) (map[string]string, error) {
	out := make(map[string]string)

	for key, value := range mapInput {
		switch v := value.(type) {
		case map[string]interface{}:
			tmp, err := flattenMap(v)
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
			innerValue, err := cnab.WriteParameterToString(key, value)
			if err != nil {
				return nil, fmt.Errorf("error representing %s as a build argument: %w", key, err)
			}
			out[key] = innerValue
		}
	}
	return out, nil
}
