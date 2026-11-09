package cnab

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"gopkg.in/yaml.v3"
)

// CredentialDefaultSourcesExtensionKey represents the full key for the credential default sources extension.
// The extension only holds hints for the user, a tool can ignore it and still
// run the bundle, so it is not listed in the bundle's required extensions.
const CredentialDefaultSourcesExtensionKey = "org.getporter.credential-default-sources"

// CredentialDefaultSource is where the value of a credential is usually found on
// the user's machine, for example ~/.kube/config. It is a suggestion from the bundle
// author that the user must accept, so it is limited to sources that only read
// a value: a bundle must not be able to suggest a command to run on the user's machine.
type CredentialDefaultSource struct {
	// EnvironmentVariable is the name of the environment variable that usually holds the credential.
	EnvironmentVariable string `json:"env,omitempty" yaml:"env,omitempty"`

	// Path is the path of the file that usually holds the credential.
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
}

// UnmarshalJSON rejects keys other than env and path, so that an unsupported
// source in bundle.json is reported instead of silently dropped.
func (s *CredentialDefaultSource) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if err := checkCredentialDefaultSourceKeys(raw); err != nil {
		return err
	}

	type rawSource CredentialDefaultSource
	var source rawSource
	if err := json.Unmarshal(data, &source); err != nil {
		return err
	}
	*s = CredentialDefaultSource(source)
	return nil
}

// UnmarshalYAML rejects keys other than env and path, so that an unsupported
// source in porter.yaml is reported instead of silently dropped.
func (s *CredentialDefaultSource) UnmarshalYAML(value *yaml.Node) error {
	var raw map[string]interface{}
	if err := value.Decode(&raw); err != nil {
		return err
	}
	if err := checkCredentialDefaultSourceKeys(raw); err != nil {
		return fmt.Errorf("line %d: %w", value.Line, err)
	}

	type rawSource CredentialDefaultSource
	var source rawSource
	if err := value.Decode(&source); err != nil {
		return err
	}
	*s = CredentialDefaultSource(source)
	return nil
}

func checkCredentialDefaultSourceKeys[T any](raw map[string]T) error {
	var unsupported []string
	for key := range raw {
		if key != "env" && key != "path" {
			unsupported = append(unsupported, key)
		}
	}
	if len(unsupported) == 0 {
		return nil
	}
	sort.Strings(unsupported)
	return fmt.Errorf("unsupported credential default source %q, only env or path can be set", unsupported[0])
}

// Validate checks that exactly one source is defined.
func (s CredentialDefaultSource) Validate() error {
	if (s.EnvironmentVariable == "") == (s.Path == "") {
		return errors.New("exactly one of env or path must be set")
	}
	return nil
}

// CredentialDefaultSources maps the name of a credential to its default source.
type CredentialDefaultSources map[string]CredentialDefaultSource

// ReadCredentialDefaultSources reads the credential default sources extension
// from the bundle. It returns nil when the bundle doesn't define the extension.
func (b ExtendedBundle) ReadCredentialDefaultSources() (CredentialDefaultSources, error) {
	data, ok := b.Custom[CredentialDefaultSourcesExtensionKey]
	if !ok {
		return nil, nil
	}

	dataB, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("could not marshal the untyped %q extension data: %w", CredentialDefaultSourcesExtensionKey, err)
	}

	sources := CredentialDefaultSources{}
	if err = json.Unmarshal(dataB, &sources); err != nil {
		return nil, fmt.Errorf("could not unmarshal the %q extension %q: %w", CredentialDefaultSourcesExtensionKey, string(dataB), err)
	}

	for name, source := range sources {
		if err = source.Validate(); err != nil {
			return nil, fmt.Errorf("invalid %q extension, credential %q: %w", CredentialDefaultSourcesExtensionKey, name, err)
		}
	}

	return sources, nil
}
