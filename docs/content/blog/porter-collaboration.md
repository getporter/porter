---
title: "Managing Bundles as a Team with Porter"
description: "Porter's latest release makes managing a bundle's lifecycle as a team possible."
date: "2020-02-19"
authorname: "Carolyn Van Slyck"
author: "@carolynvs"
authorlink: "https://carolynvanslyck.com/"
authorimage: "https://github.com/carolynvs.png"
image: "images/porter-twitter-card.png"
tags: ["plugins"]
---

One tenet of working on a great team, is that no one deploys alone. I may have
originally deployed our application, but that doesn't mean that I am responsible
for patching it until the end of time, or until I run away screaming. I am
excited to announce that Porter now fully supports teams managing a bundle
collaboratively! 🙌

🔷 Sally deploys the application on Tuesday using secrets stored in the team's
Key Vault. 

🔷 When a new version of the application comes out on Thursday, Adnan updates
the deployed bundle. 

🔷 Their manager Qi downloads Porter onto her laptop over the weekend and in
minutes is running a custom bundle action to view logs to investigate a bug
report.

These seamless handoffs are now possible because of two big efforts: Porter's
new [plugin framework][plugins] and changes to [cnab-go][cnabgo] supporting
generic storage and credential resolution strategies.

Now your team can setup a cloud account, and share a Porter
config file that says which plugin to use and how to connect to the account.
Porter uses the plugin to resolve credentials against the team's secret store,
such as Azure Key Vault, and stores the bundle instance in cloud
storage, like Azure Blob Storage.

This is a big step forward for collaborating on bundles. More importantly it is
**much more secure**. Secrets plugins move the storage of credentials used by
your bundles off of laptops and CI machines, back into secure secret stores
where they can be managed by the team, encrypted at rest and not left around
after the bundle is executed. Even if you are using Porter as a single user, you
should move to using a secrets plugin.

With the [latest release][release] of Porter, the [Azure plugin][azure-plugin]
is installed by default so that people can try it out. Nothing about the plugin framework is
specific to Azure, it is just the first plugin we implemented. We would love to
see more plugins for other providers! Just like with mixins, anyone can write a
plugin, distribute, and list it alongside the porter-authored plugins. Please
reach out to us on the [#porter Slack][slack] if you are interested in making a
plugin.

😎 Already using Porter? [Install][install] the latest release, give the
[plugins tutorial][tutorial] a try and let us know what you think!

🎉 Ready to try Porter for the first time? [Install][install], head over to the
quickstart and then check out our [learning][learning] page for a high level
overview of CNAB, demos of bundles in action and more.

[release]: https://github.com/deislabs/porter/releases/tag/v0.23.0-beta.1
[plugins]: /plugins/
[cnabgo]: https://github.com/cnabio/cnab-go/
[azure-plugin]: /plugins/azure/
[slack]: /community/#slack
[install]: /install/
[tutorial]: /plugins/tutorial/
[learning]: /learning/
