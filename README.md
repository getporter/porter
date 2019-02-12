<img align="right" src="docs/static/images/porter-logo.png" width="140px" />

[![Build Status](https://dev.azure.com/cnlabs/porter/_apis/build/status/deislabs.porter?branchName=master)](https://dev.azure.com/cnlabs/porter/_build/latest?definitionId=6?branchName=master)


# Porter is a tool for building cloud installers

Porter gives you building blocks to create a cloud installer for your application, handling all the
necessary infrastructure and configuration setup. It is a declarative authoring experience that lets you
focus on what you know best: your application.

Learn more at [porter.sh](https://porter.sh)

*Want to work on Porter with us? See our [Contributing Guide](CONTRIBUTING.md)*

## Getting Started

### Creating a new installer
Use the `porter create` command to start a new project:
```
mkdir -p my-installer/ && cd my-installer/
porter create
```

This will create a file called `porter.yaml` which contains the configuration for your installer. Modify and customize this file for your application's needs.

Here is a very basic `porter.yaml` example:
```
name: my-installer
version: 0.1.0
description: "this application is extremely important"
invocationImage: my-dockerhub-user/my-installer:latest
mixins:
  - exec
install:
  - description: "Install Hello World"
    exec:
      command: bash
      arguments:
        - -c
        - echo Hello World
uninstall:
  - description: "Uninstall Hello World"
    exec:
      command: bash
      arguments:
        - -c
        - echo Goodbye World
```



### Building a CNAB bundle

The `porter build` command will create a [CNAB-compliant](https://github.com/deislabs/cnab-spec/blob/master/101-bundle-json.md) `bundle.json`, as well as build and push the associated invocation image:
```
porter build
```

Note: Make sure that the `invocationImage` listed in you `porter.yaml`  is a reference that you are able to `docker push` to and that your end-users are able to `docker pull` from.

### Running your installer using Duffle

[Duffle](https://github.com/deislabs/duffle) is an open-source tool that allows you to install and manage CNAB bundles.

The file `duffle.json` is required by Duffle. Since both Porter and Duffle are CNAB-complaint, we can simply reuse the `bundle.json` created during `porter build` and run `duffle build` to save it to the local store:
```
cp bundle.json duffle.json
duffle build .
```

You can view all bundles in the local store with `duffle bundle list`.

Afterwords, use `duffle install` to run your installer ("demo" is the unique installation name):
```
duffle install demo my-installer:0.1.0
```

The `duffle list` command can be used to show all installed applications.

If you wish to uninstall the application, you can use `duffle uninstall`:
```
duffle uninstall demo
```

## Installation

Please inspect installation scripts below prior to running.

### Mac
```
curl https://deislabs.blob.core.windows.net/porter/latest/install-mac.sh | bash
```

### Linux
```
curl https://deislabs.blob.core.windows.net/porter/latest/install-linux.sh | bash
```

### Windows
```
iwr "https://deislabs.blob.core.windows.net/porter/latest/install-windows.ps1" -UseBasicParsing | iex
```

### From source
*Requires Go 1.11+ dev environment*

```
# Clone source
mkdir -p $GOPATH/src/github.com/deislabs/
cd $GOPATH/src/github.com/deislabs/
git clone git@github.com:deislabs/porter.git
cd porter/

# Build binary
make build

# Move to PATH
sudo mv bin/porter /usr/local/bin/
```