package cnabprovider

import (
	"encoding/json"
	"fmt"

	"github.com/deislabs/cnab-go/claim"
	"github.com/deislabs/cnab-go/utils/crud"
	"github.com/deislabs/porter/pkg/config"
	"github.com/pkg/errors"
)

type TestDuffle struct {
	*Duffle
	*config.TestConfig
	ClaimStore TestStore
}

type TestStore struct {
	backingStore map[string][]byte
}

func NewTestDuffle(tc *config.TestConfig) *TestDuffle {
	d := NewDuffle(tc.Config)

	return &TestDuffle{
		Duffle:     d,
		TestConfig: tc,
		ClaimStore: NewTestStore(),
	}
}

func (t *TestDuffle) FetchClaim(name string) (*claim.Claim, error) {
	bytes, err := t.ClaimStore.Read(name)
	if err != nil {
		return nil, errors.Wrapf(err, "could not retrieve claim %s", name)
	}

	var claim claim.Claim
	err = json.Unmarshal(bytes, &claim)
	if err != nil {
		return nil, errors.Wrapf(err, "error encountered unmarshaling claim %s", name)
	}

	return &claim, nil
}

func (t *TestDuffle) NewClaimStore() crud.Store {
	return NewTestStore()
}

func NewTestStore() TestStore {
	return TestStore{
		backingStore: make(map[string][]byte),
	}
}

// The following are the necessary methods for TestStore to implement
// to satisfy the crud.Store interface

func (ts TestStore) List() ([]string, error) {
	claimList := []string{}
	for name := range ts.backingStore {
		claimList = append(claimList, name)
	}
	return claimList, nil
}

func (ts TestStore) Store(name string, bytes []byte) error {
	ts.backingStore[name] = bytes
	return nil
}

func (ts TestStore) Read(name string) ([]byte, error) {
	bytes, exists := ts.backingStore[name]
	if !exists {
		return []byte{}, fmt.Errorf("claim %s not found", name)
	}
	return bytes, nil
}

func (ts TestStore) ReadAll() ([]claim.Claim, error) {
	claims := make([]claim.Claim, 0)

	list, err := ts.List()
	if err != nil {
		return claims, err
	}

	for _, c := range list {
		bytes, err := ts.Read(c)
		if err != nil {
			return claims, err
		}

		var cl claim.Claim
		err = json.Unmarshal(bytes, &cl)
		if err != nil {
			return nil, err
		}
		claims = append(claims, cl)
	}
	return claims, nil
}

func (ts TestStore) Delete(name string) error {
	_, ok := ts.backingStore[name]
	if ok {
		delete(ts.backingStore, name)
	}
	return nil
}
