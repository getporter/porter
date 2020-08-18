package storage

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"get.porter.sh/porter/pkg/config"
	"github.com/cnabio/cnab-go/claim"
	"github.com/cnabio/cnab-go/credentials"
	"github.com/cnabio/cnab-go/schema"
	"github.com/cnabio/cnab-go/utils/crud"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

// Manager handles high level functions over Porter's storage systems such as
// migrating data formats.
type Manager struct {
	*config.Config

	// BackingStore is the underlying storage managed by this instance. It
	// shouldn't be used for typed read/access the data, for that use the ClaimsProvider
	// or CredentialsProvider which works with the Storage.Manager.
	*crud.BackingStore

	// connMgr is responsible for providing a consolidated HandleConnect
	// implementation that merges our Connect/Close with those of the datastore.
	connMgr *crud.BackingStore

	// schemaLoaded specifies if we have loaded the schema document.
	schemaLoaded bool

	// schema document that defines the current version of each storage system.
	// We use this to detect when we are out-of-date and require a migration.
	schema Schema

	// Allow the schema to be out-of-date, defaults to false. Prevents
	// connections to underlying storage when the schema is out-of-date
	allowOutOfDateSchema bool
}

// NewManager creates a storage manager for a backing datastore.
func NewManager(c *config.Config, datastore crud.Store) *Manager {
	mgr := &Manager{
		Config:       c,
		BackingStore: crud.NewBackingStore(datastore),
	}

	mgr.connMgr = crud.NewBackingStore(mgr)

	return mgr
}

func (m *Manager) Connect() error {
	err := m.BackingStore.Connect()
	if err != nil {
		return err
	}

	if !m.schemaLoaded {
		if err := m.loadSchema(); err != nil {
			return err
		}

		if !m.allowOutOfDateSchema && m.MigrationRequired() {
			m.Close()
			return errors.New(`The schema of Porter's data is in an older format than supported by this version of Porter. 
Refer to https://porter.sh/storage-migrate for more information and instructions to back up your data. 
Once your data has been backed up, run the following command to perform the migration:

    porter storage migrate
`)
		}
		m.schemaLoaded = true
	}

	return nil
}

func (m *Manager) Close() error {
	return m.BackingStore.Close()
}

func (m *Manager) HandleConnect() (func() error, error) {
	// Use our own HandleConnect, override other crud.Store methods
	// so that we can call it instead of using the underlying datastore's connect.
	return m.connMgr.HandleConnect()
}

func (m *Manager) GetDataStore() crud.Store {
	return m.BackingStore
}

func (m *Manager) Count(itemType string, group string) (int, error) {
	handleClose, err := m.HandleConnect()
	defer handleClose()
	if err != nil {
		return 0, err
	}

	return m.BackingStore.Count(itemType, group)
}

func (m *Manager) List(itemType string, group string) ([]string, error) {
	handleClose, err := m.HandleConnect()
	defer handleClose()
	if err != nil {
		return nil, err
	}

	return m.BackingStore.List(itemType, group)
}

func (m *Manager) Save(itemType string, group string, name string, data []byte) error {
	handleClose, err := m.HandleConnect()
	defer handleClose()
	if err != nil {
		return err
	}

	return m.BackingStore.Save(itemType, group, name, data)
}

func (m *Manager) Read(itemType string, name string) ([]byte, error) {
	handleClose, err := m.HandleConnect()
	defer handleClose()
	if err != nil {
		return nil, err
	}

	return m.BackingStore.Read(itemType, name)
}

func (m *Manager) ReadAll(itemType string, group string) ([][]byte, error) {
	handleClose, err := m.HandleConnect()
	defer handleClose()
	if err != nil {
		return nil, err
	}

	return m.BackingStore.ReadAll(itemType, group)
}

func (m *Manager) Delete(itemType string, name string) error {
	handleClose, err := m.HandleConnect()
	defer handleClose()
	if err != nil {
		return err
	}

	return m.BackingStore.Delete(itemType, name)
}

// loadSchema parses the schema file at the root of PORTER_HOME. This file (when present) contains
// a list of the current version of each of Porter's storage systems.
func (m *Manager) loadSchema() error {
	b, err := m.BackingStore.Read("", "schema")
	if err != nil {
		if strings.Contains(err.Error(), crud.ErrRecordDoesNotExist.Error()) {
			emptyHome, err := m.initEmptyPorterHome(ioutil.Discard)
			if emptyHome {
				// Return any errors from creating a schema document in an empty porter home directory
				return err
			} else {
				// When we don't have an empty home directory, and we can't find the schema
				// document, we need to do a migration
				return nil
			}
		}
		return errors.Wrap(err, "could not read storage schema document")
	}

	err = json.Unmarshal(b, &m.schema)
	return errors.Wrap(err, "could not parse storage schema document")
}

