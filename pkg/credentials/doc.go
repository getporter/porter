// Package credentials provides primitives for working with Porter
// credential sets, usually refered to as "credentials" as a shorthand.
//
// Credential Sets define mappings from a credential needed by a bundle to where
// to look for it when the bundle is run. For example: Bundle needs Azure
// storage connection string and it should look for it in an environment
// variable named `AZURE_STORATE_CONNECTION_STRING` or a key named `dev-conn`.
//
// Porter discourages storing the value of the credential directly, though it
// it is possible. Instead Porter encourages the best practice of defining
// mappings in the credential sets, and then storing the values in secret stores
// such as a key/value store like Hashicorp Vault, or Azure Key Vault.
// See the get.porter.sh/porter/pkg/secrets package for more on how Porter
// handles accessing secrets.
package credentials // import "get.porter.sh/porter/pkg/credentials"
