---
title: Move a bundle across an airgap
description: How to deploy in an airgapped or disconnected environment
---

An airgapped environment is an environment that doesn't have full access to common networks such as the internet and as such some actions such as pulling Docker images from Docker Hub, or downloading a build artifact, may not be possible.
These disconnected environments are common in secure environments that handle sensitive data.
Porter can help deploy across an airgap by including everything that your bundle needs to deploy, including Docker images referenced by the bundle.

At a high level, this involves the following steps:

1. Archive the bundle to a compressed file (tgz).
2. Move the archive across the airgap on physical media such as a disc or USB drive.
3. Publish the bundle to a registry on the other side of the airgap.
4. Install the bundle referencing the new location of the bundle inside the airgapped network.

<figure>
    <img src="/administrators/porter-airgap-publish.png" alt="a drawing showing two networks side by side, separated by an airgap. Network A has a docker registry with a copy of a bundle that includes the bundle.json, installer and the whalesayd image. An arrow labeled porter archive leads to a box with all those components in a single box labeled whalegap.tgz. Then another arrow labeled with a disc goes across the airgap, copying the same whalegap.tgz box wit its components into Network B. Then a final arrow labeled porter publish puts a copy of the bundle and its contents in Registry B, inside Network B."/>
    <figcaption>Moving the whalegap bundle across an airgap</figcaption>
</figure>

## Next Steps
* [Example: Airgapped Environments](/examples/airgap/)
* [Understand how Porter publishes archived bundles to a registry](/archive-bundles/)
