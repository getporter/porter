# Using the Service Fabric CLI in a Bundle

This example shows how to install and use the Service Fabric CLI (sfctl) in a
bundle.

## Try it

1. [Install porter](https://porter.sh/install).
1. Clone this repository:
    ```
    git clone https://github.com/deislabs/porter.git
    ```
1. Change to this directory:
    ```
    cd porter/examples/use-service-fabric-cli
    ```
1. Try the bundle
    ```
    porter install
    porter invoke --action=help
    ```

##  Customize It
1. Use the `dockerfile` field in **porter.yaml** to tell porter that you want to use a custom Dockerfile so that we can install the Service Fabric CLI.
1. Edit **Dockerfile.tmpl** to install the Service Fabric CLI.
1. Edit **porter.yaml** and use the `exec` mixin to execute Service Fabric CLI (sfctl) commands.
1. Run `porter install` to install the bundle.
1. Run `porter invoke --action=help` to see the custom action defined in this bundle that uses `sfctl`. In this case it calls `sfctl --help` and prints the output.
1. Run `porter uninstall` to remove the example bundle.

See the [exec mixin](https://porter.sh/mixins/exec) documentation for more
information on how to use the exec mixin.