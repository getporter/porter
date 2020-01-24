package claims

import (
	inmemory "get.porter.sh/porter/pkg/storage/in-memory"
	"github.com/cnabio/cnab-go/claim"
)

var _ ClaimProvider = &TestClaimProvider{}

type TestClaimProvider struct {
	claim.Store
}

func NewTestClaimProvider() TestClaimProvider {
	crud := inmemory.NewStore()
	return TestClaimProvider{
		Store: claim.NewClaimStore(crud),
	}
}
