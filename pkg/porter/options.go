package porter

// applyDefaultOptions applies more advanced defaults to the options
// based on values that beyond just what was supplied by the user
// such as information in the manifest itself.
func (p *Porter) applyDefaultOptions(opts *sharedOptions) error {
	//
	// Default the claim name to the bundle name
	//
	if opts.Name == "" {
		err := p.Config.LoadManifest()
		if err == nil {
			opts.Name = p.Manifest.Name
		}
	}

	//
	// Default the porter-debug param to --debug
	//
	if _, set := opts.combinedParameters["porter-debug"]; !set && p.Debug {
		opts.combinedParameters["porter-debug"] = "true"
	}

	return nil
}
