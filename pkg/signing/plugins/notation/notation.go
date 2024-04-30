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
	SigningKey       string
	InsecureRegistry bool
}

func NewSigner(c *portercontext.Context, cfg PluginConfig) *Signer {
	s := &Signer{
		SigningKey:       cfg.SigningKey,
		InsecureRegistry: cfg.InsecureRegistry,
	}
	return s
}

func (s *Signer) Connect(ctx context.Context) error {
	//lint:ignore SA4006 ignore unused ctx for now
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	log.Debug("Running notation signer")

	return nil
}

func (s *Signer) Sign(ctx context.Context, ref string) error {
	//lint:ignore SA4006 ignore unused ctx for now
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	args := []string{"sign", ref, "--key", s.SigningKey}
	if s.InsecureRegistry {
		args = append(args, "--insecure-registry")
	}
	cmd := exec.Command("notation", args...)
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

	args := []string{"verify", ref}
	if s.InsecureRegistry {
		args = append(args, "--insecure-registry")
	}
	cmd := exec.Command("notation", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w", string(out), err)
	}
	log.Infof("%s", out)
	return nil
}
