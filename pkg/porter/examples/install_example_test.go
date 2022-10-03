package examples_test

import (
	"context"
	"fmt"
	"log"

	"get.porter.sh/porter/pkg/porter"
)

func ExamplePorter_install() {
	// Create an instance of the Porter application
	p := porter.New()

	// Specify any of the command-line arguments to pass to the install command
	installOpts := porter.NewInstallOptions()
	// install a bundle with older cnab bundle schema version. It should succeed
	installOpts.Reference = "ghcr.io/getporter/examples/porter-hello:v0.2.0"

	// Always call validate on the options before executing. There is defaulting
	// logic in the Validate calls.
	const installationName = "porter-hello"
	err := installOpts.Validate(context.Background(), []string{installationName}, p)
	if err != nil {
		log.Fatal(err)
	}

	// porter install porter-hello --reference ghcr.io/getporter/examples/porter-hello:v0.2.0
	err = p.InstallBundle(context.Background(), installOpts)
	if err != nil {
		log.Fatal(err)
	}

	// Get the bundle's status after installing.
	showOpts := porter.ShowOptions{}
	err = showOpts.Validate([]string{installationName}, p.Context)
	if err != nil {
		log.Fatal(err)
	}

	installation, _, err := p.GetInstallation(context.Background(), showOpts)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(installation.Status)
}
