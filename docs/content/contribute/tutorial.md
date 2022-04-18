---
title: Contributing Tutorial
---

We are excited to have you join the Porter community! This tutorial will walk 
you through how to modify Porter and try out your changes. We encourage you to
follow the tutorial before submitting your first pull request because this will
help you get your environment set up and become familiar with how to make a change
and test it out.

If you run into any trouble, please let us know! The best way to ask for help is
to start a new discussion on our [forum].

{{< toc >}}

[forum]: /forum/

## Try Porter

Before contributing to any open-source project, you should start with becoming
familiar with the software first. 

1. [Install Porter](/install/)
1. Run through the [QuickStart]

After this you should be able to answer:

* What is Porter?
* What is a bundle?
* How do I make a bundle?
* How do I install a bundle?

We are always iterating on our documentation. If something isn't clear and you
have read our [documentation], let us know by opening an issue. Say what you
were trying to learn about, where you looked, if you could find a relevant page,
and what still isn't clear. We can then answer your question and improve our docs
so the next person has a better experience.

[documentation]: /docs/

## Setup Environment

First let's get your computer setup so that you can work on Porter. You will
need a few things.

We are improving our support for building Porter on Windows. In a few weeks, we
will have it all working on any Windows shell. For now, if you are on Windows,
please use [Windows Subsystem for Linux][wsl] (WSL). The rest of this tutorial
assumes that you are inside your WSL distribution when installing prerequisites
and executing the commands.

* [Git]
   
   If you are new to Git or GitHub, we recommend reading the [GitHub Guides].
   They will walk you through installing Git, forking a repository and
   submitting a pull request.
