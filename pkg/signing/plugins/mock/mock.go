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
	_, log := tracing.StartSpan(ctx)
	defer log.EndSpan()
	return nil
}

func (s *Signer) Sign(ctx context.Context, ref string) error {
	_, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	s.Signatures[ref] = b64.StdEncoding.EncodeToString([]byte(ref))
	return nil
}

func (s *Signer) Verify(ctx context.Context, ref string) error {
	_, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	if _, ok := s.Signatures[ref]; !ok {
		return log.Errorf("%s is not signed", ref)
	}

	return nil
}
