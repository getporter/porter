---
title: Semantic Versioning
description: How Porter uses semantic versions for bundles, publishing, dependencies, and mixins
weight: 10
aliases:
  - /semver/
  - /reference/semver/
---

Porter uses [Semantic Versioning v2.0.0][semver] (semver) for bundle versions and version constraints.

- [Syntax](#syntax)
- [Bundle version](#bundle-version)
- [Publishing](#publishing)
- [Dependencies](#dependencies)
- [Mixins](#mixins)
- [Advanced usage](#advanced-usage)

## Syntax

A semantic version has the format `MAJOR.MINOR.PATCH`, with an optional prerelease and build metadata suffix:

```
MAJOR.MINOR.PATCH[-PRERELEASE][+BUILD]
```

For example `1.2.3`, `1.2.3-beta.1`, or `1.2.3+20260101`.

A leading `v` prefix is allowed, but it is not part of the version.
Porter removes it when it stores the version, so `v1.2.3` is saved as `1.2.3`.

Porter parses and compares versions with [v3 of the Masterminds/semver library][masterminds].
The "v3" refers to the version of the library, not the version of the specification.
The supported specification is always semver v2.0.0.

## Bundle version

The `version` field in porter.yaml, and the `--version` flag of `porter build`, must be a valid semantic version.

```yaml
name: mybundle
version: 1.2.3
registry: example.com/myorg
```

| Value            | Valid | Stored as        | Notes                                                    |
|------------------|-------|------------------|----------------------------------------------------------|
| `1.2.3`          | yes   | `1.2.3`          |                                                          |
| `v1.2.3`         | yes   | `1.2.3`          | The leading `v` is removed.                              |
| `1.2.3-beta.1`   | yes   | `1.2.3-beta.1`   | Prerelease version.                                      |
| `1.2.3+20260101` | yes   | `1.2.3+20260101` | Build metadata.                                          |
| `1.2`            | yes   | `1.2.0`          | Missing parts are filled with zero. Prefer `1.2.0`.      |
| `1.2.3.4`        | no    |                  | Only three version numbers are allowed.                  |
| `latest`         | no    |                  | Not a version. Use `porter publish --tag` to set a tag.  |

## Publishing

When the bundle reference does not include a tag, Porter uses the bundle version as the tag, with a `v` prefix.
OCI tags cannot contain a plus sign (`+`), so the build metadata delimiter is replaced with an underscore (`_`).

| Version          | Tag               | Bundle reference                             |
|------------------|-------------------|----------------------------------------------|
| `1.2.3`          | `v1.2.3`          | `example.com/myorg/mybundle:v1.2.3`          |
| `1.2.3-beta.1`   | `v1.2.3-beta.1`   | `example.com/myorg/mybundle:v1.2.3-beta.1`   |
| `1.2.3+20260101` | `v1.2.3_20260101` | `example.com/myorg/mybundle:v1.2.3_20260101` |

Override the version when building, or the tag when publishing:

```
porter build --version 1.2.4
porter publish --tag latest
```

`porter upgrade --version 1.2.4` follows the same convention and upgrades to the bundle tagged `v1.2.4`.

## Dependencies

The `version` field of a dependency is a version constraint.
It describes which versions of the dependency bundle are acceptable.

```yaml
dependencies:
  requires:
    - name: mysql
      bundle:
        reference: example.com/mysql:v1.2.3
        version: ">=1.2.0, <2.0.0"
```

| Constraint           | Matches                                   |
|----------------------|-------------------------------------------|
| `1.2.3`              | Exactly `1.2.3`                           |
| `>=1.2.0`            | `1.2.0` or higher                         |
| `>=1.2.0, <2.0.0`    | `1.2.0` or higher, but lower than `2.0.0` |
| `1.2.x` or `1.2.*`   | `>=1.2.0, <1.3.0`                         |
| `~1.2.3`             | `>=1.2.3, <1.3.0`                         |
| `^1.2.3`             | `>=1.2.3, <2.0.0`                         |
| `2.*`                | `>=2.0.0, <3.0.0`                         |
| `1.x \|\| 3.x`       | Either `1.x` or `3.x`                     |

Porter reads the tags of the dependency repository, and only considers tags that are semantic versions, such as `v1.2.3`.
Prerelease versions are not selected when resolving a constraint.
Constraints are only resolved when the `dependencies.version-strategy` configuration setting is not `exact`.
See [Version Ranges](/docs/development/authoring-a-bundle/working-with-dependencies/#version-ranges) for how Porter selects a version.

## Mixins

A mixin declaration can include a version constraint after an `@`.
`porter lint` and `porter build` report an error if the installed mixin does not satisfy the constraint.

```yaml
mixins:
  - exec@1
  - helm3@~1.2
  - terraform@>=2
```

## Advanced usage

The examples above cover the most common cases.
See the [Masterminds/semver documentation][constraints] for the full constraint syntax.

[semver]: https://semver.org/spec/v2.0.0.html
[masterminds]: https://github.com/Masterminds/semver
[constraints]: https://github.com/Masterminds/semver#checking-version-constraints
