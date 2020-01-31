package claims

import (
	"github.com/cnabio/cnab-go/claim"
)

// ClaimProvider interface for claim storage.
type ClaimProvider interface {
	List() ([]string, error)
	Save(claim.Claim) error
	Read(name string) (claim.Claim, error)
	ReadAll() ([]claim.Claim, error)
	Delete(name string) error
}
