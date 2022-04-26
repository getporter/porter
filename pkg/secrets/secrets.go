package secrets

import (
	"context"
)

const SourceSecret = "secret"

// Store is the interface that Porter uses to interact with secrets.
type Store interface {
	Close() error

	// Resolve a credential's value from a secret store
	// - keyName is name of the key where the secret can be found.
	// - keyValue is the value of the key.
	// Examples:
	// - keyName=env, keyValue=CONN_STRING
	// - keyName=key, keyValue=conn-string
	// - keyName=path, keyValue=/tmp/connstring.txt
	Resolve(ctx context.Context, keyName string, keyValue string) (string, error)

	// Create stores a secret value in a secret store.
	// - keyName is name of the key where the secret can be found.
	// - keyValue is the value of the key.
	// Examples:
	// - keyName=env, keyValue=CONN_STRING, value=redis://foo
	// - keyName=key, keyValue=conn-string, value=redis://foo
	// - keyName=path, keyValue=/tmp/connstring.txt, value=redis://foo
	Create(ctx context.Context, keyName string, keyValue string, value string) error
}
