package migrate

import (
	"context"
	"database/sql"
	"time"

	"github.com/cnabio/cnab-go/bundle"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/secrets"
)

func (m *migrations) Up0001(ctx context.Context, tx *sql.Tx) error {
	db, err := m.GORM(ctx, tx)
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

	return db.Migrator().CreateTable(
		&Installation{},
		&Result{},
		&Output{},
		&Run{},
	)
}

func (m *migrations) Down0001(ctx context.Context, tx *sql.Tx) error {
	db, err := m.GORM(ctx, tx)
	if err != nil {
		return err
	}

	// tables
	type (
		Installation struct{}
		Result       struct{}
		Output       struct{}
		Run          struct{}
	)

	return db.Migrator().DropTable(
		&Installation{},
		&Result{},
		&Output{},
		&Run{},
	)
}
