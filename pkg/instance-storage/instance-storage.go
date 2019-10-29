package instancestorage

import (
	"github.com/deislabs/cnab-go/claim"
)

// Provider interface for instance storage (claims).
type Provider interface {
	List() ([]string, error)
	Store(claim.Claim) error
	Read(name string) (claim.Claim, error)
	ReadAll() ([]claim.Claim, error)
	Delete(name string) error
}
