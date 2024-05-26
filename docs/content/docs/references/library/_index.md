---
title: "Porter Go Library"
description: "How to use Porter's Go libraries to programmatically automate Porter"
weight: 8
---

Porter's CLI is built upon a [public Go
library](https://pkg.go.dev/get.porter.sh/porter/pkg/porter) that is available
to anyone who would like to automate Porter programmatically, or access a useful
bit of functionality that isn't exposed perfectly through the CLI.

🚨 Porter does not guarantee backwards compatibility in its library. From release to release
we may make breaking changes to support new features or fix bugs in Porter. Especially before
we reach v1.0.0, as more refactoring **is** going to happen.

Every Porter command is backed by a single function in the
`get.porter.sh/pkg/porter` package that accepts a struct defining the flags and
arguments specified at the command line. You should ALWAYS call `opts.Validate`
when it is defined because that contains defaulting logic.

We recommend using the `porter.Porter` struct, which is created by
`porter.New()` when automating Porter. There are more functions and packages
exposed, but those are much more likely to change over time.

```go
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
```

If you need to set stdin/stdout/stderr, you can set `Porter.Out`. The example below demonstrates how to capture stdout.

```go
package examples_test

import (
	"bytes"
	"context"
	"fmt"
	"log"

	"get.porter.sh/porter/pkg/porter"
)

func ExamplePorter_captureOutput() {
	// Create an instance of the Porter application
	p := porter.New()

	// Save output to a buffer
	output := bytes.Buffer{}
	p.Out = &output

	// porter schema
	err := p.PrintManifestSchema(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// Print the json schema for porter
	fmt.Println(output.String())
}
```
