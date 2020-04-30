package claims

import (
	"github.com/cnabio/cnab-go/claim"
	"github.com/cnabio/cnab-go/utils/crud"
)

var _ claim.Provider = &TestClaimProvider{}

type TestClaimProvider struct {
	claim.Store
}

func NewTestClaimProvider() TestClaimProvider {
	crud := crud.NewMockStore()
	return TestClaimProvider{
		Store: claim.NewClaimStore(crud, nil, nil),
	}
}
