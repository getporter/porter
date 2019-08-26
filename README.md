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

_2019/08/26 Wait where did my summer go?_ 🌦

Porter could go in lots of directions! Here are our top goals right now:

1. CNAB Specification Compliance - Milestone [CNAB 1.0](https://github.com/deislabs/porter/milestone/12)

    The CNAB Core 1.0 spec has been frozen.
    
    Progress Report: 99% there! We have full support for CNAB Core 1.0, just dotting some i's and crossing those t's.
    
1. Dependency Distribution - Milestone [Dependencies](https://github.com/deislabs/porter/milestone/8)

    Solve end-to-end how bundle authors use porter to build, publish and then use someone's bundle as a dependency.
    
    Progress Report: 75% there! You can make a bundle with a dependency, give it a try. 👍

1. Thick Bundles - Milestone [Thick and Relocated Bundles](https://github.com/deislabs/porter/milestone/13)

    CNAB thick bundles are used for transmitting across air-gapped networks.
