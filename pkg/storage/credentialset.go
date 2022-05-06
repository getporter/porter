package storage

import (
	"fmt"
	"time"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/secrets"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/schema"
	"github.com/pkg/errors"
)

var _ Document = &CredentialSet{}

// CredentialSet defines mappings from a credential needed by a bundle to where
// to look for it when the bundle is run. For example: Bundle needs Azure
// storage connection string and it should look for it in an environment
// variable named `AZURE_STORATE_CONNECTION_STRING` or a key named `dev-conn`.
//
// Porter discourages storing the value of the credential directly, though
// it is possible. Instead Porter encourages the best practice of defining
// mappings in the credential sets, and then storing the values in secret stores
// such as a key/value store like Hashicorp Vault, or Azure Key Vault.
// See the get.porter.sh/porter/pkg/secrets package for more on how Porter
// handles accessing secrets.
type CredentialSet struct {
	CredentialSetSpec `yaml:",inline"`
	Status            CredentialSetStatus `json:"status" yaml:"status" toml:"status"`
}

// CredentialSetSpec represents the set of user-modifiable fields on a CredentialSet.
type CredentialSetSpec struct {
	// ID is the unique ULID assigned to the CredentialSet.
	ID string `json:"_id" yaml:"_id" toml:"_id"`

	// SchemaVersion is the version of the credential-set schema.
	SchemaVersion schema.Version `json:"schemaVersion" yaml:"schemaVersion" toml:"schemaVersion"`

	// Namespace to which the credential set is scoped.
	Namespace string `json:"namespace" yaml:"namespace" toml:"namespace"`

	// Name of the credential set.
	Name string `json:"name" yaml:"name" toml:"name"`

	// Labels applied to the credential set.
	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty" toml:"labels,omitempty"`

	// Credentials is a list of credential resolution strategies.
	Credentials []secrets.Strategy `json:"credentials" yaml:"credentials" toml:"credentials"`
}

// CredentialSetStatus contains additional status metadata that has been set by Porter.
type CredentialSetStatus struct {
	// Created timestamp.
	Created time.Time `json:"created" yaml:"created" toml:"created"`

	// Modified timestamp.
	Modified time.Time `json:"modified" yaml:"modified" toml:"modified"`
}

// NewCredentialSet creates a new CredentialSet with the required fields initialized.
func NewCredentialSet(namespace string, name string, creds ...secrets.Strategy) CredentialSet {
	now := time.Now()
	cs := CredentialSet{
		CredentialSetSpec: CredentialSetSpec{
			ID:            cnab.NewULID(),
			SchemaVersion: CredentialSetSchemaVersion,
			Name:          name,
			Namespace:     namespace,
			Credentials:   creds,
		},
		Status: CredentialSetStatus{
			Created:  now,
			Modified: now,
		},
	}

	return cs
}

func (s CredentialSet) DefaultDocumentFilter() map[string]interface{} {
	return map[string]interface{}{"namespace": s.Namespace, "name": s.Name}
}

func (s CredentialSet) Validate() error {
	if CredentialSetSchemaVersion != s.SchemaVersion {
		if s.SchemaVersion == "" {
			s.SchemaVersion = "(none)"
		}
		return errors.Errorf("invalid schemaVersion provided: %s. This version of Porter is compatible with %s.", s.SchemaVersion, CredentialSetSchemaVersion)
	}
	return nil
}

func (s CredentialSet) String() string {
	return fmt.Sprintf("%s/%s", s.Namespace, s.Name)
}

// Validate compares the given credentials with the spec.
//
// This will result in an error only when the following conditions are true:
// - a credential in the spec is not present in the given set
// - the credential is required
// - the credential applies to the specified action
//
// It is allowed for spec to specify both an env var and a file. In such case, if
// the given set provides either, it will be considered valid.
func Validate(given secrets.Set, spec map[string]bundle.Credential, action string) error {
	for name, cred := range spec {
		if !cred.AppliesTo(action) {
			continue
		}

		if !given.IsValid(name) && cred.Required {
			return fmt.Errorf("bundle requires credential for %s", name)
		}
	}
	return nil
}
