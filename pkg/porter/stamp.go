package porter

import (
	"fmt"

	"get.porter.sh/porter/pkg/build"
	"get.porter.sh/porter/pkg/cnab"
	configadapter "get.porter.sh/porter/pkg/cnab/config-adapter"
	"github.com/pkg/errors"
)

// ensureLocalBundleIsUpToDate ensures that the bundle is up to date with the porter manifest,
// if it is out-of-date, performs a build of the bundle.
func (p *Porter) ensureLocalBundleIsUpToDate(opts bundleFileOptions) (cnab.BundleReference, error) {
	if opts.File == "" {
		return cnab.BundleReference{}, nil
	}

	upToDate, err := p.IsBundleUpToDate(opts)
	if err != nil {
		fmt.Fprintln(p.Err, "warning", err)
	}

	if !upToDate {
		fmt.Fprintln(p.Out, "Building bundle ===>")
		// opts.File is non-empty, which overrides opts.CNABFile if set
		// (which may be if a cached bundle is fetched e.g. when running an action)
		opts.CNABFile = ""
		buildOpts := BuildOptions{bundleFileOptions: opts}
		buildOpts.Validate(p)
		err := p.Build(buildOpts)
		if err != nil {
			return cnab.BundleReference{}, err
		}
	}

	bun, err := cnab.LoadBundle(p.Context, build.LOCAL_BUNDLE)
	if err != nil {
		return cnab.BundleReference{}, err
	}

	return cnab.BundleReference{
		Definition: bun,
	}, nil
}

// IsBundleUpToDate checks the hash of the manifest against the hash in cnab/bundle.json.
func (p *Porter) IsBundleUpToDate(opts bundleFileOptions) (bool, error) {
	if opts.File == "" {
		return false, errors.New("File is required")
	}
	err := p.LoadManifestFrom(opts.File)
	if err != nil {
		return false, err
	}

	if exists, _ := p.FileSystem.Exists(opts.CNABFile); exists {
		bun, err := cnab.LoadBundle(p.Context, opts.CNABFile)
		if err != nil {
			return false, errors.Wrapf(err, "could not marshal data from %s", opts.CNABFile)
		}

		// Check whether invocation images exist in host registry.
		for _, invocationImage := range bun.InvocationImages {
			isInvocationImageExists, err := p.Registry.IsInvocationImageExists(invocationImage.Image)
			if err != nil {
				return false, errors.Wrapf(err, "error while checking for existing invocation image")
			}

			if !isInvocationImageExists {
				if p.Debug {
					fmt.Fprintln(p.Err, errors.New(fmt.Sprintf("Invocation image %s doesn't exist in host registry, will need to build first", invocationImage.Image)))
				}
				return false, nil
			}
		}

		oldStamp, err := configadapter.LoadStamp(bun)
		if err != nil {
			return false, errors.Wrapf(err, "could not load stamp from %s", opts.CNABFile)
		}

		mixins, err := p.getUsedMixins()
		if err != nil {
			return false, errors.Wrapf(err, "error while listing used mixins")
		}

		converter := configadapter.NewManifestConverter(p.Context, p.Manifest, nil, mixins)
		newDigest, err := converter.DigestManifest()
		if err != nil {
			if p.Debug {
				fmt.Fprintln(p.Err, errors.Wrap(err, "could not determine if the bundle is up-to-date so will rebuild just in case"))
			}
			return false, nil
		}
		return oldStamp.ManifestDigest == newDigest, nil
	}

	return false, nil
}
