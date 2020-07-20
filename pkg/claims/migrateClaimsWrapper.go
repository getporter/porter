package claims

import (
	"encoding/json"
	"fmt"
	"strings"

	"get.porter.sh/porter/pkg/storage"

	"get.porter.sh/porter/pkg/context"
	"github.com/cnabio/cnab-go/claim"
	"github.com/cnabio/cnab-go/schema"
	"github.com/cnabio/cnab-go/utils/crud"
	"github.com/pkg/errors"
)

var _ crud.Store = &migrateClaimsWrapper{}

// Migrate old claims when they are accessed.
// OLD FORMAT
// claims/
//   - INSTALLATION.json
//
// NEW FORMAT
// schema.json
// claims/
//   - INSTALLATION/
//       - CLAIMID.json
// results/
//    - CLAIMID/
//        - RESULTID.json
// outputs/
//    - RESULTID/
//        - RESULTID-OUTPUTNAME
type migrateClaimsWrapper struct {
	schemaChecked bool
	schema        storage.Schema
	*context.Context
	crud.BackingStore
	claims claim.Store
}

func newMigrateClaimsWrapper(cxt *context.Context, wrappedStore crud.Store) *migrateClaimsWrapper {
	backingStore := crud.NewBackingStore(wrappedStore)
	return &migrateClaimsWrapper{
		Context:      cxt,
		BackingStore: *backingStore,
		claims:       claim.NewClaimStore(backingStore, nil, nil),
	}
}

func (w *migrateClaimsWrapper) Connect() error {
	err := w.BackingStore.Connect()
	if err != nil {
		return err
	}

	migrate := false
	if !w.schemaChecked {
		w.AutoClose = false
		b, err := w.BackingStore.Read("", "schema")
		if err != nil {
			if strings.Contains(err.Error(), crud.ErrRecordDoesNotExist.Error()) {
				// Do a migration if we don't have a schema for the current storage layout
				migrate = true
			} else {
				return errors.Wrapf(err, "could not read storage schema document")
			}
		} else {
			err = json.Unmarshal(b, &w.schema)
			if string(w.schema.Claims) != claim.CNABSpecVersion {
				// Do a migration if the current layout doesn't match the supported layout
				migrate = true
			}
		}

		if migrate {
			err := w.MigrateAll()
			if err != nil {
				return err
			}
		}
		w.schemaChecked = true
		w.AutoClose = true
	}

	return nil
}

func (w *migrateClaimsWrapper) MigrateAll() error {
	fmt.Fprint(w.Err, "!!! Migrating claims data to match the CNAB Claim spec https://cdn.cnab.io/schema/cnab-claim-1.0.0-DRAFT+b5ed2f3/claim.schema.json. This is a one-way migration !!!\n")

	installationNames, err := w.BackingStore.List(claim.ItemTypeClaims, "")
	if err != nil {
		return errors.Wrapf(err, "!!! Migration failed, unable to list installation names")
	}

	for _, installationName := range installationNames {
		err = w.MigrateInstallation(installationName)
		if err != nil {
			fmt.Fprintf(w.Err, errors.Wrapf(err, "Error migrating installation %s. Skipping.\n", installationName).Error())
		}
	}

	w.schema.Claims = schema.Version(claim.CNABSpecVersion)
	w.schema.Credentials = schema.Version("cnab-credentials-1.0.0-DRAFT-b6c701f")
	schemaB, err := json.Marshal(w.schema)
	if err != nil {
		return errors.Wrap(err, "!!! Migration failed, unable to marshal storage schema file")
	}

	err = w.Save("", "", "schema", schemaB)
	if err != nil {
		return errors.Wrap(err, "!!! Migration failed, unable to save storage schema file")
	}

	fmt.Fprintf(w.Err, "!!! Migration complete !!!\n")
	return nil
}

