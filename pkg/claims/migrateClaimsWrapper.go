package claims

import (
	"encoding/json"
	"fmt"

	"get.porter.sh/porter/pkg/context"
	"github.com/cnabio/cnab-go/utils/crud"
	"github.com/pkg/errors"
)

var _ storeWithConnect = &migrateClaimsWrapper{}

type storeWithConnect interface {
	crud.Store
	crud.HasConnect
}

// Migrate old claims when they are accessed.
type migrateClaimsWrapper struct {
	*context.Context
	storeWithConnect
}

func newMigrateClaimsWrapper(cxt *context.Context, wrappedStore storeWithConnect) *migrateClaimsWrapper {
	return &migrateClaimsWrapper{
		Context:          cxt,
		storeWithConnect: crud.NewBackingStore(wrappedStore),
	}
}

// Store is the data store being wrapped.
func (w *migrateClaimsWrapper) Store() crud.Store {
	return w.storeWithConnect
}

func (w *migrateClaimsWrapper) Read(itemType string, name string) ([]byte, error) {
	// Read is also called during List (ReadAll) so this will migrate everything for a list too
	return w.Migrate(itemType, name)
}

func (w *migrateClaimsWrapper) Migrate(itemType string, name string) ([]byte, error) {
	data, err := w.Store().Read(itemType, name)
	if err != nil {
		return nil, err
	}

	// If we can't migrate the data at any point, print an error message and
	// return the original claim so that the entire operation isn't halted for
	// remaining claims

	var rawData map[string]interface{}
	err = json.Unmarshal(data, &rawData)
	if err != nil {
		// Migration check failed, return original claim
		err = errors.Wrapf(err, "error unmarshaling %s/%s to a map for the claim migration\n%s", itemType, name, string(data))
		fmt.Fprintln(w.Err, err)
		return data, nil
	}

	legacyName, hasLegacyName := rawData["name"]
	_, hasInstallationName := rawData["installation"]

	// Migrate claim.Name to claim.Installation, ignoring claims that have
	// already been migrated
	if hasLegacyName && !hasInstallationName {
		if w.Debug {
			fmt.Fprintf(w.Err, "Migrating bundle instance %s (Name -> Installation) to match the CNAB Claim spec https://cnab.io/schema/cnab-claim-1.0.0-DRAFT+d7ffba8/claim.schema.json. The Name field will be preserved for compatibility with previous versions of the spec.\n", name)
		}
		rawData["installation"] = legacyName

		migratedData, err := json.MarshalIndent(rawData, "", "  ")
		if err != nil {
			// Migration failed, return original claim
			err = errors.Wrapf(err, "error unmarshaling claim %s to a map for the claim migration\n%s", name, string(data))
			fmt.Fprintln(w.Err, err)
			return data, nil
		}

		err = w.Store().Save(itemType, name, migratedData)
		if err != nil {
			// Migration failed, return original claim
			err = errors.Wrapf(err, "error persisting migrated claim %s", name)
			fmt.Fprintln(w.Err, err)
			return data, nil
		}

		return migratedData, nil
	}

	// No migration required
	return data, nil
}
