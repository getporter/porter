package syft

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/sbom/plugins"
	"get.porter.sh/porter/pkg/tracing"
)

var _ plugins.SBOMGeneratorProtocol = &Syft{}

// SBOMGenerator implements an in-memory sbomGenerator for testing.
type Syft struct {
}

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

func (s *Syft) Generate(ctx context.Context, bundleRef string, sbomPath string, insecureRegistry bool) (err error) {
	_, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	log.Infof("Generating SBOM for bundle %s...", bundleRef)
	args := []string{"-o", fmt.Sprintf("spdx-json=%s", sbomPath), bundleRef}
	cmd := exec.Command("syft", args...)
	cmd.Env = append(cmd.Env, os.Environ()...)
	if insecureRegistry {
		log.Info("Setting Syft environment variable to skip TLS verification for insecure registries")
		cmd.Env = append(cmd.Env, "SYFT_REGISTRY_INSECURE_SKIP_TLS_VERIFY=true")
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w", string(out), err)
	}
	log.Infof("%s", out)

	log.Infof("SBOM for bundle %s written to %s", bundleRef, sbomPath)
	return err
}
