package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/schema"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/tracing"
	"github.com/cnabio/cnab-go/bundle"
	"go.opentelemetry.io/otel/attribute"
)

var _ Document = CredentialSet{}

// CredentialSet defines mappings from a credential needed by a bundle to where
// to look for it when the bundle is run. For example: Bundle needs Azure
// storage connection string, and it should look for it in an environment
// variable named `AZURE_STORATE_CONNECTION_STRING` or a key named `dev-conn`.
//
// Porter discourages storing the value of the credential directly, though
// it is possible. Instead, Porter encourages the best practice of defining
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
	// SchemaType is the type of resource in the current document.
	SchemaType string `json:"schemaType,omitempty" yaml:"schemaType,omitempty" toml:"schemaType,omitempty"`

	// SchemaVersion is the version of the credential-set schema.
	SchemaVersion cnab.SchemaVersion `json:"schemaVersion" yaml:"schemaVersion" toml:"schemaVersion"`

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
			SchemaType:    SchemaTypeCredentialSet,
			SchemaVersion: DefaultCredentialSetSchemaVersion,
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

func (s *CredentialSet) Validate(ctx context.Context, strategy schema.CheckStrategy) error {
	//lint:ignore SA4006 ignore unused context for now
	ctx, span := tracing.StartSpan(ctx,
		attribute.String("credentialset", s.String()),
		attribute.String("schemaVersion", string(s.SchemaVersion)),
		attribute.String("defaultSchemaVersion", string(DefaultCredentialSetSchemaVersion)))
	defer span.EndSpan()

	// Before we can validate, get our resource in a consistent state
	// 1. Check if we know what to do with this version of the resource
	if warnOnly, err := schema.ValidateSchemaVersion(strategy, SupportedCredentialSetSchemaVersions, string(s.SchemaVersion), DefaultCredentialSetSemverSchemaVersion); err != nil {
		if warnOnly {
			span.Warn(err.Error())
		} else {
			return span.Error(err)
		}
	}

	// 2. Check if they passed in the right resource type
	if s.SchemaType != "" && !strings.EqualFold(s.SchemaType, SchemaTypeCredentialSet) {
		return span.Errorf("invalid schemaType %s, expected %s", s.SchemaType, SchemaTypeCredentialSet)
	}

	if s.SchemaType == "" {
		// Default the schema type before importing into the database if it's not set already
		// SchemaType isn't really used by our code, it's a type hint for editors, but this will ensure we are consistent in our persisted documents
		s.SchemaType = SchemaTypeCredentialSet
	} else if !strings.EqualFold(s.SchemaType, SchemaTypeCredentialSet) {
		return fmt.Errorf("invalid schemaType %s, expected %s", s.SchemaType, SchemaTypeCredentialSet)
	}

	// OK! Now we can do resource specific validations
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
