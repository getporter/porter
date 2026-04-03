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

func (s *Syft) Generate(
	ctx context.Context,
	bundleRef string,
	sbomPath string,
	insecureRegistry bool,
) (err error) {
	_, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	log.Infof("Generating SBOM for bundle %s...", bundleRef)

	// Pre-pull the image into the local Docker daemon so syft can find it immediately.
	// Without this, syft's first attempt pulls the image as a side effect but fails
	// because the image isn't available in the daemon until after that pull completes.
	pullCmd := exec.Command("docker", "pull", bundleRef)
	if insecureRegistry {
		pullCmd.Env = append(os.Environ(), "DOCKER_TLS_VERIFY=0")
	}
	if pullOut, err := pullCmd.CombinedOutput(); err != nil {
		log.Warnf("Failed to pre-pull image %s before SBOM generation: %s: %v", bundleRef, string(pullOut), err)
	}

	args := []string{"-o", fmt.Sprintf("spdx-json=%s", sbomPath), bundleRef}
	cmd := exec.Command("syft", args...)
	if insecureRegistry {
		log.Info(
			"Setting Syft environment variable to skip TLS verification for insecure registries",
		)
		cmd.Env = append(os.Environ(), "SYFT_REGISTRY_INSECURE_SKIP_TLS_VERIFY=true")
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w", string(out), err)
	}
	log.Infof("%s", out)

	log.Infof("SBOM for bundle %s written to %s", bundleRef, sbomPath)
	return err
}
