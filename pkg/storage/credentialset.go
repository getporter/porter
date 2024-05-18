package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/schema"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/tracing"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/valuesource"
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
	Status            CredentialSetStatus `json:"status,omitempty" yaml:"status,omitempty" toml:"status,omitempty"`
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
	Credentials secrets.StrategyList `json:"credentials,omitempty" yaml:"credentials,omitempty" toml:"credentials,omitempty"`
}

// We implement a custom json marshal instead of using tags, so that we can omit zero-value timestamps
var _ json.Marshaler = CredentialSetStatus{}

// CredentialSetStatus contains additional status metadata that has been set by Porter.
type CredentialSetStatus struct {
	// Created timestamp.
	Created time.Time `json:"created,omitempty" yaml:"created,omitempty" toml:"created,omitempty"`

	// Modified timestamp.
	Modified time.Time `json:"modified,omitempty" yaml:"modified,omitempty" toml:"modified,omitempty"`
}

func NewInternalCredentialSet(creds ...secrets.SourceMap) CredentialSet {
	return CredentialSet{
		CredentialSetSpec: CredentialSetSpec{Credentials: creds},
	}
}

// NewCredentialSet creates a new CredentialSet with the required fields initialized.
func NewCredentialSet(namespace string, name string, creds ...secrets.SourceMap) CredentialSet {
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

func (c CredentialSetStatus) MarshalJSON() ([]byte, error) {
	raw := make(map[string]interface{}, 2)
	if !c.Created.IsZero() {
		raw["created"] = c.Created
	}
	if !c.Modified.IsZero() {
		raw["modified"] = c.Modified
	}
	return json.Marshal(raw)
}

func (s CredentialSet) DefaultDocumentFilter() map[string]interface{} {
	return map[string]interface{}{"namespace": s.Namespace, "name": s.Name}
}

func (s *CredentialSet) Validate(ctx context.Context, strategy schema.CheckStrategy) error {
	_, span := tracing.StartSpan(ctx,
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

	// Default the schemaType before importing into the database if it's not set already
	// SchemaType isn't really used by our code, it's a type hint for editors, but this will ensure we are consistent in our persisted documents
	if s.SchemaType == "" {
		s.SchemaType = SchemaTypeCredentialSet
	}

	// OK! Now we can do resource specific validations
	return nil
}

func (s CredentialSet) String() string {
	return fmt.Sprintf("%s/%s", s.Namespace, s.Name)
}

// ToCNAB converts this to a type accepted by the cnab-go runtime.
func (s CredentialSet) ToCNAB() valuesource.Set {
	values := make(valuesource.Set, len(s.Credentials))
	for _, cred := range s.Credentials {
		values[cred.Name] = cred.ResolvedValue
	}
	return values
}

// HasCredential determines if the specified credential is defined in the set.
func (s CredentialSet) HasCredential(name string) bool {
	for _, cred := range s.Credentials {
		if cred.Name == name {
			return true
		}
	}

	return false
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
func (s CredentialSet) ValidateBundle(spec map[string]bundle.Credential, action string) error {
	for name, cred := range spec {
		if !cred.AppliesTo(action) {
			continue
		}

		if !s.HasCredential(name) && cred.Required {
			return fmt.Errorf("bundle requires credential for %s", name)
		}
	}
	return nil
}
