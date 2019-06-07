package porter

// applyDefaultOptions applies more advanced defaults to the options
// based on values that beyond just what was supplied by the user
// such as information in the manifest itself.
func (p *Porter) applyDefaultOptions(opts *sharedOptions) error {
	if opts.File != "" {
		err := p.LoadManifestFrom(opts.File)
		if err != nil {
			return err
		}
	}

	//
	// Default the claim name to the bundle name
	//
	if opts.Name == "" && p.Manifest != nil {
		opts.Name = p.Manifest.Name
	}

	//
	// Default the porter-debug param to --debug
	//
	if _, set := opts.combinedParameters["porter-debug"]; !set && p.Debug {
		if opts.combinedParameters == nil {
			opts.combinedParameters = make(map[string]string)
		}
		opts.combinedParameters["porter-debug"] = "true"
	}

	return nil
}
