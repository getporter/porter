package storage

import (
	"github.com/cnabio/cnab-go/claim"
)

// StorageProvider interface for instance storage (claims).
type StorageProvider interface {
	List() ([]string, error)
	Store(claim.Claim) error
	Read(name string) (claim.Claim, error)
	ReadAll() ([]claim.Claim, error)
	Delete(name string) error
}
