# Bundle with Private Assets

Sometimes you need to include assets from secured locations, such as a private repository in your bundle.
You can use the \--secret flag to pass secrets into the bundle when it is built.

## Try it out
1. Edit secrets/token and replace the contents with a [GitHub Personal Access Token](https://github.com/settings/tokens).
    The permissions do not matter for this sample bundle.
    There should not be a newline at the end of the file.

1. Build the bundle and pass the secret into the bundle with \--secret
    ```
    porter build --secret id=token,src=secrets/token
    ```

1. Install the bundle to see the private assets embedded in the bundle
    ```
    $ porter install example-private-assets --reference ghcr.io/getporter/examples/private-assets:v0.1.0
    __________________________
    < yarr, I'm a secret whale >
    --------------------------
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
