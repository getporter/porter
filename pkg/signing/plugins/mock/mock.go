package mock

import (
	"context"
	b64 "encoding/base64"

	"get.porter.sh/porter/pkg/signing/plugins"
	"get.porter.sh/porter/pkg/tracing"
)

var _ plugins.SigningProtocol = &Signer{}

// Signer implements an in-memory signer for testing.
type Signer struct {
	Signatures map[string]string
}

func NewSigner() *Signer {
	s := &Signer{
		Signatures: make(map[string]string),
	}

	return s
}

func (s *Signer) Connect(ctx context.Context) error {
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

	log.Infof("Mock Signer is Signing %s", ref)
	s.Signatures[ref] = b64.StdEncoding.EncodeToString([]byte(ref))
	return nil
}

func (s *Signer) Verify(ctx context.Context, ref string) error {
	//lint:ignore SA4006 ignore unused ctx for now
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	log.Infof("Mock Signer is Verifying %s", ref)
	return nil
}
