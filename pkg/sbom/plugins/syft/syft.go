package syft

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/sbom/plugins"
	"get.porter.sh/porter/pkg/tracing"
)

var _ plugins.SBOMGeneratorProtocol = &Syft{}

type Syft struct{}

func NewSBOMGenerator(c *portercontext.Context, cfg PluginConfig) *Syft {
	s := &Syft{}
	return s
}

func (s *Syft) Connect(ctx context.Context) error {
	_, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	if err := exec.Command("syft", "version").Run(); err != nil {
		return errors.New("syft was not found")
	}

	return nil
}

func (s *Syft) runSyft(
	ctx context.Context,
	bundleRef string,
	sbomPath string,
	insecureRegistry bool,
) ([]byte, error) {
	args := []string{"-o", fmt.Sprintf("spdx-json=%s", sbomPath), bundleRef}
	cmd := exec.CommandContext(ctx, "syft", args...)
	if insecureRegistry {
		cmd.Env = append(os.Environ(), "SYFT_REGISTRY_INSECURE_SKIP_TLS_VERIFY=true")
	}

	return cmd.CombinedOutput()
}

func (s *Syft) Generate(
	ctx context.Context,
	bundleRef string,
	sbomPath string,
	insecureRegistry bool,
) (err error) {
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	log.Infof("Generating SBOM for bundle %s...", bundleRef)

	out, err := s.runSyft(ctx, bundleRef, sbomPath, insecureRegistry)

	// there seems to be a bug in syft/docker which means that the first attempt to run the command fails, but subsequent attempts work
	// as a work around, if the first attempt fails with a specific error, we should just attempt to re-run it!
	if err != nil && strings.Contains(string(out), "errors occurred attempting to resolve") {
		log.Warnf(
			"Initial attempt to run syft has failed with a known error, retrying. Error message: %s",
			err.Error(),
		)
		out, err = s.runSyft(ctx, bundleRef, sbomPath, insecureRegistry)
	}
	if err != nil {
		return fmt.Errorf("%s: %w", string(out), err)
	}
	log.Debug(string(out))

	log.Infof("SBOM for bundle %s written to %s", bundleRef, sbomPath)
	return err
}
