package examples_test

import (
	"fmt"
	"log"

	"get.porter.sh/porter/pkg/porter"
)

func ExampleInstall() {
	// Create an instance of the Porter application
	p := porter.New()

	// This is just for our examples, you don't need it.
	prepareExample(p)

	// Specify any of the command-line arguments to pass to the install command
	installOpts := porter.NewInstallOptions()
	installOpts.Reference = "getporter/porter-hello:v0.1.1"

	// Always call validate on the options before executing. There is defaulting
	// logic in the Validate calls.
	const installationName = "porter-hello"
	err := installOpts.Validate([]string{installationName}, p)
	if err != nil {
		log.Fatal(err)
	}

	p.Claims.ListInstallations()

	// Install the bundle
	err = p.InstallBundle(installOpts)
	if err != nil {
		log.Fatal(err)
	}

	// Get the bundle's status after installing.
	showOpts := porter.ShowOptions{}
	err = showOpts.Validate([]string{installationName}, p.Context)
	if err != nil {
		log.Fatal(err)
	}

	installation, err := p.GetInstallation(showOpts)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(installation.Status)

	// Output: installing porter-hello...
	//executing install action from porter-hello (installation: porter-hello)
	//Install Hello World
	//Hello World
	//execution completed successfully!
	//succeeded
}
