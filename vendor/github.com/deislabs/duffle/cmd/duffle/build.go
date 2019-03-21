package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/docker/cli/cli/command"
	cliconfig "github.com/docker/cli/cli/config"
	dockerdebug "github.com/docker/cli/cli/debug"
	dockerflags "github.com/docker/cli/cli/flags"
	"github.com/docker/cli/opts"
	"github.com/docker/go-connections/tlsconfig"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/deislabs/duffle/pkg/builder"
	"github.com/deislabs/duffle/pkg/bundle"
	"github.com/deislabs/duffle/pkg/crypto/digest"
	"github.com/deislabs/duffle/pkg/duffle/home"
	"github.com/deislabs/duffle/pkg/duffle/manifest"
	"github.com/deislabs/duffle/pkg/imagebuilder"
	"github.com/deislabs/duffle/pkg/imagebuilder/docker"
	"github.com/deislabs/duffle/pkg/imagebuilder/mock"
	"github.com/deislabs/duffle/pkg/ohai"
	"github.com/deislabs/duffle/pkg/repo"
	"github.com/deislabs/duffle/pkg/signature"
)

const buildDesc = `
Builds a Cloud Native Application Bundle (CNAB) given a path to a directory that has a duffle build file (duffle.json).

It builds the invocation images specified in the duffle build file and then creates or updates the bundle in local storage with the latest invocation images.
`

const (
	dockerTLSEnvVar       = "DOCKER_TLS"
	dockerTLSVerifyEnvVar = "DOCKER_TLS_VERIFY"
)

var (
	dockerCertPath = os.Getenv("DOCKER_CERT_PATH")
)

type buildCmd struct {
	out        io.Writer
	src        string
	home       home.Home
	signer     string
	outputFile string

	// options common to the docker client and the daemon.
	dockerClientOptions *dockerflags.ClientOptions
}

func newBuildCmd(out io.Writer) *cobra.Command {
	build := &buildCmd{
		out:                 out,
		dockerClientOptions: dockerflags.NewClientOptions(),
	}
	var f *pflag.FlagSet

	cmd := &cobra.Command{
		Use:   "build [path]",
		Short: "build a bundle and invocation images",
		Long:  buildDesc,
		PersistentPreRun: func(c *cobra.Command, args []string) {
			build.dockerClientOptions.Common.SetDefaultOptions(f)
			dockerPreRun(build.dockerClientOptions)
		},
		RunE: func(_ *cobra.Command, args []string) (err error) {
			if len(args) > 0 {
				build.src = args[0]
			}
			if build.src == "" || build.src == "." {
				if build.src, err = os.Getwd(); err != nil {
					return err
				}
			}
			build.home = home.Home(homePath())
			return build.run()
		},
	}

	f = cmd.Flags()
	f.StringVarP(&build.signer, "user", "u", "", "the user ID of the signing key to use. Format is either email address or 'NAME (COMMENT) <EMAIL>'")
	f.StringVarP(&build.outputFile, "output-file", "o", "", "If set, writes the bundle to this file in addition to saving it to the local store")

	f.BoolVar(&build.dockerClientOptions.Common.Debug, "docker-debug", false, "Enable debug mode")
	f.StringVar(&build.dockerClientOptions.Common.LogLevel, "docker-log-level", "info", `Set the logging level ("debug"|"info"|"warn"|"error"|"fatal")`)
	f.BoolVar(&build.dockerClientOptions.Common.TLS, "docker-tls", defaultDockerTLS(), "Use TLS; implied by --tlsverify")
	f.BoolVar(&build.dockerClientOptions.Common.TLSVerify, fmt.Sprintf("docker-%s", dockerflags.FlagTLSVerify), defaultDockerTLSVerify(), "Use TLS and verify the remote")
	f.StringVar(&build.dockerClientOptions.ConfigDir, "docker-config", cliconfig.Dir(), "Location of client config files")

	build.dockerClientOptions.Common.TLSOptions = &tlsconfig.Options{
		CAFile:   filepath.Join(dockerCertPath, dockerflags.DefaultCaFile),
		CertFile: filepath.Join(dockerCertPath, dockerflags.DefaultCertFile),
		KeyFile:  filepath.Join(dockerCertPath, dockerflags.DefaultKeyFile),
	}

	tlsOptions := build.dockerClientOptions.Common.TLSOptions
	f.Var(opts.NewQuotedString(&tlsOptions.CAFile), "docker-tlscacert", "Trust certs signed only by this CA")
	f.Var(opts.NewQuotedString(&tlsOptions.CertFile), "docker-tlscert", "Path to TLS certificate file")
	f.Var(opts.NewQuotedString(&tlsOptions.KeyFile), "docker-tlskey", "Path to TLS key file")

	hostOpt := opts.NewNamedListOptsRef("docker-hosts", &build.dockerClientOptions.Common.Hosts, opts.ValidateHost)
	f.Var(hostOpt, "docker-host", "Daemon socket(s) to connect to")

	return cmd
}

