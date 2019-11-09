package instancestorage

import (
	"github.com/deislabs/cnab-go/claim"
	inmemory "github.com/deislabs/porter/pkg/instance-storage/in-memory"
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