// MigrationRequired determines if a migration of Porter's storage system is necessary.
func (m *Manager) MigrationRequired() bool {
	return m.ShouldMigrateClaims() || m.ShouldMigrateCredentials()
}

// Migrate executes a migration on any/all of Porter's storage sub-systems.
func (m *Manager) Migrate() (string, error) {
	// Let us call connect and not have it kick us out because the schema is out-of-date
	m.allowOutOfDateSchema = true
	defer func() {
		m.allowOutOfDateSchema = false
	}()

	// Reuse the same connection for the entire migration
	err := m.Connect()
	if err != nil {
		return "", err
	}
	defer m.Close()

	home, err := m.GetHomeDir()
	if err != nil {
		return "", err
	}

	logfilePath := filepath.Join(home, fmt.Sprintf("%s-migrate.log", time.Now().Format("20060102150405")))
	logfile, err := m.FileSystem.Create(logfilePath)
	if err != nil {
		return "", errors.Wrapf(err, "error creating logfile for migration at %s", logfilePath)
	}
	defer logfile.Close()
	w := io.MultiWriter(m.Err, logfile)

	var migrationErr *multierror.Error
	if m.ShouldMigrateClaims() {
		fmt.Fprintf(w, "Claims schema is out-of-date (want: %s got: %s)\n", claim.CNABSpecVersion, m.schema.Claims)
		err = m.migrateClaims(w)
		migrationErr = multierror.Append(migrationErr, err)
	} else {
		fmt.Fprintln(w, "Claims schema is up-to-date")
	}

	if m.ShouldMigrateCredentials() {
		fmt.Fprintf(w, "Credentials schema is out-of-date (want: %s got: %s)\n", credentials.CNABSpecVersion, m.schema.Credentials)
		err = m.migrateCredentials(w)
		migrationErr = multierror.Append(migrationErr, err)
	} else {
		fmt.Fprintln(w, "Credentials schema is up-to-date")
	}

	if migrationErr.ErrorOrNil() == nil {
		err = m.writeSchema(w)
		migrationErr = multierror.Append(migrationErr, err)
	}

	return logfilePath, migrationErr.ErrorOrNil()
}

// When there is no schema, and no existing storage data, create an initial
// schema file and allow the operation to continue. Don't require a
// migration.
func (m *Manager) initEmptyPorterHome(w io.Writer) (bool, error) {
	if m.schema != (Schema{}) {
		return false, nil
	}

	itemCheck := func(itemType string) (bool, error) {
		itemCount, err := m.BackingStore.Count(itemType, "")
		if err != nil {
			return false, errors.Wrapf(err, "error checking for existing %s when checking if PORTER_HOME is new", itemType)
		}

		return itemCount > 0, nil
	}

	hasClaims, err := itemCheck("claims")
	if hasClaims || err != nil {
		return false, err
	}

	hasCredentials, err := itemCheck("credentials")
	if hasCredentials || err != nil {
		return false, err
	}

	hasParameters, err := itemCheck("parameters")
	if hasParameters || err != nil {
		return false, err
	}

	return true, m.writeSchema(w)
}

// ShouldMigrateClaims determines if the claims storage system requires a migration.
func (m *Manager) ShouldMigrateClaims() bool {
	return string(m.schema.Claims) != claim.CNABSpecVersion
}

// ShouldMigrateCredentials determines if the credentials storage system requires a migration.
func (m *Manager) ShouldMigrateCredentials() bool {
	return string(m.schema.Credentials) != credentials.CNABSpecVersion
}

func (m *Manager) migrateClaims(w io.Writer) error {
	fmt.Fprintln(w, "Migrating claims data to match the CNAB Claim spec https://cdn.cnab.io/schema/cnab-claim-1.0.0-DRAFT+b5ed2f3/claim.schema.json. This is a one-way migration.")

	installationNames, err := m.BackingStore.List(claim.ItemTypeClaims, "")
	if err != nil {
		return errors.Wrapf(err, "Migration failed, unable to list installation names")
	}

	var migrationErr *multierror.Error
	for _, installationName := range installationNames {
		err = m.migrateInstallation(w, installationName)
		if err != nil {
			fmt.Fprintf(w, errors.Wrapf(err, "Error migrating installation %s. Skipping.\n", installationName).Error())
			migrationErr = multierror.Append(migrationErr, err)
		}
	}

	return migrationErr.ErrorOrNil()
}

