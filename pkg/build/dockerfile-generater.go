package build

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/mixin/query"
	"get.porter.sh/porter/pkg/pkgmgmt"
	"get.porter.sh/porter/pkg/templates"
	"github.com/pkg/errors"
)

type DockerfileGenerator struct {
	*config.Config
	*manifest.Manifest
	*templates.Templates
	Mixins pkgmgmt.PackageManager
}

func NewDockerfileGenerator(config *config.Config, m *manifest.Manifest, tmpl *templates.Templates, mp pkgmgmt.PackageManager) *DockerfileGenerator {
	return &DockerfileGenerator{
		Config:    config,
		Manifest:  m,
		Templates: tmpl,
		Mixins:    mp,
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

	mixinsTokenIndex := g.getIndexOfPorterMixinsToken(lines)
	if mixinsTokenIndex == -1 {
		lines = append(lines, mixinLines...)
	} else {
		pretoken := make([]string, mixinsTokenIndex)
		copy(pretoken, lines)
		posttoken := lines[mixinsTokenIndex+1:]
		lines = append(pretoken, append(mixinLines, posttoken...)...)
	}

	lines = append(lines, g.buildPorterSection()...)
	lines = append(lines, g.buildCNABSection()...)
	lines = append(lines, g.buildWORKDIRSection())
	lines = append(lines, g.buildCMDSection())

	if g.IsVerbose() {
		for _, line := range lines {
			fmt.Fprintln(g.Out, line)
		}
	}

	return lines, nil
}

// ErrorMessage to be displayed when no ARG BUNDLE_DIR is in Dockerfile
const ErrorMessage = `
Dockerfile.tmpl must declare the build argument BUNDLE_DIR.
Add the following line to the file and re-run porter build: ARG BUNDLE_DIR`

func (g *DockerfileGenerator) readAndValidateDockerfile(s *bufio.Scanner) ([]string, error) {
	hasBuildArg := false
	buildArg := "ARG BUNDLE_DIR"
	var lines []string
	for s.Scan() {
		if strings.TrimSpace(s.Text()) == buildArg {
			hasBuildArg = true
		}
		lines = append(lines, s.Text())
	}

	if !hasBuildArg {
		return nil, errors.New(ErrorMessage)
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
	scanner := bufio.NewScanner(reader)
	lines, e := g.readAndValidateDockerfile(scanner)
	if e != nil {
		return nil, e
	}
	return lines, nil
}

func (g *DockerfileGenerator) buildPorterSection() []string {
	return []string{
		// Remove the user-provided Porter manifest as the canonical version
		// will migrate via its location in .cnab
		`RUN rm $BUNDLE_DIR/porter.yaml`,
	}
}

func (g *DockerfileGenerator) buildCNABSection() []string {
	return []string{
		// Putting RUN before COPY here as a workaround for https://github.com/moby/moby/issues/37965, back to back COPY statements in the same directory (e.g. /cnab) _may_ result in an error from Docker depending on unpredictable factors
		`RUN rm -fr $BUNDLE_DIR/.cnab`,
		`COPY .cnab /cnab`,
	}
}

func (g *DockerfileGenerator) buildWORKDIRSection() string {
	return `WORKDIR $BUNDLE_DIR`
}

func (g *DockerfileGenerator) buildCMDSection() string {
	return `CMD ["/cnab/app/run"]`
}

func (g *DockerfileGenerator) buildMixinsSection() ([]string, error) {
	q := query.New(g.Context, g.Mixins)
	q.RequireAllMixinResponses = true
	q.LogMixinErrors = true
	results, err := q.Execute("build", query.NewManifestGenerator(g.Manifest))
	if err != nil {
		return nil, err
	}

	lines := make([]string, 0)
	for _, result := range results {
		l := strings.Split(result, "\n")
		lines = append(lines, l...)
	}
	return lines, nil
}

func (g *DockerfileGenerator) PrepareFilesystem() error {
	// clean up previously generated files
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

	err = g.Context.CopyDirectory(
		filepath.Join(g.GetHomeDir(), "runtimes"),
		filepath.Join(LOCAL_APP, "runtimes"), false)
	if err != nil {
		return err
	}

	fmt.Fprintf(g.Out, "Copying mixins ===> \n")
	for _, m := range g.Manifest.Mixins {
		err := g.copyMixin(m.Name)
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *DockerfileGenerator) copyMixin(mixin string) error {
	fmt.Fprintf(g.Out, "Copying mixin %s ===> \n", mixin)
	mixinDir, err := g.Mixins.GetPackageDir(mixin)
	if err != nil {
		return err
	}

	err = g.Context.CopyDirectory(mixinDir, filepath.Join(LOCAL_APP, "mixins"), true)
	return errors.Wrapf(err, "could not copy mixin directory contents for %s", mixin)
}

func (g *DockerfileGenerator) getIndexOfPorterMixinsToken(a []string) int {
	for i, n := range a {
		if INJECT_PORTER_MIXINS_TOKEN == strings.TrimSpace(n) {
			return i
		}
	}
	return -1
}
