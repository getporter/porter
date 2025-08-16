
---
title: "Quickly set up a Porter environment with required plugins"
description: "How to install multiple plugins with Porter"
date: "2023-01-24"
authorname: "Yingrong Zhao"
author: "@vinozzz"
authorlink: "https://github.com/vinozzz"
authorimage: "https://github.com/vinozzz.png"
tags: ["best-practice", "plugins"]
summary: | 
    Setting up your Porter environment with your required plugins using the new `--file` flag with `porter plugins install` command.
---

### Breaking change
The recent porter v1.0.5 release introduced a new flag `--file` on `porter plugins install` command. Its intention is to allow users to install multiple plugins through a plugins definition file with a single porter command. However, it did not work as expected due to bad file format.

The fix that contains the correct schema has been published with a new v1.0.6 release. If you have an existing plugins file, please update it to work with v1.0.6+.

### Install multiple plugins with a single command
Now, you can install multiple plugins using a plugin definition yaml file like below:
```yaml
schemaType: Plugins
schemaVersion: 1.0.0
plugins:
  azure:
    version: v1.0.1
  kubernetes:
    version: v1.0.1
```

After creating the file, you can run the command:
```bash
porter plugins install -f <path-to-the-file>
```

The output from the command should look like this:
```
installed azure plugin v1.0.1 (e361abc)
installed kubernetes plugin v1.0.1 (f01c944)
```

Make sure to update your current plugins schema file to the [latest format](/docs/references/file-formats/plugins/) 
Please [let us know][contact] how the change went (good or bad), and we are happy to help if you have questions, or you would like help with your migration.

[announced]: https://github.com/docker/roadmap/issues/209
[Install Porter]: /install/
[contact]: /community/