package claims

import (
	"testing"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/claim"
	"github.com/stretchr/testify/require"
)

var _ claim.Provider = TestClaimProvider{}

type TestClaimProvider struct {
	claim.Store
	t *testing.T
}

func NewTestClaimProvider(t *testing.T) TestClaimProvider {
	return TestClaimProvider{
		t:     t,
		Store: claim.NewMockStore(nil, nil),
	}
}

// CreateClaim creates a new claim and saves it.
func (p TestClaimProvider) CreateClaim(installation string, action string, bun bundle.Bundle, parameters map[string]interface{}) claim.Claim {
	c, err := claim.New(installation, action, bun, parameters)
	require.NoError(p.t, err, "New claim failed")
	err = p.SaveClaim(c)
	require.NoError(p.t, err, "SaveClaim failed")
	return c
}

// CreateResult creates a new result from the specified claim and saves it.
func (p TestClaimProvider) CreateResult(c claim.Claim, status string) claim.Result {
	r, err := c.NewResult(status)
	require.NoError(p.t, err, "NewResult failed")
	err = p.SaveResult(r)
	require.NoError(p.t, err, "SaveResult failed")
	return r
}

// CreateOutput creates a new output from the specified claim and result and saves it.
func (p TestClaimProvider) CreateOutput(c claim.Claim, r claim.Result, name string, value []byte) claim.Output {
	o := claim.NewOutput(c, r, name, value)
	err := p.SaveOutput(o)
	require.NoError(p.t, err, "SaveOutput failed")
	return o
}
