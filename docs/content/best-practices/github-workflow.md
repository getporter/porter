---
title: Best practices for Porter in a CI Pipeline
description: How to effectively use a GitHub workflow to create a CI pipeline using Porter.
---

To properly test your bundle in a CI pipeline, you can utilize a GitHub workflow. 
You can have the workflow run when you create a pull request, when you merge into 
main, or both. We will go through the best practices of a CI pipeline for your 
bundle and show you how to set up the GitHub workflow. [Here](https://github.com/getporter/pipeline-demo/blob/main/.github/workflows/publish.yaml) is the full working example workflow explained in this article.

## Parts of the Workflow

We will go through each of the following components of the workflow. 

* [Check out code](#check-out-code)
* [Set up Porter](#set-up-porter)
* [Login to DockerHub to publish the bundle](#login-to-dockerhub)
* [Install mixins](#install-mixins)
* [Run Porter commands](#run-porter-commands)

### Check out code
In order to test your bundle in a workflow, you need to checkout your repo. We 
used the GitHub action [Checkout](https://github.com/actions/checkout) to do this. 
By default, the action checks out the single commit that triggered the workflow. You
can customize what you want to happen in the checkout action based on their documentation, 
but if you do not need to customize, the code below in a step in your workflow is all that 
is needed.
````yaml
- uses: actions/checkout@v1
````

### Set up Porter
After checking out the code in the repository, you need to install porter in the workflow
so that you can run porter commands. The [Porter GitHub Action](https://github.com/getporter/gh-action) takes care of installing Porter for you. Adding this 
action to your workflow will install Porter for you. For example:
````yaml
- name: Setup Porter
  uses: getporter/gh-action@v0.1.1
  with:
    porter_version: v0.27.2
````
The porter_version should be the version of Porter you want installed. You can check [our releases](https://github.com/getporter/porter/releases) for the list of recent versions of Porter. When not specified, porter_version defaults to latest version of Porter. 

### Login to DockerHub
Next, you will want to login to Docker Hub so that you can publish your bundle to a Docker Hub registry. If you do not wish to publish your bundle to a Docker Hub registry, you can configure another registry in this step instead.
In order to do this, you can use the [docker-login](https://github.com/Azure/docker-login) action
to do it. Below is an example of how it is used:
````yaml
- uses: azure/docker-login@v1
  name: Docker Login
  with:
    username: ${{ secrets.DOCKER_USERNAME }}
    password: ${{ secrets.DOCKER_PASSWORD }}
````
You will need to set your docker username and password as [secrets](https://docs.github.com/en/actions/configuring-and-managing-workflows/creating-and-storing-encrypted-secrets) in the repository you are setting up the workflow in.

### Install mixins
Next, you should install any mixins your bundle will use so that it can be tested properly. You
can add a line as a run command to install your mixin as shown below:
````yaml
run: porter mixins install az
````
You can also specify the version of the mixin by adding the version flag. For example:
````yaml
run: porter mixins install az --version v0.4.2
````
Take a look at the [documentation](/mixins) to see all available mixins and how to know what command to run to install them. The documentation provides install commands for each mixin. The command `porter mixin search` lists all the mixins created by the community. Depending on if the URL is an atom feed or a github url you can install the mixin using
```yaml
porter mixins install NAME --feed-url ATOM_URL
```
or
```yaml
porter mixins install NAME --url GITHUB_URL
```

### Run Porter commands
The final part of the workflow is running porter commands. The commands we suggest running are porter install, porter upgrade, porter uninstall, and porter publish. Porter install will install your bundle and give an error if something is wrong. This will be useful as part of your pipeline in testing your bundle because you can go fix the problem right away instead of finding it later when users try to run your bundle. Porter upgrade and porter uninstall will execute the code and also give errors if there are problems. If all these commands run successfully, you can run porter publish to publish your working bundle image to a registry. In this example, we publish to Docker Hub, but porter publish can publish to any registry. 

## Set up the Workflow
Now that we know the parts that are needed in a workflow, we can learn how to set one up. Here are the steps we will go through to set up the workflow: 

* [Make yaml files](#make-yaml-files)
* [Use credential files](#use-credential-files)

### Make yaml files
The way you set up your yaml files depends on who will be contributing to your repository and how they will be contributing. GitHub actions and workflows do not run automatically on pull requests from forks. 

So, if you are the only one working on a project and no one else can make pull requests (you have a private repository) or no one will be contributing from a fork, you can set up one yaml file to run when a pull request opens and one yaml file to run when a commit is merged. You can have the one that runs on a pull request run your bundle to test it, and you can have the second yaml file publish the bundle to Docker Hub. 

If you are not the only one contributing to the repository and other contributors will be making pull requests from forks, the yaml file setup is slightly different. Since the GitHub action won't be triggered from a fork, you will only need one yaml file that runs when a pull request is merged. You can test the bundle and publish in the same yaml file.  


### Use credential files
If you were using credentials in your bundle, you will need to set up a credential file in your repository to use with your workflow. For example, if you run `porter credentials create mybun.json`, a JSON file named mybun.json is created. The resulting [credential set template file](/reference/file-formats/#credential-set) (which does not contain any sensitive credentials) is located in the current directory. You can run `porter credentials apply mybun.json` to apply the changes. If the specified credential set already exists, the changes will override the existing credential set, otherwise the apply command will create a new credential set. Then, to install your bundle with the newly created credential set, you would run the following:
```yaml
porter install -c <name-of-the-credential-set>
```

## Example Code

Now, we will show example code for a workflow and explain what the code does.

```yaml
name: CI

# On controls when the action will run. This triggers the workflow on push event for the main branch. 
# Can change push to pull_request to run when a PR is made. 
# You can also change the branch name from main to the name of the branch you want the workflow to run on. 
on:
  push:
    branches: [ main ]

# Set up environment variables needed for the bundle. 
# If these are sensitive, they should be set as secrets in the repository. 
# To do this, go to settings -> secrets -> new secret.
env:
  DOCKER_USERNAME: ${{ secrets.DOCKER_USERNAME }}
  DOCKER_PASSWORD: ${{ secrets.DOCKER_PASSWORD }}

# A workflow can be made of many jobs, but this example puts all the steps under one job.
jobs:
  # Publish is the name of this job.
  publish: 
    # Specify the type of machine you want the job to run on. We chose ubuntu-latest. 
    runs-on: ubuntu-latest
    
    steps: 
    # Check out code
    - uses: actions/checkout@v1
    # Use Porter GH action to set up Porter. 
    # You can specify the version of Porter that you want installed by adding the lines for with and porter_version as explained above. 
    - name: Setup Porter
      uses: getporter/gh-action@v0.1.1
    # Install docker mixin needed for this bundle. 
    # Add lines to install any of the mixins your bundle needs to be able to run.
    - name: Install Docker mixin
      run: porter mixins install docker
    # Run install
    - name: Porter install
      run: porter install -c ./docker-example.json --allow-docker-host-access
    # Run upgrade
    - name: Porter upgrade
      run: porter upgrade -c ./docker-example.json --allow-docker-host-access
    # Run uninstall
    - name: Porter uninstall
      run: porter uninstall -c ./docker-example.json --allow-docker-host-access
    # Login to Docker Hub to publish the bundle
    - uses: azure/docker-login@v1
      name: Docker Login
      with:
        username: ${{ secrets.DOCKER_USERNAME }}
        password: ${{ secrets.DOCKER_PASSWORD }}
    # Run publish. 
    # If any of the porter commands above fail, the workflow will stop, so your bundle will only be published if it works properly.
    - name: Porter Publish
      run: porter publish
```
