package cosign

import (
	"context"
	"fmt"
	"os/exec"

	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/signing/plugins"
	"get.porter.sh/porter/pkg/tracing"
)

var _ plugins.SigningProtocol = &Cosign{}

// Signer implements an in-memory signer for testing.
type Cosign struct {
	PublicKey  string
	PrivateKey string
}

func NewSigner(c *portercontext.Context, cfg PluginConfig) *Cosign {

	s := &Cosign{
		PublicKey:  cfg.PublicKey,
		PrivateKey: cfg.PrivateKey,
	}

	return s
}

// we should get the certificate... here?
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
	cmd := exec.Command("cosign", "sign", ref, "--tlog-upload=false", "--key", s.PrivateKey)
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

	log.Infof("Mock Signer is Verifying %s", ref)
	cmd := exec.Command("cosign", "verify", "--key", s.PublicKey, ref, "--insecure-ignore-tlog")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w", string(out), err)
	}
	log.Infof("%s", out)
	return nil
}