func (w *migrateClaimsWrapper) MigrateInstallation(name string) error {
	fmt.Fprintf(w.Err, " - Migrating claim %s to the new claim layout...\n", name)

	oldClaimData, err := w.BackingStore.Read(claim.ItemTypeClaims, name)
	if err != nil {
		return errors.Wrap(err, "could not read claim file")
	}

	if getSchemaVersion(oldClaimData) == "" {
		oldClaimData, err = w.MigrateUnversionedClaim(name, oldClaimData)
		if err != nil {
			return err
		}
	}

	var old claimd7ffba8
	err = json.Unmarshal(oldClaimData, &old)
	if err != nil {
		return errors.Wrapf(err, "could not load claim file:\n%s", string(oldClaimData))
	}

	newClaims, newResults, newOutputs, err := w.splitClaim(old)
	if err != nil {
		return errors.Wrapf(err, "could not split claim:\n%v", old)
	}

	for _, c := range newClaims {
		err = w.claims.SaveClaim(c)
		if err != nil {
			return errors.Wrapf(err, "could not save new claim:\n%v", c)
		}
	}

	for _, r := range newResults {
		err = w.claims.SaveResult(r)
		if err != nil {
			return errors.Wrapf(err, "could not save new result:\n%v", r)
		}
	}

	for _, o := range newOutputs {
		err = w.claims.SaveOutput(o)
		if err != nil {
			return errors.Wrapf(err, "could not save new output:\n%v", o)
		}
	}

	// Cleanup old / migrated data now that it has been replaced
	err = w.Delete(claim.ItemTypeClaims, name)
	if err != nil {
		return errors.Wrap(err, "could not remove migrated claim")
	}

	return nil
}

func (w *migrateClaimsWrapper) splitClaim(old claimd7ffba8) ([]claim.Claim, []claim.Result, []claim.Output, error) {
	// Handle old status values
	switch old.Result.Status {
	case "success":
		old.Result.Status = claim.StatusSucceeded
	case "failure":
		old.Result.Status = claim.StatusFailed
	}

	claims := make([]claim.Claim, 0, 2)
	results := make([]claim.Result, 0, 2)
	// Create a claim to represent when the bundle was first installed, if more than one action is packed into the claim
	if old.Created != old.Modified && old.Result.Action != claim.ActionInstall {
		c, err := claim.New(old.Installation, claim.ActionInstall, *old.Bundle, nil)
		if err != nil {
			return nil, nil, nil, errors.Wrap(err, "error creating placeholder install claim")
		}
		c.Created = old.Created
		c.Revision = old.Revision
		c.Custom = old.Custom
		c.BundleReference = old.BundleReference
		claims = append(claims, c)

		// Record an unknown status for the install since it was overwritten on the claim
		r, err := c.NewResult(claim.StatusUnknown)
		r.Created = c.Created
		results = append(results, r)
	}

	// Create a claim to represent the last action
	c, err := claim.New(old.Installation, old.Result.Action, *old.Bundle, old.Parameters)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "error creating migrated claim")
	}
	c.Created = old.Modified
	claims = append(claims, c)

	// Record the status of the last action
	r, err := c.NewResult(old.Result.Status)
	r.Created = c.Created
	results = append(results, r)

	outputs := make([]claim.Output, 0, len(old.Outputs))
	for outputName, outputDump := range old.Outputs {
		// The outputs are map[string]interface{} but are really map[string]string, so
		// safely force them into strings and then to []byte
		outputValue := fmt.Sprintf("%v", outputDump)
		o := claim.NewOutput(c, r, outputName, []byte(outputValue))
		outputs = append(outputs, o)
	}

	return claims, results, outputs, nil
}

// MigrateUnversionedClaim migrates a claim from Name -> Installation from before
// claims has a schemaVersion field.
func (w *migrateClaimsWrapper) MigrateUnversionedClaim(name string, data []byte) ([]byte, error) {
	var rawData map[string]interface{}
	err := json.Unmarshal(data, &rawData)
	if err != nil {
		return nil, errors.Wrapf(err, "error unmarshaling claim %s to a map for the claim migration\n%s", name, string(data))
	}

	legacyName, hasLegacyName := rawData["name"]
	_, hasInstallationName := rawData["installation"]

	// Migrate claim.Name to claim.Installation, ignoring claims that have
	// already been migrated
	if hasLegacyName && !hasInstallationName {
		fmt.Fprintf(w.Err, " - Migrating claim %s from claim.Name to claim.Installation\n", name)
		fmt.Fprintf(w.Err, "Migrating installation %s (Name -> Installation) to match the CNAB Claim spec https://cnab.io/schema/cnab-claim-1.0.0-DRAFT+d7ffba8/claim.schema.json. The Name field will be preserved for compatibility with previous versions of the spec.\n", name)
		rawData["installation"] = legacyName
		delete(rawData, "name")

		return json.Marshal(rawData)
	}

	return data, nil
}

// needsMigration determines if the schema version is different
// then the version supported by the cnab-go schema version we
// are compiled against.
func getSchemaVersion(data []byte) string {
	var peek struct {
		SchemaVersion schema.Version `json:"schemaVersion"`
	}

	err := json.Unmarshal(data, &peek)
	if err != nil {
		return "unknown"
	}

	return string(peek.SchemaVersion)
}
