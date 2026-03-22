---
title: How to reference images in your bundle
description: When and how to use the images section to control where your bundle finds its images
weight: 2
---

Bundles often deploy workloads that require container images.
Porter gives you two ways to handle this: hardcoded image references or declaring
images in the manifest's `images` section.
Choosing the right approach depends on whether you need Porter to manage image
relocation for you.

## Hardcoded image references

The simplest approach is to reference images directly in your bundle steps, just
like you would in any shell script or Helm chart:

```yaml
install:
  - helm3:
      description: "Install my app"
      name: myapp
      chart: ./charts/myapp
```

In this case the chart controls which image to pull and where to pull it from.
Porter has no knowledge of those images — it will not copy them when the bundle
is published or archived, and it will not update the references when the bundle
is moved to a different registry.

This approach is fine when:

- The bundle is only ever run in environments that have direct access to the
  original image registry (e.g., Docker Hub, ghcr.io).
- You are iterating quickly and want to avoid the overhead of declaring digests.
- The images are managed by a third-party Helm chart that already controls its
  own image references and doesn't support overriding them.

## Declaring images in the manifest

When you add images to the `images` section of `porter.yaml`, Porter tracks them
and manages their location for you:

```yaml
images:
  myapp:
    description: "My application image"
    imageType: "docker"
    repository: "myregistry.example.com/myorg/myapp"
    digest: "sha256:abc123..."
```

You then reference the image in your bundle steps using template variables
instead of a hardcoded string:

```yaml
install:
  - helm3:
      description: "Install my app"
      name: myapp
      chart: ./charts/myapp
      set:
        image.repository: ${ bundle.images.myapp.repository }
        image.digest: ${ bundle.images.myapp.digest }
```

Use `digest` rather than `tag` so that the reference remains stable and
immutable across environments.

### How relocation works

When you run `porter publish`, Porter copies every image declared in the
`images` section into the same repository as the bundle itself.
At runtime, `${ bundle.images.myapp.repository }` and
`${ bundle.images.myapp.digest }` resolve to the relocated image, not the
original source.

```
Original:  myregistry.example.com/myorg/myapp@sha256:abc123...
Published: myregistry.example.com/myorg/mybundle@sha256:abc123...
```

This means the bundle and all its images travel together as a unit — no
external registry access is needed at runtime.

The same applies when you run `porter archive`: the bundle archive (tgz file)
contains all declared images so they can be moved to a disconnected environment
and re-published there.

See [Image References After Publishing] for a deeper walkthrough.

[Image References After Publishing]: /docs/development/authoring-a-bundle/distribute-bundles/#image-references-after-publishing

## Which approach to choose

| Scenario | Approach |
|---|---|
| Need to deploy in an [airgapped environment] | Declare images in `images` section |
| Moving bundles between registries | Declare images in `images` section |
| Images managed by a third-party chart | Hardcoded (or skip if chart doesn't support overriding) |
| Fast iteration / local testing | Hardcoded is fine; add `images` section before publishing |
| Bundle always runs against the same registry | Either works |

[airgapped environment]: /docs/administration/move-bundles-airgapped/

## Next steps

- [Images section in the manifest reference](/docs/bundle/manifest/#images)
- [Example: Airgapped Environments](/docs/references/examples/airgap/)
- [Archiving bundles](/archive-bundles/)
- [Copying bundles between registries](/docs/administration/copy-bundles/)
