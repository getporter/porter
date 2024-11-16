# Contribution Ladder

---
* [Roles](#roles)
* [Community Member](#community-member)
* [Contributor](#contributor)
  * [How to become a contributor](#how-to-become-a-contributor)
* [Maintainer](#maintainer)
  * [How to become a maintainer](#how-to-become-a-maintainer)
  * [Involuntary Removal or Demotion](#involuntary-removal-or-demotion)
  * [Stepping Down/Emeritus Process](#stepping-downemeritus-process)
* [Admin](#admin)
  * [How to become an admin](#admin)
* [Release Manager](#release-manager)
  * [How to become an release manager](#how-to-become-a-release-manager)
---

Our ladder defines the roles and responsibilities for this project and how to
participate with the goal of moving from a user to a maintainer. You will need
to gain people's trust, demonstrate your competence and understanding, and meet
the requirements of the role.

## Roles
* Community Member
* Contributor
* Maintainer 
  * Porter Maintainer
  * Porter Operator Maintainer
  * Porter Wesbite Maintainer 
  * Porter Triage Lead
* Release Manager

## Community Member

Everyone is a community member! 😄 You've read this far so you are already ahead. 💯

Here are some ideas for how you can be more involved and participate in the community:

* Comment on an issue that you’re interested in.
* Submit a pull request to fix an issue.
* Report a bug.
* Share a bundle that you made and how it went.
* Come chat with us in [Slack][slack].

They must follow our [Code of Conduct](CODE_OF_CONDUCT.md).

[slack]: https://porter.sh/community#slack

## Contributor

[Contributors][contributors] have the following capabilities:

* Have issues and pull requests assigned to them
* Apply labels, milestones and projects
* [Mark issues as duplicates](https://help.github.com/en/articles/about-duplicate-issues-and-pull-requests)
* Close, reopen, and assign issues and pull requests

They must agree to and follow this [Contributing Guide](CONTRIBUTING.md).

### How to become a contributor

To become a contributor, the maintainers of the project would like to see you:

* Comment on issues with your experiences and opinions.
* Add your comments and reviews on pull requests.
* Contribute pull requests.
* Open issues with bugs, experience reports, and questions.

Contributors and maintainers will do their best to watch for community members
who may make good contributors. But don’t be shy, if you feel that this is you,
please reach out to one or more of the contributors or maintainers.

[contributors]: https://github.com/orgs/getporter/teams/contributors

## Maintainer

[Maintainers][maintainers] are members with extra capabilities:

* Be a [Code Owner](.github/CODEOWNERS) and have reviews automatically requested.
* Review pull requests.
* Merge pull requests.

There are three sub-types of specialization that maintainers can have:
  * Porter Maintainer - This is someone who focuses on [Porter Core](https://github.com/getporter/porter) functionality
  * Porter Operator Maintainer - This is someone who focuses on [Porter Operator](https://github.com/getporter/operator) functionality 
  * Porter Wesbite Maintainer  - This is someone who helps our frontend, which leverages Hugo.
  * Porter Community Lead - This is someone who handles the development of the community through scheduling meetings, encouraging Porter activties within the community (talks, blogposts, etc), and is the face of Porter
  * Porter Mixin & Plugins Specialist - This is someone who builds and maintains the mixins used to help Porter work with other tooling. 

Maintainers also have additional responsibilities beyond just merging code:

* Help foster a safe and welcoming environment for all project participants.
  This will include understanding and enforcing our [Code of Conduct](CODE_OF_CONDUCT.md).
* Organize and promote pull request reviews, e.g. prompting community members,
  contributors, and other maintainers to review.
* Triage issues, e.g. adding labels, promoting discussions, finalizing decisions.
* Help organize our development meetings, e.g. schedule, organize and
  execute agenda.

They must agree to and follow the [Reviewing Guide](REVIEWING.md).

[maintainers]: https://github.com/orgs/getporter/teams/maintainers

### How to become a maintainer

To become a maintainer, we would like you to see you be an effective
contributor, and show that you can do some of the things maintainers do.
Maintainers will do their best to regularly discuss promoting contributors. But
don’t be shy, if you feel that this is you, please reach out to one or more of
the maintainers.

## Release Managers

[Release Managers][release managers] can be either contributors or maintainers.
Porter releases on a quarterly candence, and a release manager handles kicking off
the release process & communicating the release. The release manager role is set **per**
release.

Release Manager responsibilities are:
* Help foster a safe and welcoming environment for all project participants.
  This will include understanding and enforcing our [Code of Conduct](CODE_OF_CONDUCT.md).
* Start, and if necessary, troubleshoot the [release process](./GOVERNANCE.md#release-process) for that release
* Communicate through channels (Slack, [Groups](https://groups.io/g/porter), and website)
with key achievements in that release 
* Update [release documentation](./GOVERNANCE.md#release-process) with any new findings


[release managers]: https://github.com/orgs/getporter/teams/release

### How to become a release manager

Anyone can become a release manager, all you have to do is reach out to a maintainer
who will give you the proper documentation, help you identify the release date and be
available on the date of the release. It is recommended release managers try to sign up
for at least 2 (two) releases, so they can get comfortable with the release process. 


## Inactivity
It is important for maintainers to stay active to set an example and show commitment to the project.
Inactivity is harmful to the project as it may lead to unexpected delays, contributor attrition, and a loss of trust in the project.

* Inactivity is measured by:
    * Periods of no contributions for longer than 6 months, where contributions must include maintainer-level tasks:
      reviewing and merging others pull requests, project administration, release management, mentoring, etc.
      Code contributions are not strictly required to be considered active.
    * Periods of no communication for longer than 6 months.
* Consequences of being inactive include:
    * Involuntary removal or demotion.
    * Being asked to move to Emeritus status.

## Involuntary Removal or Demotion

Involuntary removal/demotion of a maintainer happens when responsibilities and requirements aren't being met.
This may include repeated patterns of inactivity, extended period of inactivity, a period of failing to meet the requirements of your role, and/or a violation of the Code of Conduct.
This process is important because it protects the community and its deliverables while also opens up opportunities for new contributors to step in.

Removal or demotion is handled first by attempting to contact the maintainer in question to suggest stepping down.
If they cannot be reached, or will not resume their maintainer responsibilities, involuntary removal is initiated through a vote by a majority of the other current Maintainers.

## Stepping Down/Emeritus Process
If and when contributors' commitment levels change, contributors can consider stepping down (moving down the contributor ladder) instead of moving to emeritus status (completely stepping away from the project).

Contact the Maintainers about changing to Emeritus status, or reducing your contributor level.
When an Emeritus Maintainer has been an active contributor for 1 month, they can reapply to be considered for the Maintainer role again.

## Admin

[Admins][admins] are maintainers with extra responsibilities:

* Create new mixin repositories
* Manage getporter repositories
* Manage getporter teams

[admins]: https://github.com/orgs/getporter/teams/admins

### How to become an admin

It isn't expected that all maintainers will need or want to move up to admin. If
you are a maintainer, and find yourself often asking an admin to do certain
tasks for you and you would like to help out with administrative tasks, please
reach out to one or more of the admins.
