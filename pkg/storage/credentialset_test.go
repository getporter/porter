package storage

import (
	"strings"
	"testing"
	"time"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/encoding"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/tests"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCredentialSet(t *testing.T) {
	cs := NewCredentialSet("dev", "mycreds", secrets.Strategy{
		Name: "password",
		Source: secrets.Source{
			Key:   "env",
			Value: "MY_PASSWORD",
		},
	})

	assert.Equal(t, "mycreds", cs.Name, "Name was not set")
	assert.Equal(t, "dev", cs.Namespace, "Namespace was not set")
	assert.NotEmpty(t, cs.Status.Created, "Created was not set")
	assert.NotEmpty(t, cs.Status.Modified, "Modified was not set")
	assert.Equal(t, cs.Status.Created, cs.Status.Modified, "Created and Modified should have the same timestamp")
	assert.Equal(t, SchemaTypeCredentialSet, cs.SchemaType, "incorrect SchemaType")
	assert.Equal(t, CredentialSetSchemaVersion, cs.SchemaVersion, "incorrect SchemaVersion")
	assert.Len(t, cs.Credentials, 1, "Credentials should be initialized with 1 value")
}

func TestValidate(t *testing.T) {
	t.Run("valid - credential specified", func(t *testing.T) {
		spec := map[string]bundle.Credential{
			"kubeconfig": {},
		}
		values := secrets.Set{
			"kubeconfig": "top secret creds",
		}
		err := Validate(values, spec, "install")
		require.NoError(t, err, "expected Validate to pass because the credential was specified")
	})

	t.Run("valid - credential not required", func(t *testing.T) {
		spec := map[string]bundle.Credential{
			"kubeconfig": {ApplyTo: []string{"install"}, Required: false},
		}
		values := secrets.Set{}
		err := Validate(values, spec, "install")
		require.NoError(t, err, "expected Validate to pass because the credential isn't required")
	})

	t.Run("valid - missing inapplicable credential", func(t *testing.T) {
		spec := map[string]bundle.Credential{
			"kubeconfig": {ApplyTo: []string{"install"}, Required: true},
		}
		values := secrets.Set{}
		err := Validate(values, spec, "custom")
		require.NoError(t, err, "expected Validate to pass because the credential isn't applicable to the custom action")
	})

	t.Run("invalid - missing required credential", func(t *testing.T) {
		spec := map[string]bundle.Credential{
			"kubeconfig": {ApplyTo: []string{"install"}, Required: true},
		}
		values := secrets.Set{}
		err := Validate(values, spec, "install")
		require.Error(t, err, "expected Validate to fail because the credential applies to the specified action and is required")
		assert.Contains(t, err.Error(), "bundle requires credential")
	})
}

func TestDisplayCredentials_Validate(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name          string
		schemaType    string
		schemaVersion cnab.SchemaVersion
		wantError     string
	}{
		{
			name:          "schemaType: none",
			schemaType:    "",
			schemaVersion: CredentialSetSchemaVersion,
			wantError:     ""},
		{
			name:          "schemaType: CredentialSet",
			schemaType:    SchemaTypeCredentialSet,
			schemaVersion: CredentialSetSchemaVersion,
			wantError:     ""},
		{
			name:          "schemaType: CREDENTIALSET",
			schemaType:    strings.ToUpper(SchemaTypeCredentialSet),
			schemaVersion: CredentialSetSchemaVersion,
			wantError:     ""},
		{
			name:          "schemaType: credentialset",
			schemaType:    strings.ToLower(SchemaTypeCredentialSet),
			schemaVersion: CredentialSetSchemaVersion,
			wantError:     ""},
		{
			name:          "schemaType: ParameterSet",
			schemaType:    SchemaTypeParameterSet,
			schemaVersion: CredentialSetSchemaVersion,
			wantError:     "invalid schemaType ParameterSet, expected CredentialSet"},
		{
			name:          "validate embedded cs",
			schemaType:    SchemaTypeCredentialSet,
			schemaVersion: "", // required!
			wantError:     "invalid schemaVersion provided: (none)"},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cs := CredentialSet{
				CredentialSetSpec: CredentialSetSpec{
					SchemaType:    tc.schemaType,
					SchemaVersion: tc.schemaVersion,
				}}

			err := cs.Validate()
			if tc.wantError == "" {
				require.NoError(t, err)
			} else {
				tests.RequireErrorContains(t, err, tc.wantError)
			}
		})
	}
}

func TestCredentialSet_Validate_DefaultSchemaType(t *testing.T) {
	cs := NewCredentialSet("", "mycs")
	cs.SchemaType = ""
	require.NoError(t, cs.Validate())
	assert.Equal(t, SchemaTypeCredentialSet, cs.SchemaType)
}

func TestMarshal(t *testing.T) {
	now, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z07:00")

	orig := CredentialSet{
		CredentialSetSpec: CredentialSetSpec{
			SchemaVersion: "schemaVersion",
			Namespace:     "namespace",
			Name:          "name",
			Credentials: []secrets.Strategy{
				{
					Name: "cred1",
					Source: secrets.Source{
						Key:   "secret",
						Value: "mysecret",
					},
				},
			},
		},
		Status: CredentialSetStatus{
			Created:  now,
			Modified: now,
		},
	}

	formats := []string{"json", "yaml"}
	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			raw, err := encoding.Marshal(format, orig)
			require.NoError(t, err)

			var copy CredentialSet
			err = encoding.Unmarshal(format, raw, &copy)
			require.NoError(t, err)
			assert.Equal(t, orig, copy)
		})
	}
}

func TestCredentialSet_String(t *testing.T) {
	t.Run("global namespace", func(t *testing.T) {
		cs := CredentialSet{CredentialSetSpec: CredentialSetSpec{Name: "mycreds"}}
		assert.Equal(t, "/mycreds", cs.String())
	})

	t.Run("local namespace", func(t *testing.T) {
		cs := CredentialSet{CredentialSetSpec: CredentialSetSpec{Namespace: "dev", Name: "mycreds"}}
		assert.Equal(t, "dev/mycreds", cs.String())
	})
}
