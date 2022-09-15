# Reviewing Guide

* [Values](#values)
* [How to review a pull request](#how-to-review-a-pull-request)
  * [Giving feedback](#giving-feedback)
* [Cut a release](#cut-a-release)


# Values

Thank you for reviewing pull requests for us! 💖

Here are the values and etiquette that we follow to ensure a welcoming, inclusive
project that doesn't burn out our authors or our reviewers. 😅

* We ask that authors respect reviewers' time. Check out the
  [Contributing Guide](CONTRIBUTING.md) and know that you can ask the
  author to do their part to make _your_ part manageable.
* We ask that reviewers respect authors' time. Please do your best to review
  a pull request in a reasonable amount of time once you have assigned it to
  yourself.
* The definition of "reasonable amount of time" is 3 business days. The ask is
  that after each trigger: "Review Requested", "Changes Incorporated", etc that
  the other person attempt to do their part within 3 business days. If they
  can't, please leave a comment and let the other person know that it will take
  longer. If life comes up, let others know that you need to unassign yourself
  and someone else will complete the review.

# How to review a pull request

1. Do not start reviewing a pull request if it is in WIP or is a draft pull
   request. Wait until they have marked it ready for review.
1. Assign yourself to the pull request. This gives the author feedback that
   someone is going to do the review, while giving you time to do the review.
1. Do a quick check for areas that need to be addressed before the pull request
   can be reviewed.
   
   For example, it is missing an agreed upon solution, requires an explanation
   from the author, has a very large set of changes that are not easy to review,
   etc., ask the author to correct that up-front.
1. When you provide feedback, make it clear if the change must be made
   for the pull request to be approved, or if it is just a suggestion. Mark
   suggestions with **nit**, for example, `nit: I prefer that the bikeshed be
   blue`.
1. When the pull request is ready to merge, squash the commits they require
   tidying unless the author asked to do that themselves.

See [The life of a pull request](CONTRIBUTING.md#the-life-of-a-pull-request) for 
what we expect a pull request to feel like for everyone involved.

## Merge Requirements

* Unit Tests
* Documentation Updated
* Passing CI

When a pull request impacts code, i.e. it's not a documentation-only change,
the reviewer should run the manual integration tests after reviewing the code.
The tests are triggered with a comment:

```
/azp run porter-integration
```

[Admins][admins] are allowed, at their discretion, to merge administrative pull
requests without review and before the full CI suite has passed. This is
sometimes used for typo fixes, updates to markdown files, etc. This is a
judgement call based on the type of change, risk, and availability of other
reviewers.

## Giving feedback

* Be kind. Here is [good article][kind-reviews] with example code reviews and 
  how to improve your feedback. Giving feedback of this caliber is a requirement 
  of maintainers and those who cannot do so will have the maintainer role revoked.
* Request changes for bugs and program correctness.
* Request changes to be consistent with existing precedent in the codebase.
* Request tests and documentation in the same pull request.
* Prefer to optimize for performance when necessary but not up-front without
  a reason.
* Prefer [follow-on PRs](CONTRIBUTING.md#follow-on-pr).
* Do not ask the author to write in your style.

[kind-reviews]: https://product.voxmedia.com/2018/8/21/17549400/kindness-and-code-reviews-improving-the-way-we-give-feedback

# Cut a Release

🧀💨

Our CI system watches for tags, and when a tag is pushed, cuts a release
of Porter. When you are asked to cut a new release, here is the process:

1. Figure out the correct version number using our [version strategy].
    * Bump the major segment if there are any breaking changes, and the 
      version is greater than v1.0.0
    * Bump the minor segment if there are new features only.
    * Bump the patch segment if there are bug fixes only.
    * Bump the pre-release number (version-prerelease.NUMBER) if this is
      a pre-release, e.g. alpha/beta/rc.
1. First, ensure that the main CI build has already passed for 
    the [commit that you want to tag][commits], and has published the canary binaries. 
    
    Then create the tag and push it:

    ```
    git checkout main
    git pull
    git tag VERSION -a -m ""
    git push --tags
    ```
    If the CI build failed to build for the release, fix the problem first. Then increment the PATCH version, e.g. v0.7.0->v0.7.1, and go through the above steps again to publish the binaries. It's often a good pratice to finish the release first before updating any of our docs that references the latest release.

1. Generate some release notes and put them into the release on GitHub.
   - Go to Porter Github repository and find the newly created release tag. You should see a
   "auto generate release notes" button to create release notes for the release.
   - Modify the generated release note to call out any breaking or notable changes in the release.
   - Include instructions for installing or upgrading to the new release:
    ```
      # Install or Upgrade
      Run (or re-run) the installation from https://getporter.org/install to get the
    latest version of porter.
    ```
1. Announce the new release in the community.
   - Email the [mailing list](https://getporter.org/mailing-list) to announce the release. In your email, call out any breaking or notable changes.
   - Post a message in [Porter's slack channel](https://getporter.org/community/#slack).
1. If there are any issues fixed in the release and someone is waiting for the fix, comment on the issue to let them know and link to the release.
1. If the release contains new features, it should be announced through a [blog](https://getporter.org/blog/) post and on Porter's twitter account.

[maintainers]: https://github.com/orgs/getporter/teams/maintainers
[admins]: https://github.com/orgs/getporter/teams/admins
[commits]: https://github.com/getporter/porter/commits/main
[version strategy]: https://getporter.org/project/version-strategy/