// writeSchema updates the schema with the most recent version then writes it to disk.
func (m *Manager) writeSchema(w io.Writer) error {
	m.schema = Schema{
		Claims:      schema.Version(claim.CNABSpecVersion),
		Credentials: schema.Version(credentials.CNABSpecVersion),
	}
	schemaB, err := json.Marshal(m.schema)
	if err != nil {
		return errors.Wrap(err, "Unable to marshal storage schema file")
	}

	err = m.BackingStore.Save("", "", "schema", schemaB)
	if err != nil {
		errors.Wrap(err, "Unable to save storage schema file")
	}

	fmt.Fprintln(w, "Wrote updated schema.json to storage")
	return nil
}

// migrateInstallation moves the data from the older claim schema into the new format.
// This is a destructive operation, and once the migration is complete older versions
// of Porter cannot read the new format.
//
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
func (m *Manager) migrateInstallation(w io.Writer, name string) error {
	fmt.Fprintf(w, " - Migrating claim %s to the new claim layout...\n", name)

	oldClaimData, err := m.BackingStore.Read(claim.ItemTypeClaims, name)
	if err != nil {
		return errors.Wrap(err, "could not read claim file")
	}

	if getSchemaVersion(oldClaimData) == "" {
		oldClaimData, err = m.migrateUnversionedClaim(w, name, oldClaimData)
		if err != nil {
			return err
		}
	}

	var old claimd7ffba8
	err = json.Unmarshal(oldClaimData, &old)
	if err != nil {
		return errors.Wrapf(err, "could not load claim file:\n%s", string(oldClaimData))
	}

	newClaims, newResults, newOutputs, err := m.splitClaim(old)
	if err != nil {
		return errors.Wrapf(err, "could not split claim:\n%v", old)
	}

	claimStore := claim.NewClaimStore(m.BackingStore, nil, nil)
	for _, c := range newClaims {
		err = claimStore.SaveClaim(c)
		if err != nil {
			return errors.Wrapf(err, "could not save new claim:\n%v", c)
		}
	}

	for _, r := range newResults {
		err = claimStore.SaveResult(r)
		if err != nil {
			return errors.Wrapf(err, "could not save new result:\n%v", r)
		}
	}

	for _, o := range newOutputs {
		err = claimStore.SaveOutput(o)
		if err != nil {
			return errors.Wrapf(err, "could not save new output:\n%v", o)
		}
	}

	// Cleanup old / migrated data now that it has been replaced
	err = m.BackingStore.Delete(claim.ItemTypeClaims, name)
	if err != nil {
		return errors.Wrap(err, "could not remove migrated claim")
	}

	return nil
}

// splitClaim takes a claim in the old single document format and creates a set of claims, results and outputs
// in the current supported schema.
func (m *Manager) splitClaim(old claimd7ffba8) ([]claim.Claim, []claim.Result, []claim.Output, error) {
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

// migrateUnversionedClaim migrates a claim from Name -> Installation from before
// claims has a schemaVersion field.
func (m *Manager) migrateUnversionedClaim(w io.Writer, name string, data []byte) ([]byte, error) {
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
		fmt.Fprintf(w, " - Migrating claim %s from claim.Name to claim.Installation\n", name)
		fmt.Fprintf(w, "Migrating installation %s (Name -> Installation) to match the CNAB Claim spec https://cnab.io/schema/cnab-claim-1.0.0-DRAFT+d7ffba8/claim.schema.json. The Name field will be preserved for compatibility with previous versions of the spec.\n", name)
		rawData["installation"] = legacyName
		delete(rawData, "name")

		return json.Marshal(rawData)
	}

	return data, nil
}

func (m *Manager) migrateCredentials(w io.Writer) error {
	// Ensure all credentials have a schemaVersion set
	fmt.Fprintln(w, "Migrating credential set data to match the CNAB Credential Set spec https://github.com/cnabio/cnab-spec/blob/cnab-credentialsets-1.0.0-DRAFT+b6c701f/802-credential-sets.md. This is a one-way migration.")

	credStore := credentials.NewCredentialStore(m)
	creds, err := credStore.ReadAll()
	if err != nil {
		return errors.Wrapf(err, "Migration failed, unable to read all credentials")
	}

	var migrationErr *multierror.Error
	version, _ := credentials.GetDefaultSchemaVersion()
	for _, cred := range creds {
		// Set a schema version on credentials that don't have it yet
		if cred.SchemaVersion == "" {
			cred.SchemaVersion = version
		}

		err = credStore.Save(cred)
		if err != nil {
			fmt.Fprintf(w, errors.Wrapf(err, "Error migrating credential set %s. Skipping.\n", cred.Name).Error())
			migrationErr = multierror.Append(migrationErr, err)
		}
	}

	return migrationErr.ErrorOrNil()
}

// getSchemaVersion attempts to read the schemaVersion stamped on a document.
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
