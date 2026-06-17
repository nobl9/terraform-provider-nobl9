# Release process

The internal release process is described in great detail
[here](http://go/terraform-provider-release).

## Release automation details

We're using [Release Drafter](https://github.com/release-drafter/release-drafter)
to automate release notes creation. Drafter also does its best to propose
the next release version based on commit messages from `main` branch.

Release Drafter is also responsible for auto-labeling pull requests.
It checks both title and body of the pull request and adds appropriate labels. \
**NOTE:** The auto-labeling mechanism will not remove labels once they're
created. For example, If you end up changing PR title from `sec:` to `fix:`
you'll have to manually remove `security` label.

On each commit to `main` branch, Release Drafter will update the next release
draft.

To start a release, run the `Promote Release Draft` workflow manually.
The workflow resolves the next tag from Release Drafter, updates the generated
release notes, and pushes that tag to GitHub.
The tag push triggers the `Release` workflow, which runs acceptance tests,
waits for QA approval, and publishes assets with GoReleaser.

Do not update provider versions manually in the Makefile, README, or examples.
Local builds derive the provider version from the latest reachable Git tag,
and GoReleaser derives published artifact versions from the release tag.

In addition to Release Drafter, we're also running a script which extracts
explicitly listed release notes and breaking changes which are optionally
defined in `## Release Notes` and `## Breaking Changes` headers.
It also performs a cleanup of the PR draft mitigating Release Drafter
shortcomings.
