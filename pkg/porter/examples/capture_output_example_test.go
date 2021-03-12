package examples_test

import (
	"bytes"
	"fmt"
	"log"

	"get.porter.sh/porter/pkg/porter"
)

func ExampleCaptureOutput() {
	// Create an instance of the Porter application
	p := porter.New()

	// Save output to a buffer
	output := bytes.Buffer{}
	p.Out = &output

	// porter schema
	err := p.PrintManifestSchema()
	if err != nil {
		log.Fatal(err)
	}

	// Print the json schema for porter
	fmt.Println(output.String())
}
