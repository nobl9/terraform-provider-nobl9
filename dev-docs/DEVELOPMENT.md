# Development

This document describes the intricacies of Terraform Provider development
workflow.
If you see anything missing, feel free to contribute :)

## Pull requests

[Pull request template](../.github/pull_request_template.md)
is provided when you create new PR.
Section worth noting and getting familiar with is located under
`## Release Notes` header.

## Makefile

Run `make help` to display short description for each target.
The provided Makefile will automatically install dev dependencies if they're
missing and place them under `bin`
(this does not apply to `yarn` managed dependencies).
However, it does not detect if the binary you have is up to date with the
versions declaration located in Makefile.
If you see any discrepancies between CI and your local runs, remove the
binaries from `bin` and let Makefile reinstall them with the latest version.

## CI

Continuous integration pipelines utilize the same Makefile commands which
you run locally. This ensures consistent behavior of the executed checks
and makes local debugging easier.

## Testing

Terraform Provider is mainly tested with acceptance tests, which are plain Go
tests run with an overlay of Terraform SDK orchestration.
You can run them with `make test/acc` (recommended) or from GitHub by dispatching
[this workflow](https://github.com/nobl9/terraform-provider-nobl9/actions/workflows/acc-tests-dispatch.yml).
More on acceptance tests can be found
[here](https://developer.hashicorp.com/terraform/plugin/sdkv2/testing/acceptance-tests).

The acceptance tests are only run automatically for releases, be it official
version or pre-release (release candidate).
The tests are executed against the production application.
If you want to run the tests manually against a different environment, you can
run the following command:

```shell
NOBL9_CLIENT_ID=<client_id> \
NOBL9_CLIENT_SECRET=<client_secret> \
NOBL9_OKTA_URL=https://accounts.nobl9.dev \
NOBL9_OKTA_AUTH=<dev_auth_server> \
NOBL9_URL=<ingest_server_url> \
make test/acc
```

## How to use local provider in Terraform

### Installing

1. Go to the repo root.
2. Before the next step, verify if the Makefile variable `OS_ARCH` matches your
  system (for example *darwin_arm64* for Apple Silicon based Mac's).
  If not override it.
3. Run `make install`. Make sure that the plugin was installed:
  `ls ~/.terraform.d/plugins/nobl9.com/nobl9/nobl9/`
  It will show you the current version of the plugin, ex: *0.19.0*.
4. Copy the path to the plugin after ~/.terraform.d/plugins/, for example:
  `nobl9.com/nobl9/nobl9/0.19.0/linux_amd64/terraform-provider-nobl9`
  and configure your `.tf` file with it.
  Usually it will look like this, just change the version:
  ```terraform
  terraform {
    required_providers {
      nobl9 = {
        source = "nobl9.com/nobl9/nobl9"
        version = "0.19.0"
      }
    }
  }
  ```
  Now you're all set, you can use the locally built provider anywhere, as long
  as you use the right version (see above).

## Releases

Refer to [RELEASE.md](./RELEASE.md) for more information on release process.

## Dependencies

Renovate is configured to automatically merge minor and patch updates.
For major versions, which sadly includes GitHub Actions, manual approval
is required.
