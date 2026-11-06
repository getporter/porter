package cnab

import (
	"encoding/json"
	"errors"
	"fmt"
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
