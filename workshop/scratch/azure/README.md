# Build an Azure + Terraform Cloud Native Application Bundle - CNAB The Hardway

This exercise will show you how to build a Terraform based Cloud Native Application Bundle (CNAB) for Azure without using any tooling so that you can better understand what is involved in creating a CNAB.

## Prerequisites

In order to complete this exercise, you will need to have a recent Docker Desktop installed or you'll want to use [Play With Docker](https://labs.play-with-docker.com/) and you'll also need a Docker registry that you can push to. If you don't have one, a free DockerHub account will work. To create one of those, please visit https://hub.docker.com/signup.

Once you have a Docker account, export your username to an environment variable named REGISTRY:


```
export REGISTRY=<yourusername>
```

## Review the cnab/ directory

Start with the `cnab\app\terraform` directory. This directory contains a set of Terraform configurations that will utilize the Azure provider to create an Azure MySQL instance and store the TF state file in an Azure storage account. The files in here aren't special for use with CNAB or Porter.

Next, review the `cnab\app\run` script. This script serves as the CNAB [run tool](https://github.com/cnabio/cnab-spec/blob/master/103-bundle-runtime.md#the-run-tool-main-entry-point). In this bash script, you'll see some special things like:

```
action=$CNAB_ACTION
name=$CNAB_INSTALLATION_NAME
tfdir=/cnab/app/terraform/
```
and

```
case $action in
    install)
        echo "Applying Terraform template"
        terraform apply -auto-approve -input=false
        ;;
    uninstall)
        echo "Destroying Terraform template"
        terraform destroy -auto-approve
        ;;
    upgrade)
        echo "Applying Terraform template"
        terraform apply -auto-approve -input=false
        ;;
    status)
        echo "Status action"
        terraform plan
        ;;
    *)
    echo "No action for $action"
    ;;
esac
```

These lines are how our run script knows what CNAB action is being performed and what it should do for each action. When you build a bundle from scratch, you'll build a file very similar to this in each instance. You first check the action, then perform the appropriate command or commands that fulfil that action. In this case, we are using the `terraform` command line tool to execute the  sub-command that corresponds to our CNAB action. When you review the file, you'll also see that our bundle will use the `azure cli` to create a storage account and will use that storage account to configure the Terraform Azure backend.

## Review the Dockerfile

Once you've reviewed the `cnab` directory, move on to the `Dockerfile`. One of the core CNAB components is the `invocation image`. You can think of this like the "installer". In most cases, it will be a Dockerfile that contains all the configuration, executables and glue code necessary to install your cloud native application. In this case, we need to install both `terraform` and the `azure cli`. Of course, this also means we need to install `python` and `pip`, as they are required for the `azure cli`. When you build a bundle from scratch, you will need to ensure that the invocation image has all the tools needed to complete the actions for your bundle. You'll also need to ensure that any configuration is included, along with the run tool that we saw in the `cnab` directory. In our case, we add the `cnab` directory to the Docker image, and ensure that both the `run` script and the `init-backend` script are executable. Finally, we define the `CMD` that will invoke the `run` script when a CNAB installer uses our invocation image.

## Build The Invocation Image

Now that you've reviewed the Dockerfile, go ahead and build the invocation image.

```
docker build -t $REGISTRY/tf-mysql .
```

Once that is complete, you've build the first part of the CNAB! Next, you should push it to a registry so that you can obtain the `digest` for the invocation image.

## Push the Invocation Image

In order to obtain the `digest` for the invocation image, you'll need to push it to a Docker registry/

```
docker push $REGISTRY/tf-mysql:latest
```

Once that completes, you'll want to grab the digest for the image. For example, given the following output:

```
$ docker push $REGISTRY/tf-mysql:latest
The push refers to repository [docker.io/jeremyrickard/tf-mysql]
6b195cf9fe0e: Layer already exists
a41c721cfff0: Layer already exists
69b63698390a: Layer already exists
c61049e04aa3: Layer already exists
41637553cad0: Layer already exists
256a7af3acb1: Layer already exists
latest: digest: sha256:566e1d2daf5b5d3a97f20249d53611be285ea79a9a43570da85f84d653c2622f size: 1574
```

You'd want to copy the digest `sha256:566e1d2daf5b5d3a97f20249d53611be285ea79a9a43570da85f84d653c2622f`. This will be used in the bundle.json in the next step.

## Modify the bundle.json File

Finally, you're ready to complete the CNAB! To do this, open up the `bundle.json` file. This file serves as the `metadata` for the bundle. In this file, you'll declare what invocation image to use, define the `parameters`, `credentials` and `outputs` and list any other Docker images that will be used. In this case, we don't have any other Docker images or any outputs, but we will need users to provide some parameters and credentials. We define those like this:

```
	"credentials": {
		"CLIENT_ID": {
			"env": "TF_VAR_client_id"
		},
		"CLIENT_SECRET": {
			"env": "TF_VAR_client_secret"
		},
		"SUBSCRIPTION_ID": {
			"env": "TF_VAR_subscription_id"
		},
		"TENANT_ID": {
			"env": "TF_VAR_tenant_id"
		}
  },
  "definitions" : {
    "admin" :{
      "default": "mysql-admin",
      "type" : "string"
    },
    "backend" : {
      "default": true,
      "type": "boolean"
    },
    "backend_storage_account" : {
      "type": "string"
    }, 
    "backend_storage_container" : {
      "default": "tf-storage",
      "type": "string"
    }, 
    "backend_storage_resource_group" : {
      "default" : "devops-days-msp",
      "type": "string"
    }, 
    "location" : {
      "default" : "EastUS",
      "type": "string"
    }, 
    "string": {
      "type": "string"
    }
  },
  "parameters": {
		"fields": {
			"server_name": {
        "definition": "string",
				"destination": {
					"env": "TF_VAR_server-name"
				}
			},
			"admin_user": {
        "definition": "admin",
				"destination": {
					"env": "TF_VAR_mysql-admin"
				}
			
      },
      "backend": {
        "definition" : "backend",
        "destination" : {
          "env" : "TF_VAR_backend"
        }
      },
      "backend_storage_account": {
        "definition": "backend_storage_account",
        "destination": {
					"env": "TF_VAR_backend_storage_account"
        }
      },
      "backend_storage_container" : {
        "definition": "backend_storage_container",
        "destination": {
					"env": "TF_VAR_backend_storage_container"
				}
      },
      "backend_storage_resource_group" : {
        "definition": "backend_storage_resource_group",
        "destination": {
					"env": "TF_VAR_backend_storage_resource_group"
				}
      },
      "database-name" : {
        "definition": "string",
				"destination": {
					"env": "TF_VAR_database-name"
				}
      },
      "location": {
        "definition": "location",
				"destination": {
					"env": "TF_VAR_location"
				}
      }
		},
		"required": ["backend_storage_account", "server-name", "database-name"]
	}
```

JSON Schema is used to define the data type for each parameter. A `bundle.json` will normally have 1 or more `definitions`. Each of these `definitions` is referenced in the `parameters` section. Parameters and credentials have some similarities, but CNAB uses them in different ways. Check out the [cnab-spec](https://github.com/cnabio/cnab-spec/blob/master/101-bundle-json.md#definitions) for more information on definitions, parameters and credentials.

In order to produce a valid `bundle.json` using your new invocation image, you'll need to change the following inside the `bundle.json`:

```
"invocationImages": [{
    "image": "<your image>",
    "contentDigest" : "<your digest>",
		"imageType": "docker"
	}]
```

You'll want to replace `your image` and `your digest` with the correct values from our previous steps.

At this point, you could run the bundle using a CNAB tool, like `Porter` or `Duffle`. Before we do that, recreate this bundle using `Porter`.
