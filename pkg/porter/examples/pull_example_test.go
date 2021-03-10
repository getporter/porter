package examples_test

import (
	"fmt"
	"log"
	"path/filepath"

	"get.porter.sh/porter/pkg/porter"
)

func ExamplePullBundle() {
	// Create an instance of the Porter application
	p := porter.New()

	// This is just for our examples, you don't need it.
	prepareExample(p)

	// Specify which bundle to pull and any additional flags such as --force (repull) or --insecure-registry
	pullOpts := porter.BundlePullOptions{}
	pullOpts.Reference = "getporter/porter-hello:v0.1.1"
	// This doesn't have a validate function, otherwise we would call it now

	// Pull a bundle to Porter's cache, ~/.porter/cache
	cachedBundle, err := p.PullBundle(pullOpts)
	if err != nil {
		log.Fatal(err)
	}

	// Get the relative path to the bundle.json in Porter's cache
	// so that our output for this example is consistent and doesn't change
	// depending on the user.
	home, _ := p.GetHomeDir()
	relativeCacheDir, _ := filepath.Rel(home, cachedBundle.BundlePath)

	fmt.Printf("cached the bundle to PORTER_HOME/%s\n", relativeCacheDir)

	// Output: cached the bundle to PORTER_HOME/cache/dfae9ef8480ec49ba194a97c8743b0e9/cnab/bundle.json
}
