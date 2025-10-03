# Development

This document describes the intricacies of Terraform Provider development
workflow.
If you see anything missing, feel free to contribute :)

> [!IMPORTANT]
> Nobl9 Terraform Provider is being actively rewritten in
> Terraform Plugin Framework.
> For more details on this process and how to work with
> both the old SDK and Plugin Framework, please see
> [this document](./plugin-framework-migration.md).

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

## Debugging

If you want to debug a specific configuration,
you can create a temporary test scenario, like the one below,
and run it in debug mode.
You can place breakpoints directly in the provider code
and go through every step of the process.

```go
func TestAccSLOResource_customScenario(t *testing.T) {
   t.Parallel()
   testAccSetup(t)

   sloConfig := `resource "nobl9_slo" "this" {

  name = "test-slo-framework"
  project = "amazon-prometheus"
  description = "Example SLO from testing Framework migration"

  service = "amazon-prometheus"
  budgeting_method = "Occurrences"

  indicator {
    name = "amazon-prometheus"
    project = "amazon-prometheus"
    kind = "Agent"
  }

  objective {
    name = "tf-objective-1"
    op = "lt"
    target = 0.7
    value = 1.2
    raw_metric {
      query {
        amazon_prometheus {
          promql = "some_metric{job=\"test-job\"}"
        }
      }
    }
  }

  time_window {
    count = 1
    is_rolling = true
    unit = "Hour"
  }
}`

   resource.Test(t, resource.TestCase{
      ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
      Steps: []resource.TestStep{
         // Create and Read.
         {
            Config: sloConfig,
            ConfigPlanChecks: resource.ConfigPlanChecks{
               PreApply: []plancheck.PlanCheck{
                  plancheck.ExpectNonEmptyPlan(),
                  plancheck.ExpectResourceAction("nobl9_slo.this", plancheck.ResourceActionCreate),
               },
            },
         },
      },
   })
}
```

## Generating documentation

Documentation is generated using the
[tfplugindocs](https://github.com/hashicorp/terraform-plugin-docs) tool.
In order to generate or update the docs run the following command:

```sh
make generate
```

**How does it work (in short)?**

[Go templates](https://pkg.go.dev/text/template) are used
to render Markdown template files located under [./templates](../templates/).
The variables used in templates (e.g. `{{ .Name }}`) are populated by the
_tfplugindocs_ tool based on predefined and standardized fields which
the provider exposes.

The rendered Markdown files are stored under [./docs](../docs/) directory.
This directory is a standardized path, required by Terraform registry and
it is scraped and rendered on
[the registry website](https://registry.terraform.io/providers/nobl9/nobl9/latest/docs).

It's worth to highlight one field in particular: `{{ .SchemaMarkdown }}`.
This field contains the entire resource schema definition, which is already
rendered as Markdown.
Each attribute and block in the resource's schema is defined in code and
what's getting rendered is a combination of its type, name and custom
description we provide. Only the latter can be changed, for instance,
SLO's service attribute is defined like this:

```go
"service": schema.StringAttribute{
	Required:    true,
	Description: "Name of the service.",
},
```

This renders as:

```md
- `service` (String) Name of the service.
```

> [!WARNING]
> Note, that you can only change the `Description`!

Additionally, we often provide example Terraform configurations
for each resource.
The examples are located under [./examples](../examples/) directory
and you can place them in the templates using the following functions:

```md
{{ tffile (printf "examples/resources/%s/resource.tf" .Name) }}
```

## How to use local provider in Terraform

### Installing

1. Go to the repo root.
2. Before the next step, verify if the Makefile variable `OS_ARCH` matches your
    system (for example _darwin_arm64_ for Apple Silicon based Mac's).
    If not override it.
3. Run `make install/provider`. Make sure that the plugin was installed:
    `ls ~/.terraform.d/plugins/nobl9.com/nobl9/nobl9/`
    It will show you the current version of the plugin, ex: _0.19.0_.
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
