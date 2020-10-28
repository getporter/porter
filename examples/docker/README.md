# Docker Example

This bundle demonstrates how to use Docker inside of a bundle! 🐳

Looking for an [example bundle with docker-compose](https://github.com/getporter/docker-compose-mixin/tree/master/examples/compose)?

```
 ____________________
< whale hello there! >
 --------------------
    \
     \
      \
                    ##        .
              ## ## ##       ==
           ## ## ## ##      ===
       /""""""""""""""""___/ ===
  ~~~ {~~ ~~~~ ~~~ ~~~~ ~~ ~ /  ===- ~~~
       \______ o          __/
        \    \        __/
          \____\______/
```

Running the bundle requires elevated privileges because you are giving the
bundle access to your host when you give it access to your local Docker daemon.
Bundles that need this access declare it in their porter.yaml with 

```yaml
required:
- docker
```

Porter requires that elevated privileges are explicitly granted to the bundle in
order to run them. You can do that in one of three ways using the [Allow Docker
Host Access setting](https://porter.sh/configuration/#allow-docker-host-access):

* Specifying the `--allow-docker-host-access` flag.
* Setting the environment variable `PORTER_ALLOW_DOCKER_HOST_ACCESS` to `true`.
* Setting the config value `allow_docker_host_access="true"` in ~/.porter/config.toml.

## Try It

1. [Install porter](https://porter.sh/install).
1. Clone this repository:
    ```
    git clone https://github.com/getporter/porter.git
    ```
1. Change to this directory:
    ```
    cd porter/examples/docker
    ```
1. Try the bundle. We are going to use the `--allow-docker-host-access` flag but you
    can also use one of the other configuration options to keep things shorter.

    ```
    porter install --allow-docker-host-access
    porter invoke --action=say --param msg='i love whales' --allow-docker-host-access
    porter uninstall --allow-docker-host-access
    ```

## Customize It

* Docker is installed using the [Dockerfile template](Dockerfile.tmpl).
* Each command is implemented in the [helper script](helpers.sh). You can
  customize the `docker` command to change the image used or pass different
  arguments.