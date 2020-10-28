# Discourse Example Bundle
This example is a work in progress and not finished.

## Why Discourse?
[Discourse](https://www.discourse.org/about) is the perfect example bundle because it will illustrate the need for bundling your application and using Porter. Discourse is a platform for discussion with many exciting [features](https://www.discourse.org/features).

 The installation and setup process for Discourse is deceptively simple at first glance. However, it is actually extremely complicated. Although there is a standalone docker image that you can run, there is still a lot of infrastructure that needs to be configured, especially if you wish to use Discourse in production and want to take advantage of all the features that are possible. 
 
 Because of this, you cannot just run the docker image and do a simple install. You will have to go through the complex process in order to set up the infrastructure and customize your Discourse. 
 
 In order to install, first, you need to create a cloud server and access it. Then, you need to install docker and install discourse. To get email working, you will need to set up a mail server and get the credentials. You also need a domain name. 
 
 Then, you will launch a discourse set up tool and answer questions about your hostname, email address, server address, etc. After Discourse is up and running, if an update needs to be made or maintenance is needed, another process is required. 
 
 Additionally, if you want more features such as single-sign-on, plugins, or encryption, those need to be configured as well. Because of all these complex steps that are needed to install Discourse and upgrade it, putting Discourse in a bundle would greatly simplify the process. 
 
 Bundling Discourse would take away the need for individuals installing it to read through all the installation instructions. Users would simply need to enter their credentials and any parameters and Porter would do the work for them. Discourse would be installed with just one command.

## What bundle should do and look like (parameters and credentials etc)
install:
* cloud server (DigitalOcean, Azure, etc)
    - need credentials 
* cloud storage for user uploads, pictures (Amazon S3, Azure Blob Storage, etc)
    - can set up backups 
    - need credentials
* email (Mailgun, SendGrid, Mailjet)
    - can configure reply via email
    - need credentials and email as parameter
* domain name
    - hostname parameter
* ssl certificate (Let’s Encrypt free certificate) 
    - need certificate and key
    - have to configure NGINX and a docker container
* Virtual Machine
    - need credentials
* postgres database
    - need parameters for username, password, database name
* configure SSO
    - enable_sso parameter must be enabled
    - sso_url: the offsite URL users will be sent to when attempting to log on
    - sso_secret: a secret string used to hash SSO payloads
* login via Google, Twitter, GitHub, Facebook
* install plugins
    - need plugin's git clone url 
    - add it to app.yml
    - rebuild the container
* multisite configuration with docker
    - if you want to host multiple Discourse sites on the same server
* set up webhooks
* enable a CDN 
    - origin address
    - CNAME
    - CDN URL
    - need to edit DNS map to map CNAME to CDN URL

upgrade:
- have exec run a script that will do the upgrade (probably run rebuild)

potential custom actions from launcher:
* start:      Start/initialize a container
* stop:       Stop a running container
* restart:    Restart a container
* destroy:    Stop and remove a container
* enter:      Use nsenter to get a shell into a container
* logs:       View the Docker logs for a container
* bootstrap:  Bootstrap a container for the config based on a template
* rebuild:    Rebuild a container (destroy old, bootstrap, start new)
* cleanup:    Remove all containers that have stopped for > 24 hours

uninstall:
* delete azure storage account
* delete azure storage container
* delete azure vm
* delete postgres database
