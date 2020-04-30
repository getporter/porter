package claims

import (
	"encoding/json"
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
	t.Skip("TODO: Migrate Claim Data #937")

	const installation = "example-exec-outputs"
	testcases := []struct {
		name          string
		fileName      string
		shouldMigrate bool
	}{
		{name: "new claims do not migrate", fileName: "newschema", shouldMigrate: false},
		{name: "unmigrated claim migrates", fileName: "unmigrated", shouldMigrate: true},
		{name: "migrated claims do not migrated", fileName: "migrated", shouldMigrate: false},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cxt := context.NewTestContext(t)
			dataStore := crud.NewMockStore()
			wrapper := newMigrateClaimsWrapper(cxt.Context, dataStore)
			claimStore := claim.NewClaimStore(wrapper, nil, nil)

			loadTestClaim(t, tc.fileName, claimStore)

			c, err := claimStore.ReadLastClaim(installation)
			require.NoError(t, err, "could not read claim")
			require.NotNil(t, c, "claim should be populated")
			assert.Equal(t, "example-exec-outputs", c.Installation, "claim.Installation was not populated")

			if tc.shouldMigrate {
				assert.Contains(t, cxt.GetError(), "Migrating installation", "the claim should have been migrated")
			} else {
				assert.NotContains(t, cxt.GetError(), "Migrating installation", "the claim should NOT be migrated")
			}
		})
	}
}

func TestMigrateClaimsWrapper_List(t *testing.T) {
	cxt := context.NewTestContext(t)
	dataStore := crud.NewMockStore()
	wrapper := newMigrateClaimsWrapper(cxt.Context, dataStore)
	claimStore := claim.NewClaimStore(wrapper, nil, nil)

	loadTestClaim(t, "newschema", claimStore)
	loadTestClaim(t, "migrated", claimStore)

	names, err := claimStore.ListInstallations()
	sort.Strings(names)
	require.NoError(t, err, "could not list installations")
	assert.Equal(t, []string{"example-exec-outputs"}, names, "unexpected list of installation names")
}

func loadTestClaim(t *testing.T, filename string, store claim.Store) {
	testfile := fmt.Sprintf("testdata/%s.json", filename)
	claimB, err := ioutil.ReadFile(testfile)
	require.NoError(t, err, "could not read %s", testfile)

	var c claim.Claim
	err = json.Unmarshal(claimB, &c)
	require.NoError(t, err, "could not unmarshal %s", testfile)

	err = store.SaveClaim(c)
	require.NoError(t, err, "could not save testdata %s into mock store", testfile)
}
