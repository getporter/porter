package storage

import (
	inmemory "get.porter.sh/porter/pkg/instance-storage/in-memory"
	"github.com/cnabio/cnab-go/claim"
)

var _ StorageProvider = &TestStorageProvider{}

type TestStorageProvider struct {
	ClaimStore
}

func NewTestStorageProvider() TestStorageProvider {
	crud := inmemory.NewStore()
	return TestStorageProvider{
		ClaimStore: claim.NewClaimStore(crud),
	}
}
