<img align="right" src="docs/static/images/porter-notext.png" width="125px" />

[![Build Status](https://dev.azure.com/cnlabs/porter/_apis/build/status/deislabs.porter?branchName=master)](https://dev.azure.com/cnlabs/porter/_build/latest?definitionId=6?branchName=master)


# Porter is a Cloud Installer

Porter gives you building blocks to create a cloud installer for your application, handling all the
necessary infrastructure and configuration setup. It is a declarative authoring experience that lets you
focus on what you know best: your application.

Want to start using Porter? Check out the [QuickStart Guide](https://porter.sh/quickstart/) for a brief walkthrough.

Learn more at [porter.sh](https://porter.sh)

---

_Want to work on Porter with us? See our [Contributing Guide](CONTRIBUTING.md)_

---

## Roadmap

_2019/03/14 pi day_ 🥧

Porter go in lots of directions! Here are our top 4 goals at the moment:

1. Use Porter without installing Duffle - Milestone [Look Ma, No Duffle](https://github.com/deislabs/porter/milestone/3)

    Compile duffle functionality into porter as needed, instead of having the user switch between the two CLIs.

2. Dependency Distribution - Milestone TBC

    Solve end-to-end how bundle authors use porter to build, publish and then use someone's bundle as a dependency.

3. Mixin Distribution - Milestone TBC

    Make it easy for anyone to create and distribute mixins that porter can discover and install.

4. CNAB Specification Compliance - Milestone TBC

    As the [CNAB specification](https://github.com/deislabs/cnab-spec) moves toward 1.0, update Porter to be compliant with the spec. Currently, if you build a bundle with Porter, you'll be able to install it with Porter. There are some gaps with the spec that limit compatibility with other CNAB tooling. See the [CNAB 1.0 Milestone](https://github.com/deislabs/porter/milestone/12) for more information on these gaps.
