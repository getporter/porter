# Porter's CDN Setup

Porter uses a variety of services to distribute its CLI, mixin and plugin binaries.

## GitHub Releases

Binaries and install scripts are attached to GitHub releases upon relevant events in our GitHub repositories, such as merges to the "main" branch (producing `canary` artifacts) and official releases (producing `latest` and semver-tagged artifacts). We create releases for the following tags:

* vX.Y.Z - a tagged version of Porter. Recommended.
* latest - the most recent tagged version of Porter. Stable but you should use a specific version in prod.
* canary - the tip of the main branch. Unstable.

We use our own tag for latest (instead of relying upon GitHub's latest release logic) so that we have consistent URLs.

## GitHub Packages Repo

We have a [packages](https://github.com/getporter/packages) repository that has our official mixin and plugin atom feeds used by porter mixin install and porter plugin install. It also contains an index of all known mixins and plugins from both the Porter Authors and the community, which is used by porter mixins search and porter plugins search.

## DNS

DNS entries for the `porter.sh` and `getporter.org` domains are managed via a [Netlify](https://www.netlify.com/) account.  A `cdn` CNAME record exists in this configuration such that `cdn.porter.sh` can represent the hostname for all artifact URLs.

Netlify handles redirecting traffic to cdn.porter.sh to the appropriate backing store (either a GH release artifact or a file in a repo).
