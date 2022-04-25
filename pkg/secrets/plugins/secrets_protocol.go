package plugins

// SecretsProtocol is the interface that secrets plugins must implement.
// This defines the protocol used to communicate with secrets plugins.
type SecretsProtocol interface {
	// Resolve a secret value from a secret store
	// - keyName is name of the key where the secret can be found.
	// - keyValue is the value of the key.
	// if keyName can not be reolved by a plugin implementation, the plugin will
	// fall back to use the host plugin.
	// Examples:
	// - keyName=env, keyValue=CONN_STRING
	// - keyName=key, keyValue=conn-string
	// - keyName=path, keyValue=/tmp/connstring.txt
	Resolve(keyName string, keyValue string) (string, error)

	// Create stores a secret value in a secret store.
	// - keyName is name of the key where the secret can be found.
	// - keyValue is the value of the key.
	// Examples:
	// - keyName=env, keyValue=CONN_STRING, value=redis://foo
	// - keyName=key, keyValue=conn-string, value=redis://foo
	// - keyName=path, keyValue=/tmp/connstring.txt, value=redis://foo
	Create(keyName string, keyValue string, value string) error
}