* [Go](https://golang.org/doc/install) version 1.17 or higher
* [Docker]
* Make. You can install it with a package manager such as apt-get, or homebrew.
* [Mage](#install-mage)

[GitHub Guides]: https://guides.github.com/
[Git]: https://git-scm.com/book/en/v2/Getting-Started-Installing-Git
[wsl]: https://docs.microsoft.com/en-us/windows/wsl/install-win10
[Docker]: https://docs.docker.com/get-docker/

### Add GOPATH/bin to your PATH

Porter relies on a few tools that will be installed in your $GOPATH/bin directory.
When go tools are installed, they automatically are put into GOPATH/bin and in
order to use them that directory needs to be included in your PATH environment
variable. This is a standard Go developer environment configuration and will be
useful for other Go projects.

1. Open your shell profile file* and add the following line to the file:

    **Posix shells like bash and zsh**
    ```bash
    export PATH=$(go env GOPATH)/bin:$PATH
    ```
   
   **PowerShell**
   ```powershell
   $env:PATH="$(go env GOPATH)\bin;$env:PATH"
   ```
   
1. Now load the changes to your bash profile with the source command or
   open a new terminal.
   
   **Posix shells like bash and zsh**
   
   Replace PROFILE_PATH with the path to your profile
   ```bash
   source PROFILE_PATH
   ```
   
   **PowerShell**
   ```powershell
   . $profile
   ```

\* There are a bunch of different shell profiles depending on your shell and customizations.
The default locations are:

* ~/.bash_profile or ~/.bashrc
* ~/.zshrc
* PowerShell doesn't have a profile by default. Here's how to find or create a [PowerShell profile].

[PowerShell profile]: https://www.howtogeek.com/126469/how-to-create-a-powershell-profile/

## Checkout Code

Porter can either be cloned into your GOPATH (usually ~/go) or anywhere on your
computer. Porter has [many repositories](https://github.com/getporter/). If you
decide to not clone Porter into your GOPATH, you may still want to clone Porter
and its other repositories into a dedicated directory so they are easier to find
later.

```bash
git clone https://github.com/getporter/porter.git ~/go/src/get.porter.sh/porter
```

If you are planning on contributing back to the project, you'll need to
[fork](https://guides.github.com/activities/forking/) and fetch your fork. Here
is a suggested setup for managing your fork, that uses `upstream` to refer to
the original porter repository and `origin` to refer to your fork:

1. Fork the repository in GitHub and copy your fork's reference.
    ![click on the green code button on the repository's homepage, and copy the reference](/images/clone-ref.png)
1. In a terminal, cd to where you cloned Porter.
1. Rename the origin remote to upstream: 
    ```bash
    git remote rename origin upstream
    ```
1. Add a remote to your local checkout of Porter for your fork:
    ```bash
    git remote add origin https://github.com/YOURNAME/porter.git
    ```
1. Run `git remote -vv` to verify that origin is your fork and upstream is the 
    canonical repository.
    ```bash
    $ git remote -vv
    origin      https://github.com/YOURNAME/porter.git (fetch)
    origin      https://github.com/YOURNAME/porter.git (push)
    upstream    https://github.com/getporter/porter.git (fetch)
    upstream    https://github.com/getporter/porter.git (push)
    ```
1. Create a branch for your changes:
    ```bash
    git checkout -b YOURBRANCH main
    ```
1. Push your branch to your fork (origin):
    ```bash
    git push -u origin YOURBRANCH
    ```

    Afterwards you can use just `git push` to synchronize your
    local branch with your fork on GitHub.
1. Run `git branch -vv` to verify that the `main` branch is synchronized with
   `upstream/main` and your branch is synchronized with `origin/YOURBRANCH`.
    ```bash
    $ git branch -vv
    * mybranch      26d8358f [origin/mybranch] Review feedback
    main            7e120aab [upstream/main]   Bump cnab-go
    ```


### Install Mage

We are transitioning from Make to [Mage]. Installing mage isn't strictly required,
you can always run `go run mage.go TARGET` instead of `mage TARGET`. However, having
the tool saves typing and time!

Mage targets are not case-sensitive, but in our docs we use camel case to make
it easier to read. Run the following commands from the porter directory to install mage:

```bash
$ go run mage.go EnsureMage
$ mage

This is a magefile, and is a "makefile for go". See https://magefile.org/

Targets:
  <List of available targets>
```

You know that your $GOPATH/bin is configured correctly if you see a list of mage
targets.

You can enable tab completion for mage as well, so that you can type 
`mage t[TAB]` and it will complete it with the name of matching targets.

1. Install bash-completion if it isn't already installed with either `brew install
    bash-completion` (macOS) or `apt install bash-completion` (debian/ubuntu) depending
    on your operating system.
1. Copy the mage-completion.sh script to a local directory:
    ```bash
    cp scripts/mage-completion.sh ~
    ```
1. Open your ~/.bash_profile or ~/.bashrc file and add the following line to the
    file:

    ```bash
    source ~/mage-completion.sh
    ```
1. Now load the changes to your bash profile with `source ~/.bash_profile` or
   `source ~/.bashrc`.

## Configure Signing

Porter requires that [all commits are signed][dco]. Run the following command to
tell git to automatically sign your commits in the Porter repository:

```bash
make setup-dco
```

[dco]: /contribute/guide/#signing-your-commits

## Build Porter

Now that we have the source code, let's build porter and make sure everything is 
working.

```bash
make build
```

You may see a message about your Go bin not being in your PATH. If that happens,
[Add $GOPATH/bin to your PATH](#add-gopathbin-to-your-path) and then run 
`make build` again. It should work now but if it doesn't, please let us know!


## Verify Porter

After you have built Porter, the resulting `porter` command-line tool is placed 
in the bin directory. Let's try running porter to make sure everything worked:

```
./bin/porter help
```

Now you that you have built Porter, let's try running a bundle to make sure 
Docker is installed and configured properly too. This is an abbreviated version
of our [QuickStart]. If you are new to Porter, we recommend trying the 
QuickStart as well to learn how Porter works.

[QuickStart]: /quickstart/

### Use the locally built porter

First let's do some quick configuration so that you can use the porter
executable that you just built, instead of the installed porter. This change
isn't permanent and only affects your current shell session. 

If you skip this set up, and `./bin/porter` without PORTER_HOME set it
will use the files in the `~/.porter` instead of what you built. This can
result in not actually using your local binaries in bin.

**Bash**
```bash
export PORTER_HOME=`pwd`/bin
alias porter=`pwd`/bin/porter
```

**PowerShell**
```powershell
$env:PORTER_HOME="$pwd\bin"
Set-Alias -Name porter -Value "$pwd\bin\porter.exe"
```

Let's use what you just built and verify that everything is working:

1. Make a temporary directory for your bundle:
    ```bash
    mkdir -p /tmp/hello-world
    cd /tmp/hello-world
    ```
1. Create a new bundle:
    ```bash
    porter create
    ```
1. Build the bundle:
    ```bash
    porter build
    ```
1. Install the bundle:
    ```bash
    porter install
    ```
1. View your bundle's status by listing all installed bundles:
    ```bash
    porter list
    ```

## Change Porter

Let's make a change to Porter by adding a new command, `porter hello --name
YOURNAME` that prints `Hello YOURNAME!`.

1. Create a new file at pkg/porter/hello.go and paste the following content.

    This is the implementation of our new command. We are adding a new function
    to the `Porter` struct called `Hello` that accepts a `HelloOptions` struct
    containing the flags and arguments from the command. Our options struct
    has validation logic so that we can enforce rules such as `--name is required`.
    
    ```go
    package porter
    
    import (
        "fmt"
        
        "github.com/pkg/errors"
    )
    
    // Define flags and arguments for `porter hello`.
    type HelloOptions struct {
       Name string
    }
    
    // Validate the options passed to `porter hello`.
    func (o HelloOptions) Validate() error {
        if o.Name == "" {
            return errors.New("--name is required")
        }
        return nil
    }
    
    // Hello contains the implementation for `porter hello`.
    func (p *Porter) Hello(opts HelloOptions) error {
        fmt.Printf("Hello %s!\n", opts.Name)
        return nil
    }
    ```
1. Create a new file at cmd/porter/hello.go and open it in your editor.

    This is the definition of the command. It includes the name of the command,
    its help text displayed by `porter help`, defines the flags for the command,
    and says how to validate and run the command.
     
    ```go
    package main
    
    import (
        "get.porter.sh/porter/pkg/porter"
        "github.com/spf13/cobra"
    )
    
    func buildHelloCommand(p *porter.Porter) *cobra.Command {
        // Store arguments and flags specified by the user
        opts := porter.HelloOptions{}
        
        // Define the `porter hello` command
        cmd := &cobra.Command{
            Use:   "hello",
            Short: "Say hello",
            PreRunE: func(cmd *cobra.Command, args []string) error {
                return opts.Validate()
            },
            RunE: func(cmd *cobra.Command, args []string) error {
                return p.Hello(opts)
            },
        }
        
        // Define the --name flag. Allow using -n too.
        cmd.Flags().StringVarP(&opts.Name, "name", "n", "", "Your name")
        return cmd
    }

    ```
1. Open cmd/porter/main.go and look for the line `cmd.AddCommand(buildVersionCommand(p))`.
    Add your new hello command to the supported commands by adding the following
    line before we add the version command:
    
    ```go
    cmd.AddCommand(buildHelloCommand(p))
    ```

## Test Your Changes

After you have modified Porter, and [aliased the `porter` command to use your
local changes](#use-the-locally-built-porter), let's rebuild and test your changes.

1. Build porter to incorporate your new command by running `make build`.
1. Run `porter hello --help` to see the helptext for your command.
1. Run `porter hello --name YOURNAME` to try out your new command.

```bash
$ porter hello --name Carolyn
Hello Carolyn!
```

That verifies your change but let's also run the [unit tests] and [end-to-end] tests
to make sure there aren't any regressions.

> In MacOS Monterey, port 5000 is already in use blocking `mage testSmoke` from running properly. To free port 5000, uncheck `AirPlay Receiver` in Sharing under System Preferences.

```
make test-unit
mage testSmoke
```

[unit tests]: /contribute/#unit-tests
[end-to-end]: /contribute/#end-to-end-tests

## Celebrate!

You can now build Porter, modify its code and try out your changes! 🎉 Your next
steps should be to read our [Contributing Guide] to understand how to [find an
issue] and contribute to Porter.

[Contributing Guide]: /contribute/
[find an issue]: /contribute/guide/#find-an-issue
[Mage]: https://magefile.org
