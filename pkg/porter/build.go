package porter

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/deislabs/porter/pkg/build"
	configadapter "github.com/deislabs/porter/pkg/cnab/config_adapter"
	cxt "github.com/deislabs/porter/pkg/context"
	"github.com/deislabs/porter/pkg/mixin"
	"github.com/docker/cli/cli/command"
	cliflags "github.com/docker/cli/cli/flags"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	"github.com/pkg/errors"
)

// BUNDLE_DIR is the directory where the bundle is located in the CNAB execution environment.
var BUNDLE_DIR string = "/cnab/app"

type BuildOptions struct {
	contextOptions
}

func (p *Porter) Build(opts BuildOptions) error {
	opts.Apply(p.Context)

	err := p.LoadManifest()
	if err != nil {
		return err
	}

	if err := p.prepareDockerFilesystem(); err != nil {
		return fmt.Errorf("unable to copy mixins: %s", err)
	}
	if err := p.generateDockerFile(); err != nil {
		return fmt.Errorf("unable to generate Dockerfile: %s", err)
	}
	err = p.buildInvocationImage(context.Background())
	if err != nil {
		return errors.Wrap(err, "unable to build CNAB invocation image")
	}

	return p.buildBundle(p.Config.Manifest.Image, "")
}

func (p *Porter) generateDockerFile() error {
	lines, err := p.buildDockerfile()
	if err != nil {
		return errors.Wrap(err, "error generating the Dockerfile")
	}

	fmt.Fprintf(p.Out, "\nWriting Dockerfile =======>\n")
	contents := strings.Join(lines, "\n")

	if p.IsVerbose() {
		fmt.Fprintln(p.Out, contents)
	}

	err = p.Config.FileSystem.WriteFile("Dockerfile", []byte(contents), 0644)
	return errors.Wrap(err, "couldn't write the Dockerfile")
}

func (p *Porter) buildDockerfile() ([]string, error) {
	fmt.Fprintf(p.Out, "\nGenerating Dockerfile =======>\n")

	lines, err := p.getBaseDockerfile()
	if err != nil {
		return nil, err
	}

	mixinLines, err := p.buildMixinsSection()
	if err != nil {
		return nil, errors.Wrap(err, "error generating Dockefile content for mixins")
	}
	lines = append(lines, mixinLines...)

	// The template dockerfile copies everything by default, but if the user
	// supplied their own, copy over cnab/ and porter.yaml
	if p.Manifest.Dockerfile != "" {
		lines = append(lines, p.buildCNABSection()...)
		lines = append(lines, p.buildPorterSection()...)
	}
	lines = append(lines, p.buildWORKDIRSection())
	lines = append(lines, p.buildCMDSection())

	if p.IsVerbose() {
		for _, line := range lines {
			fmt.Fprintln(p.Out, line)
		}
	}

	return lines, nil
}

func (p *Porter) getBaseDockerfile() ([]string, error) {
	var reader io.Reader
	if p.Manifest.Dockerfile != "" {
		exists, err := p.FileSystem.Exists(p.Manifest.Dockerfile)
		if err != nil {
			return nil, errors.Wrapf(err, "error checking if Dockerfile exists: %q", p.Manifest.Dockerfile)
		}
		if !exists {
			return nil, errors.Errorf("the Dockerfile specified in the manifest doesn't exist: %q", p.Manifest.Dockerfile)
		}

		file, err := p.FileSystem.Open(p.Manifest.Dockerfile)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		reader = file

	} else {
		contents, err := p.Templates.GetDockerfile()
		if err != nil {
			return nil, errors.Wrap(err, "error loading default Dockerfile template")
		}
		reader = bytes.NewReader(contents)
	}

	var lines []string
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, nil
}

func (p *Porter) buildPorterSection() []string {
	return []string{
		`COPY porter.yaml $BUNDLE_DIR/porter.yaml`,
	}
}

func (p *Porter) buildCNABSection() []string {
	return []string{
		`COPY .cnab/ /cnab/`,
	}
}

func (p *Porter) buildWORKDIRSection() string {
	return `WORKDIR $BUNDLE_DIR`
}

func (p *Porter) buildCMDSection() string {
	return `CMD ["/cnab/app/run"]`
}

