These are assets used by the CI build.

## Azure Pipelines

Our pipeline is broken into a few discrete builds so that we can control how and when they are triggered:

* **azure-pipelines.release.yml**: Validates canary and tag releases. This can be tested in a pull request 
  using `/azp run porter-release` though steps that require credentials will fail.
* **azure-pipelines.install.yml**: Validates our install scripts against canary and tag releases.
* **azure-pipelines.pr-automatic.yml**: Validates everything we can without a live environment.
* **azure-pipelines.pr-manual.yml**: Validates a pull request using a live environment. Requires manual triggering 
  using `/azp run porter-integration` by a maintainer because this accesses secrets in the environment, 
  e.g. kubeconfig. We don't want this accessible to anyone who submits a PR without a code review first.
 