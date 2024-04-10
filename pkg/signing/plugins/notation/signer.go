package notation

import (
	"context"
	"fmt"
	"os/exec"

	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/signing/plugins"
	"get.porter.sh/porter/pkg/tracing"
)

var _ plugins.SigningProtocol = &Signer{}

// Signer implements an in-memory signer for testing.
type Signer struct {

	// Need the key we want to use
	SigningKey string
}

func NewSigner(c *portercontext.Context, cfg PluginConfig) *Signer {
	s := &Signer{
		SigningKey: cfg.SigningKey,
	}
	return s
}

func (s *Signer) Connect(ctx context.Context) error {
	//lint:ignore SA4006 ignore unused ctx for now
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	log.Debug("Running mock signer")

	return nil
}

// Close implements the Close method on the signing plugins' interface.
func (s *Signer) Close() error {
	return nil
}

func (s *Signer) Sign(ctx context.Context, ref string) error {
	//lint:ignore SA4006 ignore unused ctx for now
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	cmd := exec.Command("notation", "sign", ref, "--key", s.SigningKey)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w", string(out), err)
	}
	log.Infof("%s", out)
	return nil
}

func (s *Signer) Verify(ctx context.Context, ref string) error {
	//lint:ignore SA4006 ignore unused ctx for now
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	cmd := exec.Command("notation", "verify", ref)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w", string(out), err)
	}
	log.Infof("%s", out)
	return nil
}