func (b *buildCmd) run() (err error) {
	ctx := context.Background()
	bldr := builder.New()
	bldr.LogsDir = b.home.Logs()

	mfst, err := manifest.Load("", b.src)
	if err != nil {
		return err
	}

	imagebuilders, err := b.prepareImageBuilders(mfst)
	if err != nil {
		return fmt.Errorf("cannot configure necessary image builders: %v", err)
	}

	app, bf, err := bldr.PrepareBuild(bldr, mfst, b.src, imagebuilders)
	if err != nil {
		return fmt.Errorf("cannot prepare build: %v", err)
	}

	if err := bldr.Build(ctx, app); err != nil {
		return err
	}

	digest, err := b.writeBundle(bf)
	if err != nil {
		return err
	}

	// record the new bundle in repositories.json
	if err := recordBundleReference(b.home, bf.Name, bf.Version, digest); err != nil {
		return fmt.Errorf("could not record bundle: %v", err)
	}
	ohai.Fsuccessf(b.out, "Successfully built bundle %s:%s\n", bf.Name, bf.Version)

	return nil
}

func (b *buildCmd) writeBundle(bf *bundle.Bundle) (string, error) {
	kr, err := signature.LoadKeyRing(b.home.SecretKeyRing())
	if err != nil {
		return "", fmt.Errorf("cannot load keyring: %s", err)
	}

	if kr.Len() == 0 {
		return "", errors.New("no signing keys are present in the keyring")
	}

	// Default to the first key in the ring unless the user specifies otherwise.
	key := kr.Keys()[0]
	if b.signer != "" {
		key, err = kr.Key(b.signer)
		if err != nil {
			return "", err
		}
	}

	sign := signature.NewSigner(key)
	data, err := sign.Clearsign(bf)
	data = append(data, '\n')
	if err != nil {
		return "", fmt.Errorf("cannot sign bundle: %s", err)
	}

	digest, err := digest.OfBuffer(data)
	if err != nil {
		return "", fmt.Errorf("cannot compute digest from bundle: %v", err)
	}

	if b.outputFile != "" {
		if err := ioutil.WriteFile(b.outputFile, data, 0644); err != nil {
			return "", fmt.Errorf("cannot write bundle to %s: %v", b.outputFile, err)
		}
	}

	return digest, ioutil.WriteFile(filepath.Join(b.home.Bundles(), digest), data, 0644)
}

func defaultDockerTLS() bool {
	return os.Getenv(dockerTLSEnvVar) != ""
}

func defaultDockerTLSVerify() bool {
	return os.Getenv(dockerTLSVerifyEnvVar) != ""
}

func dockerPreRun(opts *dockerflags.ClientOptions) {
	dockerflags.SetLogLevel(opts.Common.LogLevel)

	if opts.ConfigDir != "" {
		cliconfig.SetDir(opts.ConfigDir)
	}

	if opts.Common.Debug {
		dockerdebug.Enable()
	}
}

// prepareImageBuilders returns the configured image builders that the bundle builder will need to build invocation images
func (b *buildCmd) prepareImageBuilders(mfst *manifest.Manifest) ([]imagebuilder.ImageBuilder, error) {
	var imagebuilders []imagebuilder.ImageBuilder
	for _, image := range mfst.InvocationImages {
		switch image.Builder {
		case "docker":
			// setup docker
			cli := &command.DockerCli{}
			if err := cli.Initialize(b.dockerClientOptions); err != nil {
				return imagebuilders, fmt.Errorf("failed to create docker client: %v", err)
			}
			imagebuilders = append(imagebuilders, docker.NewBuilder(image, cli))

		case "mock":
			imagebuilders = append(imagebuilders, mock.NewBuilder(image))
		}
	}
	return imagebuilders, nil
}

func recordBundleReference(home home.Home, name, version, digest string) error {
	// record the new bundle in repositories.json
	index, err := repo.LoadIndex(home.Repositories())
	if err != nil {
		return fmt.Errorf("cannot create or open %s: %v", home.Repositories(), err)
	}

	index.Add(name, version, digest)

	if err := index.WriteFile(home.Repositories(), 0644); err != nil {
		return fmt.Errorf("could not write to %s: %v", home.Repositories(), err)
	}

	return nil
}
