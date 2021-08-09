package secrets

import (
	"get.porter.sh/porter/pkg/secrets/plugins"
)

const SourceSecret = "secret"

// Store is the interface that Porter uses to interact with secrets.
type Store interface {
	plugins.SecretsPlugin
}
