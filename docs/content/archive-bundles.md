---
title: Archiving Bundles
description: Archiving Bundles with Porter
---

Porter allows you to share bundles by [publishing](/distribute-bundles) them to an OCI registry. Porter also allows you to copy a bundle from one registry to another. Using these commands, bundle users have flexibility in how they leverage published bundles. What if you want to use a published bundle on a disconnected or edge network that has limited connectivity? The `porter archive` command and the `porter publish` commands allow you to take the bundle from a registry on one network, move it to the network or location, and republish it into another registry for use within that environment. The generated bundle archive contains the CNAB `bundle.json`, along with an OCI [image layout](https://github.com/opencontainers/image-spec/blob/master/image-layout.md) containing the invocation image and any images declared in the `images` section of the `bundle.json`. This enables the entire bundle to be easily moved into a private data center or across an air-gapped network, and republished within that environment.

For a working example of how to move a bundle across an airgap, read [Example: Airgapped Environments](/examples/airgap/).

## Generating a Bundle Archive With Porter

In order to generate the archive, all of the images in the bundle **must** have been published to a registry. For this reason, you must first `publish` your bundle to a registry:

```
porter publish --reference jeremyrickard/porter-do-bundle:v0.5.0
```

Now you can run the `porter archive` command and designate the archive file name and bundle tag to use:

```
porter archive --reference jeremyrickard/porter-do-bundle:v0.5.0 do-porter.tgz
```

This will generate a file in the directory named `do-porter.tgz`.

## Bundle Archive Format

The generated bundle archive is a CNAB [thick bundle](https://github.com/cnabio/cnab-spec/blob/master/104-bundle-formats.md#formatting-and-transmitting-thick-bundles). Once you have a bundle archive, you can use the `tar` command to examine the contents. If we examine the `do-porter.tgz` generated above, we would see:

```
$ tar tvf do-porter.tgz
drwx------  0 jeremyrickard staff       0 Oct 18 10:27 ./
drwxr-xr-x  0 jeremyrickard staff       0 Oct 18 10:27 ./artifacts/
drwxr-xr-x  0 jeremyrickard staff       0 Oct 18 10:27 ./artifacts/layout/
drwxr-xr-x  0 jeremyrickard staff       0 Oct 18 10:27 ./artifacts/layout/blobs/
drwxr-xr-x  0 jeremyrickard staff       0 Oct 18 10:27 ./artifacts/layout/blobs/sha256/
-rw-r--r--  0 jeremyrickard staff 13488110 Oct 18 10:27 ./artifacts/layout/blobs/sha256/2ad9e0f64b8e37364f849ea7a19a7e405bf9fea6905cfcdfe4e4d796e2170e24
-rw-r--r--  0 jeremyrickard staff     3952 Oct 18 10:27 ./artifacts/layout/blobs/sha256/3a065bb24678891307c08fa9555f76d7465afdfd22f7287547be950d270d3205
-rw-r--r--  0 jeremyrickard staff 26855472 Oct 18 10:27 ./artifacts/layout/blobs/sha256/642b6ba3f53cf3870d3597fc53fe25af87baa6b089f02258e0078834c7723cf1
-rw-r--r--  0 jeremyrickard staff 43658711 Oct 18 10:27 ./artifacts/layout/blobs/sha256/6e765adb926b303b0cd55c6984396e8124ffe1496e7eef559c75d3ddafa39693
-rw-r--r--  0 jeremyrickard staff     2223 Oct 18 10:27 ./artifacts/layout/blobs/sha256/74b8622a8b7f09a6802a3fff166c8d1827c9e78ac4e4b9e71e0de872fa5077be
-rw-r--r--  0 jeremyrickard staff 19643575 Oct 18 10:27 ./artifacts/layout/blobs/sha256/814a8fb9e6004c9cfa19f5a19e4d3d147219cb8a01ca1d4a3f78099e9181a106
-rw-r--r--  0 jeremyrickard staff     1159 Oct 18 10:27 ./artifacts/layout/blobs/sha256/8f1133d81f1b078c865cdb11d17d1ff15f55c449d3eecca50190eed0f5e5e26f
-rw-r--r--  0 jeremyrickard staff 45380663 Oct 18 10:27 ./artifacts/layout/blobs/sha256/9a0b0ce99936ce4861d44ce1f193e881e5b40b5bf1847627061205b092fa7f1d
-rw-r--r--  0 jeremyrickard staff     5067 Oct 18 10:27 ./artifacts/layout/blobs/sha256/a121f14b1e214c94de2f6b3ab5dcab0f25a5bdc52bc33e7fed506468f5f516bd
-rw-r--r--  0 jeremyrickard staff 50445916 Oct 18 10:27 ./artifacts/layout/blobs/sha256/ab8d0c1aab8051838635f6f77a91bd2bc91dfb88a2e607ae3bff4d35bdf9c9a9
-rw-r--r--  0 jeremyrickard staff 70732768 Oct 18 10:27 ./artifacts/layout/blobs/sha256/c2274a1a0e2786ee9101b08f76111f9ab8019e368dce1e325d3c284a0ca33397
-rw-r--r--  0 jeremyrickard staff 25952905 Oct 18 10:27 ./artifacts/layout/blobs/sha256/c251f693876471e137e1114afa83d0db9a45466755ab298ee9175660b56cd73b
-rw-r--r--  0 jeremyrickard staff 50510171 Oct 18 10:27 ./artifacts/layout/blobs/sha256/c384021f2cd9fe8e1062f3b0567c0bd9e53cc9f9c69374615ca9bcf505c92ada
-rw-r--r--  0 jeremyrickard staff  2757034 Oct 18 10:27 ./artifacts/layout/blobs/sha256/e7c96db7181be991f19a9fb6975cdbbd73c65f4a2681348e63a141a2192a5f10
-rw-r--r--  0 jeremyrickard staff   621589 Oct 18 10:27 ./artifacts/layout/blobs/sha256/f3d4cde0abb9067fefe5d748b046b4c0fdc6dcfa3829edf72589d960602cca4a
-rw-r--r--  0 jeremyrickard staff 50275925 Oct 18 10:27 ./artifacts/layout/blobs/sha256/f4a430fcc4abda7ae7d7161fb041d73cd2a9960a1822d67254ba839c96722f90
-rw-r--r--  0 jeremyrickard staff      238 Oct 18 10:27 ./artifacts/layout/blobs/sha256/f910a506b6cb1dbec766725d70356f695ae2bf2bea6224dbe8c7c6ad4f3664a2
-rwxr-xr-x  0 jeremyrickard staff      798 Oct 18 10:27 ./artifacts/layout/index.json
-rwxr-xr-x  0 jeremyrickard staff       37 Oct 18 10:27 ./artifacts/layout/oci-layout
-rw-r--r--  0 jeremyrickard staff     2955 Oct 18 10:27 ./bundle.json
```

In this archive file, you will see the `bundle.json`, along with all of the artifacts that represent the OCI image layout. In this case, we had two images, the invocation image and an application image. They are both written to the `artifacts/` directory as part of the OCI image layout.

## Publish a Bundle Archive

Once you have a bundle archive, the next step to make it usable is to publish it to an OCI registry. To do this, the `porter publish` command is used. Given our `do-porter.tgz` bundle above, we can publish this to a new registry with the following command:

```
$ porter publish -a do-porter.tgz --reference jrrporter.azurecr.io/do-porter-from-archive:1.0.0
Starting to copy image jrrporter.azurecr.io/do-porter-from-archive/porter-do@sha256:74b8622a8b7f09a6802a3fff166c8d1827c9e78ac4e4b9e71e0de872fa5077be...
Completed image jrrporter.azurecr.io/do-porter-from-archive/porter-do@sha256:74b8622a8b7f09a6802a3fff166c8d1827c9e78ac4e4b9e71e0de872fa5077be copy
Starting to copy image jrrporter.azurecr.io/do-porter-from-archive/spring-music@sha256:8f1133d81f1b078c865cdb11d17d1ff15f55c449d3eecca50190eed0f5e5e26f...
Completed image jrrporter.azurecr.io/do-porter-from-archive/spring-music@sha256:8f1133d81f1b078c865cdb11d17d1ff15f55c449d3eecca50190eed0f5e5e26f copy
Bundle tag jrrporter.azurecr.io/do-porter-from-archive:1.0.0 pushed successfully, with digest "sha256:1da3221ca38890d987791192e25f2634e195606f7d72bb2fea39c2865f503175"
```

This command will expand the bundle archive and copy each image up to the new registry. Once complete, you can use the bundle like any other published bundle:

```
porter explain jrrporter.azurecr.io/do-porter-from-archive:1.0.0
Name: spring-music
Description: Run the Spring Music Service on Kubernetes and Digital Ocean PostgreSQL
Version: 0.5.0

Credentials:
Name               Description                                       Required
do_access_token    Access Token for Digital Ocean Account            true
do_spaces_key      DO Spaces API Key                                 true
do_spaces_secret   DO Spaces API Secret                              true
kubeconfig         Kube config file with permissions to deploy app   true

Parameters:
Name            Description                                                     Type      Default             Required   Applies To
database_name   Name of database to create                                      string    jrrportertest       false      All Actions
helm_release    Helm release name                                               string    spring-music-helm   false      All Actions
namespace       Namespace to install Spring Music app                           string    default             false      All Actions
node_count      Number of database nodes                                        integer   1                   false      All Actions
region          Region to create Database and DO Space                          string    nyc3                false      All Actions
space_name      Name for DO Space                                               string    jrrportertest       false      All Actions

Outputs:
Name         Description                                Type     Applies To
service_ip   IP Address assigned to the Load Balancer   string   All Actions
```

## Next Steps

* [Example: Airgapped Environments](/examples/airgap/)