package build

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/mixin/query"
	"get.porter.sh/porter/pkg/pkgmgmt"
	"get.porter.sh/porter/pkg/templates"
	"get.porter.sh/porter/pkg/tracing"
)

const (
	// DefaultDockerfileSyntax is the default syntax for Dockerfiles used by Porter
	// either when generating a Dockerfile from scratch, or when a template does
	// not define a syntax
	DefaultDockerfileSyntax = "docker/dockerfile-upstream:1.4.0"
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

func (g *DockerfileGenerator) GenerateDockerFile(ctx context.Context) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	lines, err := g.buildDockerfile(ctx)
	if err != nil {
		return span.Error(fmt.Errorf("error generating the Dockerfile: %w", err))
	}

	contents := strings.Join(lines, "\n")

	// Output the generated dockerfile
	span.Debug(contents)

	err = g.FileSystem.WriteFile(DOCKER_FILE, []byte(contents), pkg.FileModeWritable)
	if err != nil {
		return span.Error(fmt.Errorf("couldn't write the Dockerfile: %w", err))
	}

	return nil
}

func (g *DockerfileGenerator) buildDockerfile(ctx context.Context) ([]string, error) {
	log := tracing.LoggerFromContext(ctx)
	log.Debug("Generating Dockerfile")

	lines, err := g.getBaseDockerfile(ctx)
	if err != nil {
		return nil, err
	}

	lines = append(lines, g.buildPorterSection()...)
	lines = append(lines, g.buildCNABSection()...)
	lines = append(lines, g.buildWORKDIRSection())
	lines = append(lines, g.buildCMDSection())

	return lines, nil
}

func (g *DockerfileGenerator) getBaseDockerfile(ctx context.Context) ([]string, error) {
	log := tracing.LoggerFromContext(ctx)

	var reader io.Reader
	if g.Manifest.Dockerfile != "" {
		exists, err := g.FileSystem.Exists(g.Manifest.Dockerfile)
		if err != nil {
			return nil, fmt.Errorf("error checking if Dockerfile exists: %q: %w", g.Manifest.Dockerfile, err)
		}
		if !exists {
			return nil, fmt.Errorf("the Dockerfile specified in the manifest doesn't exist: %q", g.Manifest.Dockerfile)
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
			return nil, fmt.Errorf("error loading default Dockerfile template: %w", err)
		}
		reader = bytes.NewReader(contents)
	}
	scanner := bufio.NewScanner(reader)
	var lines []string
	var syntaxFound bool
	for scanner.Scan() {
		line := scanner.Text()
		if !syntaxFound && strings.HasPrefix(line, "# syntax=") {
			syntaxFound = true
		}
		lines = append(lines, line)
	}

	// If their template doesn't declare a syntax, use the default
	if !syntaxFound {
		log.Warnf("No syntax was declared in the template Dockerfile, using %s", DefaultDockerfileSyntax)
		lines = append([]string{fmt.Sprintf("# syntax=%s", DefaultDockerfileSyntax)}, lines...)
	}

	return g.replaceTokens(ctx, lines)
}

func (g *DockerfileGenerator) buildPorterSection() []string {
	// The user-provided manifest may be located separate from the build context directory.
	// Therefore, we only need to add lines if the relative manifest path exists inside of
	// the current working directory.
	manifestPath := g.FileSystem.Abs(g.Manifest.ManifestPath)
	if relManifestPath, err := filepath.Rel(g.Getwd(), manifestPath); err == nil {
		if !strings.Contains(relManifestPath, "..") {
			return []string{
				// Remove the user-provided Porter manifest as the canonical version
				// will migrate via its location in .cnab
				fmt.Sprintf(`RUN rm ${BUNDLE_DIR}/%s`, relManifestPath),
			}
		}
	}
	return []string{}
}

func (g *DockerfileGenerator) buildCNABSection() []string {
	copyCNAB := "COPY .cnab /cnab"
	if g.GetBuildDriver() == config.BuildDriverBuildkit {
		copyCNAB = "COPY --link .cnab /cnab"
	}

	return []string{
		// Putting RUN before COPY here as a workaround for https://github.com/moby/moby/issues/37965, back to back COPY statements in the same directory (e.g. /cnab) _may_ result in an error from Docker depending on unpredictable factors
		`RUN rm -fr ${BUNDLE_DIR}/.cnab`,
		// Copy the non-user cnab files, like mixins and porter.yaml, from the local .cnab directory into the bundle
		copyCNAB,
		// Ensure that regardless of the container's UID, the root group (default group for arbitrary users that do not exist in the container) has the same permissions as the owner
		// See https://developers.redhat.com/blog/2020/10/26/adapting-docker-and-kubernetes-containers-to-run-on-red-hat-openshift-container-platform#group_ownership_and_file_permission
		`RUN chgrp -R ${BUNDLE_GID} /cnab && chmod -R g=u /cnab`,
		// default to running as the nonroot user that the porter agent uses.
		// When running in kubernetes, if you specify a different UID, make sure to set fsGroup to the same UID, and runasGroup to 0
		`USER ${BUNDLE_UID}`,
	}
}

