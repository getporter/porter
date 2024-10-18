package migrate

import (
	"context"
	"database/sql"
	"time"

	"github.com/cnabio/cnab-go/bundle"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/secrets"
)

func Up0001(ctx context.Context, tx *sql.Tx) error {
	db, err := fromContext(ctx, tx)
	if err != nil {
		return err
	}

	// columns
	type (
		OCIReferenceParts struct {
			Repository string
			Version    string
			Digest     string
			Tag        string
		}
		ParameterSetSpec struct {
			SchemaType    string
			SchemaVersion cnab.SchemaVersion
			Namespace     string
			Name          string
			Labels        map[string]string    `gorm:"type:jsonb"`
			Parameters    secrets.StrategyList `gorm:"type:jsonb"`
		}
		ParameterSetStatus struct {
			Created  time.Time
			Modified time.Time
		}
		ParameterSet struct {
			ParameterSetSpec
			Status ParameterSetStatus `gorm:"embedded;embeddedPrefix:status_"`
		}
		InstallationSpec struct {
			SchemaType     string
			SchemaVersion  cnab.SchemaVersion
			Name           string
			Namespace      string
			Uninstalled    bool
			Bundle         OCIReferenceParts `gorm:"embedded;embeddedPrefix:bundle_"`
			Custom         interface{}       `gorm:"type:jsonb"`
			Labels         map[string]string `gorm:"type:jsonb"`
			CredentialSets []string          `gorm:"type:jsonb"`
			ParameterSets  []string          `gorm:"type:jsonb"`
			Parameters     ParameterSet      `gorm:"embedded;embeddedPrefix:parameters_"`
		}
		InstallationStatus struct {
			RunID           string
			Action          string
			ResultID        string
			ResultStatus    string
			Created         time.Time
			Modified        time.Time
			Installed       *time.Time
			Uninstalled     *time.Time
			BundleReference string
			BundleVersion   string
			BundleDigest    string
		}
		CredentialSetSpec struct {
			SchemaType    string
			SchemaVersion cnab.SchemaVersion
			Namespace     string
			Name          string
			Labels        map[string]string    `gorm:"type:jsonb"`
			Credentials   secrets.StrategyList `gorm:"type:jsonb"`
		}
		CredentialSetStatus struct {
			Created  time.Time
			Modified time.Time
		}
		CredentialSet struct {
			CredentialSetSpec
			Status CredentialSetStatus `gorm:"embedded;embeddedPrefix:status_"`
		}
	)

	// tables
	type (
		Installation struct {
			ID string
			InstallationSpec
			Status InstallationStatus `gorm:"embedded;embeddedPrefix:status_"`
		}
		Output struct {
			SchemaVersion cnab.SchemaVersion
			Name          string
			Namespace     string
			Installation  string
			RunID         string
			ResultID      string
			Key           string
			Value         []byte
		}
		Run struct {
			SchemaVersion      cnab.SchemaVersion
			ID                 string
			Created            time.Time
			Modified           time.Time
			Namespace          string
			Installation       string
			Revision           string
			Action             string
			Bundle             bundle.Bundle `gorm:"json"`
			BundleReference    string
			BundleDigest       string
			ParameterOverrides ParameterSet `gorm:"embedded;embeddedPrefix:parameterOverrides_"`
			CredentialSets     []string     `gorm:"type:jsonb"`
			ParameterSets      []string     `gorm:"type:jsonb"`
			Parameters         ParameterSet `gorm:"embedded;embeddedPrefix:parameters_"`
			Custom             interface{}  `gorm:"type:jsonb"`
			ParametersDigest   string
			Credentials        CredentialSet `gorm:"embedded;embeddedPrefix:credentials_"`
			CredentialsDigest  string
		}
		Result struct {
			SchemaVersion  cnab.SchemaVersion
			ID             string
			Created        time.Time
			Namespace      string
			Installation   string
			RunID          string
			Message        string
			Status         string
			OutputMetadata cnab.OutputMetadata `gorm:"type:jsonb"`
			Custom         interface{}         `gorm:"type:jsonb"`
		}
	)

	err = db.Migrator().CreateTable(
		&Installation{},
		&Result{},
		&Output{},
		&Run{},
		&CredentialSet{},
		&ParameterSet{},
	)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
	CREATE UNIQUE INDEX IF NOT EXISTS "idx_installations_namespace_name" ON "installations" ("namespace", "name");
	CREATE INDEX IF NOT EXISTS "idx_runs_namespace_installation" ON "runs" ("namespace", "installation");
	CREATE INDEX IF NOT EXISTS "idx_results_namespace_installation" ON "results" ("namespace", "installation");
	CREATE INDEX IF NOT EXISTS "idx_results_run_id" ON "results" ("run_id");
	CREATE INDEX IF NOT EXISTS "idx_outputs_namespace_installation_result_id" ON "outputs" ("namespace", "installation", "result_id" DESC);
	CREATE UNIQUE INDEX IF NOT EXISTS "idx_outputs_result_id_name" ON "outputs" ("result_id", "name");
	CREATE INDEX IF NOT EXISTS "idx_outputs_namespace_installation_name_result_id" ON "outputs" ("namespace", "installation", "name", "result_id" DESC);
	CREATE UNIQUE INDEX IF NOT EXISTS "idx_credentials_namespace_name" ON "credential_sets" ("namespace", "name");
	CREATE UNIQUE INDEX IF NOT EXISTS "idx_parameters_namespace_name" ON "parameter_sets" ("namespace", "name");
`)
	if err != nil {
		return err
	}

	return nil
}

func Down0001(ctx context.Context, tx *sql.Tx) error {
	db, err := fromContext(ctx, tx)
	if err != nil {
		return err
	}

	// tables
	type (
		Installation  struct{}
		Result        struct{}
		Output        struct{}
		Run           struct{}
		CredentialSet struct{}
		ParameterSet  struct{}
	)

	return db.Migrator().DropTable(
		&Installation{},
		&Result{},
		&Output{},
		&Run{},
		&CredentialSet{},
		&ParameterSet{},
	)
}
