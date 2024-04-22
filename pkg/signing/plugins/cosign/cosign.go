package cosign

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/signing/plugins"
	"get.porter.sh/porter/pkg/tracing"
)

var _ plugins.SigningProtocol = &Cosign{}

// Signer implements an in-memory signer for testing.
type Cosign struct {
	PublicKey        string
	PrivateKey       string
	RegistryMode     string
	Experimental     bool
	InsecureRegistry bool
}

func NewSigner(c *portercontext.Context, cfg PluginConfig) *Cosign {

	s := &Cosign{
		PublicKey:        cfg.PublicKey,
		PrivateKey:       cfg.PrivateKey,
		RegistryMode:     cfg.RegistryMode,
		Experimental:     cfg.Experimental,
		InsecureRegistry: cfg.InsecureRegistry,
	}

	return s
}

// TODO: we should get the certificate... here?
func (s *Cosign) Connect(ctx context.Context) error {
	//lint:ignore SA4006 ignore unused ctx for now
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	log.Debug("Running cosign signer")

	return nil
}

// Close implements the Close method on the signing plugins' interface.
func (s *Cosign) Close() error {
	return nil
}

func (s *Cosign) Sign(ctx context.Context, ref string) error {
	//lint:ignore SA4006 ignore unused ctx for now
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()
	log.Infof("Cosign Signer is Signing %s", ref)
	args := []string{"sign", ref, "--tlog-upload=false", "--key", s.PrivateKey, "--yes"}
	if s.RegistryMode != "" {
		args = append(args, "--registry-referrers-mode", s.RegistryMode)
	}
	if s.InsecureRegistry {
		args = append(args, "--allow-insecure-registry")
	}
	cmd := exec.Command("cosign", args...)
	cmd.Env = append(cmd.Env, os.Environ()...)
	if s.Experimental {
		cmd.Env = append(cmd.Env, "COSIGN_EXPERIMENTAL=1")
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w", string(out), err)
	}
	log.Infof("%s", out)
	return nil
}

func (s *Cosign) Verify(ctx context.Context, ref string) error {
	//lint:ignore SA4006 ignore unused ctx for now
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	log.Infof("Cosign Signer is Verifying %s", ref)
	args := []string{"verify", "--key", s.PublicKey, ref, "--insecure-ignore-tlog"}
	if s.RegistryMode == "oci-1-1" {
		args = append(args, "--experimental-oci11")
	}
	if s.InsecureRegistry {
		args = append(args, "--allow-insecure-registry")
	}
	cmd := exec.Command("cosign", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w", string(out), err)
	}
	log.Infof("%s", out)
	return nil
}
