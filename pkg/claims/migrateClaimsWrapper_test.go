package claims

import (
	"fmt"
	"io/ioutil"
	"sort"
	"testing"

	"get.porter.sh/porter/pkg/context"
	"github.com/cnabio/cnab-go/claim"
	"github.com/cnabio/cnab-go/utils/crud"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrateClaimsWrapper_Read(t *testing.T) {
	testcases := []struct {
		name          string
		claimName     string
		shouldMigrate bool
	}{
		{name: "new claims do not migrate", claimName: "newschema", shouldMigrate: false},
		{name: "unmigrated claim migrates", claimName: "unmigrated", shouldMigrate: true},
		{name: "migrated claims do not migrated", claimName: "migrated", shouldMigrate: false},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cxt := context.NewTestContext(t)
			dataStore := crud.NewMockStore()
			wrapper := newMigrateClaimsWrapper(cxt.Context, dataStore)
			claimStore := claim.NewClaimStore(wrapper)

			loadTestClaim(t, tc.claimName, dataStore)

			c, err := claimStore.Read(tc.claimName)
			require.NoError(t, err, "could not read claim")
			require.NotNil(t, c, "claim should be populated")
			assert.Equal(t, "example-exec-outputs", c.Installation, "claim.Installation was not populated")

			if tc.shouldMigrate {
				assert.Contains(t, cxt.GetError(), "Migrating bundle instance", "the claim should have been migrated")
			} else {
				assert.NotContains(t, cxt.GetError(), "Migrating bundle instance", "the claim should NOT be migrated")
			}
		})
	}
}

func TestMigrateClaimsWrapper_List(t *testing.T) {
	cxt := context.NewTestContext(t)
	dataStore := crud.NewMockStore()
	wrapper := newMigrateClaimsWrapper(cxt.Context, dataStore)
	claimStore := claim.NewClaimStore(wrapper)

	loadTestClaim(t, "unmigrated", dataStore)
	loadTestClaim(t, "migrated", dataStore)

	names, err := claimStore.List()
	sort.Strings(names)
	require.NoError(t, err, "could not list claims")
	assert.Equal(t, []string{"migrated", "unmigrated"}, names, "unexpected list of claim installation names")
}

func loadTestClaim(t *testing.T, claimName string, store crud.Store) {
	testfile := fmt.Sprintf("testdata/%s.json", claimName)
	claimB, err := ioutil.ReadFile(testfile)
	require.NoError(t, err, "could not read %s", testfile)
	err = store.Save(claim.ItemType, claimName, claimB)
	require.NoError(t, err, "could not save testdata %s into mock store", testfile)
}
