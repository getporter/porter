These are assets used by the CI build.

* testdata - test data used for the integration tests, such as credentials and test bundles.
* images - source for any docker images that we publish. 

## Azure Pipelines

See all Porter related [pipelines across all repositories](https://dev.azure.com/getporter/porter/_build).

Our pipeline is broken into a few discrete builds so that we can control how and when they are triggered:

* **porter-release: azure-pipelines.release.yml**: Publishes tagged releases. Does not run tests.
  This can be tested in a pull request using `/azp run test-porter-release`.
  [View Latest Builds](https://dev.azure.com/getporter/porter/_build?definitionId=2)
* **porter-canary: azure-pipelines.canary.yml**: Validates canary releases, running the full test suite.
  [View Latest Builds](https://dev.azure.com/getporter/porter/_build?definitionId=26)
* **porter-check-install: azure-pipelines.install.yml**: Validates our installation scripts against canary and tag releases.
  [View Latest Builds](https://dev.azure.com/getporter/porter/_build?definitionId=3)
* **porter: azure-pipelines.pr-automatic.yml**: Validates everything we can without a live environment.
  [View Latest Builds](https://dev.azure.com/getporter/porter/_build?definitionId=18)
* **porter-integration: azure-pipelines.integration.yml**: Runs the full integration test suite.
  Only runs when requested with `/azp run porter-integration` because they are slow.
  [View Latest Builds](https://dev.azure.com/getporter/porter/_build?definitionId=25)

### Documentation Only Builds

The `porter` and `porter-integration` builds [detect if the build is
documentation only](doc-only-build.sh) and will short circuit the build. They
look for changes to the website, markdown files, workshop materials and
repository metadata files that do not affect the build.