func (g *DockerfileGenerator) buildWORKDIRSection() string {
	return `WORKDIR ${BUNDLE_DIR}`
}

func (g *DockerfileGenerator) buildCMDSection() string {
	return `CMD ["/cnab/app/run"]`
}

func (g *DockerfileGenerator) buildMixinsSection(ctx context.Context) ([]string, error) {
	q := query.New(g.Context, g.Mixins)
	q.RequireAllMixinResponses = true
	q.LogMixinErrors = true
	results, err := q.Execute(ctx, "build", query.NewManifestGenerator(g.Manifest))
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

func (g *DockerfileGenerator) buildInitSection() []string {
	return []string{
		"ARG BUNDLE_DIR",
		"ARG BUNDLE_UID=65532",
		"ARG BUNDLE_USER=nonroot",
		"ARG BUNDLE_GID=0",
		// Create a non-root user that is in the root group with the specified id and a home directory
		"RUN useradd ${BUNDLE_USER} -m -u ${BUNDLE_UID} -g ${BUNDLE_GID} -o",
	}
}

func (g *DockerfileGenerator) PrepareFilesystem() error {
	fmt.Fprintf(g.Out, "Copying porter runtime ===> \n")

	runTmpl, err := g.Templates.GetRunScript()
	if err != nil {
		return err
	}

	err = g.FileSystem.MkdirAll(LOCAL_APP, pkg.FileModeDirectory)
	if err != nil {
		return err
	}

	err = g.FileSystem.WriteFile(LOCAL_RUN, runTmpl, pkg.FileModeExecutable)
	if err != nil {
		return fmt.Errorf("failed to write %s: %w", LOCAL_RUN, err)
	}

	homeDir, err := g.GetHomeDir()
	if err != nil {
		return err
	}
	err = g.Context.CopyDirectory(filepath.Join(homeDir, "runtimes"), filepath.Join(LOCAL_APP, "runtimes"), false)
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

	err = g.Context.CopyDirectory(mixinDir, LOCAL_MIXINS, true)
	if err != nil {
		return fmt.Errorf("could not copy mixin directory contents for %s: %w", mixin, err)
	}

	return nil
}

func (g *DockerfileGenerator) getIndexOfToken(lines []string, token string) int {
	for lineNumber, lineContent := range lines {
		if token == strings.TrimSpace(lineContent) {
			return lineNumber
		}
	}
	return -1
}

// replaceTokens looks for lines like # PORTER_MIXINS and replaces them in the
// template with the appropriate set of Dockerfile lines.
func (g *DockerfileGenerator) replaceTokens(ctx context.Context, lines []string) ([]string, error) {
	mixinLines, err := g.buildMixinsSection(ctx)
	if err != nil {
		return nil, fmt.Errorf("error generating Dockerfile content for mixins: %w", err)
	}

	fromToken := g.getIndexOfToken(lines, "FROM")

	substitutions := []struct {
		token        string
		lines        []string
		defaultIndex int
		replace      bool
	}{
		{token: PORTER_INIT_TOKEN, lines: g.buildInitSection(), defaultIndex: fromToken, replace: true},
		{token: PORTER_MIXINS_TOKEN, lines: mixinLines, defaultIndex: -1, replace: true},
	}

	for _, substitution := range substitutions {
		index := g.getIndexOfToken(lines, substitution.token)

		// If we can't find the token, use the default for that token
		if index == -1 {
			index = substitution.defaultIndex
			substitution.replace = false
		}

		if index == -1 {
			lines = append(lines, substitution.lines...)
		} else {
			prefix := make([]string, index)
			copy(prefix, lines)

			if substitution.replace {
				// Do not keep the line at the insertion index, replace it with the new lines instead
				index = index + 1
			}

			suffix := lines[index:]
			lines = append(prefix, append(substitution.lines, suffix...)...)
		}
	}

	return lines, nil
}
