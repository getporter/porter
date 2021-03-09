---
title: Porter and the CNAB Specification
description: Porter implements the CNAB Specification and can run bundles created by other tools
---

Porter was designed so that you do not need to know anything about the [Cloud Native Application
Bundle (CNAB) Specification][cnab]. However, if you are interested in implementing your
own internal CNAB tool like Porter, or use different tools such as Docker App
alongside Porter, then it quickly becomes important to understand what the
specification means, and the guarantees that it provides.

The CNAB Specification is composed of multiple specifications:

* [CNAB Core] ✅ - Covers bundle execution. Porter supports version 1.1.0 of the
  specification.

* [CNAB Registry] ✅ - Covers storing bundles in OCI registries. The spec is still
  in DRAFT (but has not changed in quite a while) and Porter supports this
  specification.

* [CNAB Security] ❌ - Covers supply-chain security for bundles. The spec is still 
  in DRAFT, may have more changes incoming and is not yet supported by Porter.
  You can use the [signy] tool to sign and verify bundles according to the draft
  specification.

* [CNAB Claims] ✅ - Covers data storage, such as what is installed and the status
  of the installation. This allows tools to support reading and writing to other
  tools data because they follow a standard format. Porter supports version 1.0.0
  of the specification.

* [CNAB Dependencies] ✅ - Covers bundles depending upon other bundles, for example
  WordPress depending upon a MySQL bundle. The spec is still in DRAFT and has
  major changes planned. Porter suports the DRAFT and serves as the reference
  implementation of the spec, vetting proposals to this spec before they are adopted.

* [Credential Sets] ✅ - Covers managing credential sets on the client-side and passing
  them to a bundle. This is a non-normative spec, so it doesn't have a release status.
  Porter supports not only credential sets, but also parameter sets.

* [Parameter Sources] ✅ - Covers how a bundle can connect its outputs to its own
  parameters for subsequent actions. For example, a bundle may output a
  connection string during install, and then use that output as a parameter for
  all other actions in the bundle without the user having to specify it.

Any tool that says that they are CNAB-compliant are saying that they at least
implement the [CNAB Core] Specification. When deciding how to run bundles
created by one tool with a different tool, you will want to understand what
specifications the bundle uses and if they are supported by each tool.

Porter is committed to both generating CNAB compliant bundles that can be run by
other CNAB-compliant tools, and running bundles created by other tools. Porter
usually implements an experimental feature before it is adopted by the
specification, so there may be times when a bundle created by Porter won't work
with another tool. Also depending on the roadmap and release cadence of the
other tool, a feature may be in the spec, but the other tool hasn't yet
implemented support for it.

When running bundles from unknown sources, we recommend using the tool with the
broadest specification support, such as Porter, to execute the bundles.

[cnab]: https://cnab.io
[CNAB Core]: https://github.com/cnabio/cnab-spec/blob/main/100-CNAB.md
[CNAB Registry]: https://github.com/cnabio/cnab-spec/blob/main/200-CNAB-registries.md
[CNAB Security]: https://github.com/cnabio/cnab-spec/blob/main/300-CNAB-security.md
[CNAB Claims]: https://github.com/cnabio/cnab-spec/blob/main/400-claims.md
[signy]: https://github.com/cnabio/signy
[CNAB Dependencies]: https://github.com/cnabio/cnab-spec/blob/main/500-CNAB-dependencies.md
[Credential Sets]: https://github.com/cnabio/cnab-spec/blob/main/802-credential-sets.md
[Parameter Sources]: https://github.com/cnabio/cnab-spec/blob/main/810-well-known-custom-extensions.md#parameter-sources