func (p *Porter) buildMixinsSection() ([]string, error) {
	lines := make([]string, 0)
	for _, m := range p.Manifest.Mixins {
		mixinDir, err := p.GetMixinDir(m)
		if err != nil {
			return nil, err
		}

		r := mixin.NewRunner(m, mixinDir, false)
		r.Command = "build"
		r.Input = "" // TODO: let the mixin know about which steps will be executed so that it can be more selective about copying into the invocation image

		// Copy the existing context and tweak to pipe the output differently
		mixinStdout := &bytes.Buffer{}
		var mixinContext cxt.Context
		mixinContext = *p.Context
		mixinContext.Out = mixinStdout   // mixin stdout -> dockerfile lines
		mixinContext.Err = p.Context.Out // mixin stderr -> logs
		r.Context = &mixinContext

		err = r.Validate()
		if err != nil {
			return nil, err
		}

		err = r.Run()
		if err != nil {
			return nil, err
		}

		l := strings.Split(mixinStdout.String(), "\n")
		lines = append(lines, l...)
	}
	return lines, nil
}

func (p *Porter) prepareDockerFilesystem() error {
	// clean up previously generated files
	p.FileSystem.RemoveAll(build.LOCAL_CNAB)
	p.FileSystem.Remove("Dockerfile")

	fmt.Fprintf(p.Out, "Copying porter runtime ===> \n")

	runTmpl, err := p.Templates.GetRunScript()
	if err != nil {
		return err
	}

	err = p.FileSystem.MkdirAll(build.LOCAL_APP, 0755)
	if err != nil {
		return err
	}

	err = p.FileSystem.WriteFile(build.LOCAL_RUN, runTmpl, 0755)
	if err != nil {
		return errors.Wrapf(err, "failed to write %s", build.LOCAL_RUN)
	}

	pr, err := p.GetPorterRuntimePath()
	if err != nil {
		return err
	}
	err = p.CopyFile(pr, filepath.Join(build.LOCAL_APP, "porter-runtime"))
	if err != nil {
		return err
	}

	fmt.Fprintf(p.Out, "Copying mixins ===> \n")
	for _, mixin := range p.Manifest.Mixins {
		err := p.copyMixin(mixin)
		if err != nil {
			return err
		}
	}

	// Create the local outputs directory, if it doesn't already exist
	outputsDir, err := p.Config.GetOutputsDir()
	if err != nil {
		return errors.Wrap(err, "unable to get outputs directory")
	}

	err = p.FileSystem.MkdirAll(outputsDir, 0755)
	if err != nil {
		return errors.Wrapf(err, "could not create outputs directory %s", outputsDir)
	}

	return nil
}

func (p *Porter) copyMixin(mixin string) error {
	fmt.Fprintf(p.Out, "Copying mixin %s ===> \n", mixin)
	mixinDir, err := p.GetMixinDir(mixin)
	if err != nil {
		return err
	}

	err = p.Context.CopyDirectory(mixinDir, filepath.Join(build.LOCAL_APP, "mixins"), true)
	return errors.Wrapf(err, "could not copy mixin directory contents for %s", mixin)
}

func (p *Porter) buildInvocationImage(ctx context.Context) error {
	fmt.Fprintf(p.Out, "\nStarting Invocation Image Build =======> \n")
	path, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "could not get current working directory")
	}
	buildOptions := types.ImageBuildOptions{
		SuppressOutput: false,
		PullParent:     false,
		Tags:           []string{p.Config.Manifest.Image},
		Dockerfile:     "Dockerfile",
		BuildArgs: map[string]*string{
			"BUNDLE_DIR": &BUNDLE_DIR,
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
	if p.IsVerbose() {
		dockerOutput = p.Out
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

func (p *Porter) buildBundle(invocationImage string, digest string) error {
	converter := configadapter.ManifestConverter{
		Manifest: p.Manifest,
		Context:  p.Context,
	}
	bun := converter.ToBundle()
	return p.writeBundle(bun)
}

func (p Porter) writeBundle(b *bundle.Bundle) error {
	f, err := p.Config.FileSystem.OpenFile(build.LOCAL_BUNDLE, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	defer f.Close()
	if err != nil {
		return errors.Wrapf(err, "error creating %s", build.LOCAL_BUNDLE)
	}
	_, err = b.WriteTo(f)
	return errors.Wrapf(err, "error writing to %s", build.LOCAL_BUNDLE)
}
