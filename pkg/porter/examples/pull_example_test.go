package examples_test

import (
	"context"
	"fmt"
	"log"

	"get.porter.sh/porter/pkg/porter"
)

func ExamplePorter_pullBundle() {
	ctx := context.Background()
	// Create an instance of the Porter application
	p := porter.New()

	// Specify which bundle to pull and any additional flags such as --force (repull) or --insecure-registry
	pullOpts := porter.BundlePullOptions{}
	pullOpts.Reference = "ghcr.io/getporter/examples/porter-hello:v0.2.0"
	// This doesn't have a validate function, otherwise we would call it now

	// Pull a bundle to Porter's cache, ~/.porter/cache
	// This isn't exposed as a command in Porter's CLI
	cachedBundle, err := p.PullBundle(ctx, pullOpts)
	if err != nil {
		log.Fatal(err)
	}

	// Print the path to the bundle.json in Proter's cache
	fmt.Println(cachedBundle.BundlePath)
}
