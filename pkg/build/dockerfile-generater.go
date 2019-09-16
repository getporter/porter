package build

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/deislabs/porter/pkg/config"
	"github.com/deislabs/porter/pkg/context"
	"github.com/deislabs/porter/pkg/mixin"
	"github.com/deislabs/porter/pkg/templates"
	"github.com/pkg/errors"
)

type DockerfileGenerator struct {
	*config.Config
	*templates.Templates
	mixin.MixinProvider
}

func NewDockerfileGenerator(cfg *config.Config, tmpl *templates.Templates, mp mixin.MixinProvider) *DockerfileGenerator {
	return &DockerfileGenerator{
		Config:        cfg,
		Templates:     tmpl,
		MixinProvider: mp,
	}
}

func (g *DockerfileGenerator) GenerateDockerFile() error {
	lines, err := g.buildDockerfile()
	if err != nil {
		return errors.Wrap(err, "error generating the Dockerfile")
	}

	fmt.Fprintf(g.Out, "\nWriting Dockerfile =======>\n")
	contents := strings.Join(lines, "\n")

	if g.IsVerbose() {
		fmt.Fprintln(g.Out, contents)
	}

	err = g.FileSystem.WriteFile("Dockerfile", []byte(contents), 0644)
	return errors.Wrap(err, "couldn't write the Dockerfile")
}

func (g *DockerfileGenerator) buildDockerfile() ([]string, error) {
	fmt.Fprintf(g.Out, "\nGenerating Dockerfile =======>\n")

	lines, err := g.getBaseDockerfile()
	if err != nil {
		return nil, err
	}

	mixinLines, err := g.buildMixinsSection()
	if err != nil {
		return nil, errors.Wrap(err, "error generating Dockefile content for mixins")
	}
	lines = append(lines, mixinLines...)

	// The template dockerfile copies everything by default, but if the user
	// supplied their own, copy over cnab/ and porter.yaml
	if g.Manifest.Dockerfile != "" {
		lines = append(lines, g.buildCNABSection()...)
		lines = append(lines, g.buildPorterSection()...)
	}
	lines = append(lines, g.buildWORKDIRSection())
	lines = append(lines, g.buildCMDSection())

	if g.IsVerbose() {
		for _, line := range lines {
			fmt.Fprintln(g.Out, line)
		}
	}

	return lines, nil
}

func (g *DockerfileGenerator) getBaseDockerfile() ([]string, error) {
	var reader io.Reader
	if g.Manifest.Dockerfile != "" {
		exists, err := g.FileSystem.Exists(g.Manifest.Dockerfile)
		if err != nil {
			return nil, errors.Wrapf(err, "error checking if Dockerfile exists: %q", g.Manifest.Dockerfile)
		}
		if !exists {
			return nil, errors.Errorf("the Dockerfile specified in the manifest doesn't exist: %q", g.Manifest.Dockerfile)
		}

		file, err := g.FileSystem.Open(g.Manifest.Dockerfile)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		reader = file

	} else {
		contents, err := g.Templates.GetDockerfile()
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

func (g *DockerfileGenerator) buildPorterSection() []string {
	return []string{
		`COPY porter.yaml $BUNDLE_DIR/porter.yaml`,
	}
}

func (g *DockerfileGenerator) buildCNABSection() []string {
	return []string{
		`COPY .cnab/ /cnab/`,
	}
}

func (g *DockerfileGenerator) buildWORKDIRSection() string {
	return `WORKDIR $BUNDLE_DIR`
}

func (g *DockerfileGenerator) buildCMDSection() string {
	return `CMD ["/cnab/app/run"]`
}

func (g *DockerfileGenerator) buildMixinsSection() ([]string, error) {
	lines := make([]string, 0)
	for _, m := range g.Manifest.Mixins {
		// Copy the existing context and tweak to pipe the output differently
		mixinStdout := &bytes.Buffer{}
		var mixinContext context.Context
		mixinContext = *g.Context
		mixinContext.Out = mixinStdout   // mixin stdout -> dockerfile lines
		mixinContext.Err = g.Context.Out // mixin stderr -> logs
		mixinContext.In = nil            // TODO: let the mixin know about which steps will be executed so that it can be more selective about copying into the invocation image

		cmd := mixin.CommandOptions{Command: "build"}
		err := g.MixinProvider.Run(&mixinContext, m, cmd)
		if err != nil {
			return nil, err
		}

		l := strings.Split(mixinStdout.String(), "\n")
		lines = append(lines, l...)
	}
	return lines, nil
}

func (g *DockerfileGenerator) getMixinBuildInput(m string) mixin.BuildInput {
	input := mixin.BuildInput{
		Actions: make(map[string]config.Steps, 3),
	}

	// TODO: populate config once we are storing it in porter.yaml

	filterSteps := func(action config.Action, steps config.Steps) {
		mixinSteps := config.Steps{}
		for _, step := range steps {
			if step.GetMixinName() != m {
				continue
			}
			mixinSteps = append(mixinSteps, step)
		}
		input.Actions[string(action)] = mixinSteps
	}
	filterSteps(config.ActionInstall, g.Manifest.Install)
	filterSteps(config.ActionUpgrade, g.Manifest.Upgrade)
	filterSteps(config.ActionUninstall, g.Manifest.Uninstall)

	for action, steps := range g.Manifest.CustomActions {
		filterSteps(config.Action(action), steps)
	}

	return input
}

func (g *DockerfileGenerator) PrepareFilesystem() error {
	// clean up previously generated files
	g.FileSystem.RemoveAll(LOCAL_CNAB)
	g.FileSystem.Remove("Dockerfile")

	fmt.Fprintf(g.Out, "Copying porter runtime ===> \n")

	runTmpl, err := g.Templates.GetRunScript()
	if err != nil {
		return err
	}

	err = g.FileSystem.MkdirAll(LOCAL_APP, 0755)
	if err != nil {
		return err
	}

	err = g.FileSystem.WriteFile(LOCAL_RUN, runTmpl, 0755)
	if err != nil {
		return errors.Wrapf(err, "failed to write %s", LOCAL_RUN)
	}

	pr, err := g.GetPorterRuntimePath()
	if err != nil {
		return err
	}
	err = g.CopyFile(pr, filepath.Join(LOCAL_APP, "porter-runtime"))
	if err != nil {
		return err
	}

	fmt.Fprintf(g.Out, "Copying mixins ===> \n")
	for _, mixin := range g.Manifest.Mixins {
		err := g.copyMixin(mixin)
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *DockerfileGenerator) copyMixin(mixin string) error {
	fmt.Fprintf(g.Out, "Copying mixin %s ===> \n", mixin)
	mixinDir, err := g.GetMixinDir(mixin)
	if err != nil {
		return err
	}

	err = g.Context.CopyDirectory(mixinDir, filepath.Join(LOCAL_APP, "mixins"), true)
	return errors.Wrapf(err, "could not copy mixin directory contents for %s", mixin)
}
