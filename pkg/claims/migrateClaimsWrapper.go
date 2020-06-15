package claims

import (
	"encoding/json"
	"fmt"
	"strings"

	"get.porter.sh/porter/pkg/context"
	"github.com/cnabio/cnab-go/claim"
	"github.com/cnabio/cnab-go/schema"
	"github.com/cnabio/cnab-go/utils/crud"
	"github.com/pkg/errors"
)

var _ storeWithConnect = &migrateClaimsWrapper{}

type storeWithConnect interface {
	crud.Store
	crud.HasConnect
}

// Migrate old claims when they are accessed.
// OLD FORMAT
// claims/
//   - INSTALLATION.json
//
// NEW FORMAT
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
	*context.Context
	storeWithConnect
	claims claim.Store
}

func newMigrateClaimsWrapper(cxt *context.Context, wrappedStore crud.Store) *migrateClaimsWrapper {
	backingStore := crud.NewBackingStore(wrappedStore)
	return &migrateClaimsWrapper{
		Context:          cxt,
		storeWithConnect: backingStore,
		claims:           claim.NewClaimStore(backingStore, nil, nil),
	}
}

// Store is the data store being wrapped.
func (w *migrateClaimsWrapper) Store() crud.Store {
	return w.storeWithConnect
}

func (w *migrateClaimsWrapper) List(itemType string, group string) ([]string, error) {
	names, err := w.Store().List(itemType, group)
	if err != nil {
		if strings.Contains(err.Error(), crud.ErrRecordDoesNotExist.Error()) {

			// Proactively migrate all data in the claims/ dir if we detect the layout is old
			err := w.MigrateAll()
			if err != nil {
				return nil, err
			}

			return w.Store().List(itemType, group)

		}

		return nil, err
	}

	return names, nil
}

func (w *migrateClaimsWrapper) Read(itemType string, name string) ([]byte, error) {
	data, err := w.Store().Read(itemType, name)
	if err != nil {
		// If we can't read the data, check for the data layout migration from rewriting the claims spec
		if getSchemaVersion(data) != claim.CNABSpecVersion {

			// Proactively migrate all data in the claims/ dir if we detect the layout is old
			err := w.MigrateAll()
			if err != nil {
				return nil, err
			}

			return w.Store().Read(itemType, name)
		}

		return nil, err
	}

	return data, nil
}

func (w *migrateClaimsWrapper) MigrateAll() error {
	fmt.Fprint(w.Err, "!!! Migrating claims data to match the CNAB Claim spec https://cdn.cnab.io/schema/cnab-claim-1.0.0-DRAFT+b5ed2f3/claim.schema.json. This is a one-way migration !!!\n")

	installationNames, err := w.Store().List(claim.ItemTypeClaims, "")
	if err != nil {
		return errors.Wrapf(err, "!!! Migration failed, unable to list installation names")
	}

	for _, installationName := range installationNames {
		err = w.MigrateInstallation(installationName)
		if err != nil {
			fmt.Fprintf(w.Err, errors.Wrapf(err, "Error migrating installation %s. Skipping.\n", installationName).Error())
		}
	}

	fmt.Fprintf(w.Err, "!!! Migration complete !!!\n")
	return nil
}

func (w *migrateClaimsWrapper) MigrateInstallation(name string) error {
	fmt.Fprintf(w.Err, " - Migrating claim %s to the new claim layout...\n", name)

	oldClaimData, err := w.Store().Read(claim.ItemTypeClaims, name)
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
