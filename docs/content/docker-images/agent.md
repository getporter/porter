---
title: Porter Agent Docker Image
description: The Porter Operator agent image
---

The [getporter/porter-agent][porter-agent] Docker image is intended for use by
the [Porter Operator]. If you need to run Porter in a container, you should use
the [porter client] image.

It has tags that match what is available from our [install](/install/) page:
latest, canary and specific versions such as v0.20.0-beta.1.

The [configuration] file for Porter should be mounted in a volume to **/porter-config**.
The image will copy the configuration file into PORTER_HOME when the container starts
and then run the specified porter command, similar to the [porter client] image.

[configuration]: /configuration
[porter-agent]: https://hub.docker.com/r/getporter/porter-agent/tags
[porter client]: /docker-images/client/
[Porter Operator]: https://github.com/getporter/operator
