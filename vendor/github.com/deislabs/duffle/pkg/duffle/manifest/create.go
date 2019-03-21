package manifest

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/deislabs/duffle/pkg/bundle"
)

const runContent = `#!/bin/bash
action=$CNAB_ACTION

if [[ $action == "install" ]]; then
echo "hey I am installing things over here"
elif [[ $action == "uninstall" ]]; then
echo "hey I am uninstalling things now"
fi
`

const dockerfileContent = `FROM alpine:latest

RUN apk update
RUN apk add -u bash

COPY Dockerfile /cnab/Dockerfile
COPY app /cnab/app

CMD ["/cnab/app/run"]
`

// Scaffold takes a path and creates a minimal duffle manifest (duffle.json)
//  and scaffolds the components in that manifest
func Scaffold(path string) error {
	name := filepath.Base(path)
	m := &Manifest{
		Name:        name,
		Version:     "0.1.0",
		Description: "A short description of your bundle",
		Keywords:    []string{name, "cnab", "tutorial"},
		Maintainers: []bundle.Maintainer{
			{
				Name:  "John Doe",
				Email: "john.doe@example.com",
				URL:   "https://example.com",
			},
			{
				Name:  "Jane Doe",
				Email: "jane.doe@example.com",
				URL:   "https://example.com",
			},
		},
		InvocationImages: map[string]*InvocationImage{
			"cnab": {
				Name:    "cnab",
				Builder: "docker",
				Configuration: map[string]string{
					"registry": "deislabs",
				},
			},
		},
	}

	d, err := json.MarshalIndent(m, "", "    ")
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(filepath.Join(path, "duffle.json"), d, 0644); err != nil {
		return err
	}
	cnabPath := filepath.Join(path, "cnab")
	if err := os.Mkdir(cnabPath, 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(filepath.Join(cnabPath, "Dockerfile"), []byte(dockerfileContent), 0644); err != nil {
		return err
	}

	appPath := filepath.Join(cnabPath, "app")
	if err := os.Mkdir(appPath, 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(filepath.Join(appPath, "run"), []byte(runContent), 0777); err != nil {
		return err
	}

	return nil
}
