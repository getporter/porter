package cnabprovider

// Shared arguments for all CNAB actions supported by duffle
type ActionArguments struct {
	// Name of the claim.
	Claim string

	// Either a filepath to the bundle or the name of the bundle.
	BundleIdentifier string

	// BundleIdentifier is a filepath.
	BundleIsFile bool

	// Insecure bundle uninstallation allowed.
	Insecure bool

	// Params is the set of parameters to pass to the bundle.
	Params map[string]string

	// Either a filepath to a credential file or the name of a set of a credentials.
	CredentialIdentifiers []string

	// Driver is the CNAB-compliant driver used to run bundle actions.
	Driver string
}
