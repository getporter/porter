package buildkit

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"get.porter.sh/porter/pkg/build"
	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/tracing"
	"github.com/cnabio/cnab-go/driver/docker"
	buildx "github.com/docker/buildx/build"
	"github.com/docker/buildx/builder"
	"github.com/docker/buildx/controller/pb"
	_ "github.com/docker/buildx/driver/docker" // Register the docker driver with buildkit
	"github.com/docker/buildx/util/buildflags"
	"github.com/docker/buildx/util/confutil"
	"github.com/docker/buildx/util/dockerutil"
	"github.com/docker/buildx/util/progress"
	dockerconfig "github.com/docker/cli/cli/config"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/session/auth/authprovider"
	"github.com/moby/buildkit/util/progress/progressui"
	"go.opentelemetry.io/otel/attribute"
)

// MaxArgLength is the maximum length of a build argument that can be passed to Docker
// Each system has its own max (see https://stackoverflow.com/questions/70737793/max-size-of-string-cmd-that-can-be-passed-to-docker)
// I am choosing 5,000 characters because that's lower than all supported OS/ARCH combinations
const MaxArgLength = 5000

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

func (b *Builder) BuildBundleImage(ctx context.Context, manifest *manifest.Manifest, opts build.BuildImageOptions) error {
	ctx, span := tracing.StartSpan(ctx, attribute.String("image", manifest.Image))
	defer span.EndSpan()

	span.Info("Building bundle image")

	cli, err := docker.GetDockerClient()
	if err != nil {
		return span.Error(err)
	}

	bldr, err := builder.New(cli,
		builder.WithName(cli.CurrentContext()),
		builder.WithContextPathHash(b.Getwd()),
	)
	if err != nil {
		return span.Error(err)
	}
	nodes, err := bldr.LoadNodes(ctx)
	if err != nil {
		return span.Error(err)
	}

	currentSession := []session.Attachable{authprovider.NewDockerAuthProvider(dockerconfig.LoadDefaultConfigFile(b.Err), make(map[string]*authprovider.AuthTLSConfig))}

	ssh, err := buildflags.ParseSSHSpecs(opts.SSH)
	if err != nil {
		return span.Errorf("error parsing the --ssh flags: %w", err)
	}

	pbssh, err := pb.CreateSSH(ssh)
	if err != nil {
		return span.Errorf("error creating ssh ", err)
	}

	currentSession = append(currentSession, pbssh)

	secrets, err := buildflags.ParseSecretSpecs(opts.Secrets)
	if err != nil {
		return span.Errorf("error parsing the --secret flags: %w", err)
	}
	pbsecrets, err := pb.CreateSecrets(secrets)
	if err != nil {
		return span.Errorf("error creating secrets %w", err)
	}

	currentSession = append(currentSession, pbsecrets)

	args, err := b.determineBuildArgs(ctx, manifest, opts)
	if err != nil {
		return err
	}
	span.SetAttributes(tracing.ObjectAttribute("build-args", args))

	buildContexts, err := buildflags.ParseContextNames(opts.BuildContexts)
	if err != nil {
		return span.Errorf("error parsing the --build-context flags: %w", err)
	}

	buildxOpts := map[string]buildx.Options{
		"default": {
			Tags: []string{manifest.Image},
			Inputs: buildx.Inputs{
				ContextPath:    b.Getwd(),
				DockerfilePath: b.getDockerfilePath(),
				InStream:       buildx.NewSyncMultiReader(b.In),
				NamedContexts:  toNamedContexts(buildContexts),
			},
			BuildArgs: args,
			Session:   currentSession,
			NoCache:   opts.NoCache,
		},
	}

	mode := progressui.AutoMode // Auto writes to stderr regardless of what you pass in
	printer, err := progress.NewPrinter(ctx, os.Stderr, mode)
	if err != nil {
		return span.Error(err)
	}

	_, buildErr := buildx.Build(ctx, nodes, buildxOpts, dockerutil.NewClient(cli), confutil.NewConfig(cli), printer)
	printErr := printer.Wait()

	if buildErr == nil && printErr != nil {
		return span.Errorf("error with docker printer: %w", printErr)
	}

	if buildErr != nil {
		return span.Errorf("error building docker image: %w", buildErr)
	}

	return nil
}

