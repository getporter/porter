package instancestorage

import (
	inmemory "get.porter.sh/porter/pkg/instance-storage/in-memory"
	"github.com/cnabio/cnab-go/claim"
)

var _ StorageProvider = &TestInstanceStorageProvider{}

type TestInstanceStorageProvider struct {
	ClaimStore
}

func NewTestInstanceStorageProvider() TestInstanceStorageProvider {
	crud := inmemory.NewStore()
	return TestInstanceStorageProvider{
		ClaimStore: claim.NewClaimStore(crud),
	}
}