func (b *Builder) getDockerfilePath() string {
	return filepath.Join(b.Getwd(), build.DOCKER_FILE)
}

func (b *Builder) determineBuildArgs(
	ctx context.Context,
	manifest *manifest.Manifest,
	opts build.BuildImageOptions) (map[string]string, error) {

	_, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	// This will grow later when we add custom build args from the porter.yaml
	args := make(map[string]string, len(opts.BuildArgs)+1)

	// Create a map of key/values from the custom field in porter.yaml
	convertedCustomInput, err := flattenMap(manifest.Custom)
	if err != nil {
		return nil, span.Error(err)
	}

	// Determine which (if any) custom fields from porter.yaml are used in the Dockerfile
	dockerfilePath := b.getDockerfilePath()
	dockerfileContents, err := b.FileSystem.ReadFile(dockerfilePath)
	if err != nil {
		return nil, span.Errorf("Error reading Dockerfile at %s: %w", dockerfilePath, err)
	}
	customBuildArgs, err := detectCustomBuildArgsUsed(string(dockerfileContents))
	if err != nil {
		return nil, span.Errorf("Error parsing custom build arguments from the Dockerfile at %s: %w", dockerfilePath, err)
	}

	// Pass custom values as build args when building the bundle image
	argNameRegex := regexp.MustCompile(`[^A-Z0-9_]`)
	for k, v := range convertedCustomInput {
		// Make all arg names upper-case
		argName := fmt.Sprintf("CUSTOM_%s", strings.ToUpper(k))

		// replace characters that can't be in an argument name with _
		argName = argNameRegex.ReplaceAllString(argName, "_")

		// Only add build args for custom values used in the Dockerfile
		if _, ok := customBuildArgs[argName]; ok {
			args[argName] = v
		}
	}

	// Add explicit build arguments next they should override what was determined from the porter.yaml to allow the user to fix anything unexpected
	parseBuildArgs(opts.BuildArgs, args)

	// Add porter defined build arguments last
	args["BUNDLE_DIR"] = build.BUNDLE_DIR

	// Check if any arguments are too long
	for k, v := range args {
		if len(v) > MaxArgLength {
			return nil, span.Errorf("The length of the build argument %s is longer than the max (%d characters). Save the value to a file in the bundle directory, and then read the file contents out in a custom dockerfile or in the bundle at runtime to work around this limitation.", k, MaxArgLength)
		}
	}

	return args, nil
}

func detectCustomBuildArgsUsed(dockerFileContents string) (map[string]struct{}, error) {
	customBuildArgRegex := regexp.MustCompile(`ARG (CUSTOM_([a-zA-Z0-9_]+))`)

	matches := customBuildArgRegex.FindAllStringSubmatch(dockerFileContents, -1)
	argNames := make(map[string]struct{}, len(matches))
	for _, match := range matches {
		argNames[match[1]] = struct{}{}
	}

	return argNames, nil
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

func toNamedContexts(m map[string]string) map[string]buildx.NamedContext {
	m2 := make(map[string]buildx.NamedContext, len(m))
	for k, v := range m {
		m2[k] = buildx.NamedContext{Path: v}
	}
	return m2
}

func (b *Builder) TagBundleImage(ctx context.Context, origTag, newTag string) error {
	ctx, log := tracing.StartSpan(ctx, attribute.String("source-tag", origTag), attribute.String("destination-tag", newTag))
	defer log.EndSpan()

	cli, err := docker.GetDockerClient()
	if err != nil {
		return log.Error(err)
	}

	if err := cli.Client().ImageTag(ctx, origTag, newTag); err != nil {
		return log.Errorf("could not tag image %s with value %s: %w", origTag, newTag, err)
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
