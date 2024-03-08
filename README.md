<!-- markdownlint-disable line-length html -->
<h1 align="center">
   <picture>
      <source media="(prefers-color-scheme: dark)" srcset="https://github.com/nobl9/n9/assets/84388336/2895fcab-a386-498f-b7c4-14a70c66093d">
      <source media="(prefers-color-scheme: light)" srcset="https://github.com/nobl9/n9/assets/84388336/b21abc8b-8b05-448b-a8bc-5576c72e81b5">
      <img alt="N9" src="https://github.com/nobl9/n9/assets/84388336/b21abc8b-8b05-448b-a8bc-5576c72e81b5" width="500" />
   </picture>
</h1>

<div align="center">
  <table>
    <tr>
      <td>
        <img alt="checks" src="https://github.com/nobl9/terraform-provider-nobl9/actions/workflows/checks.yml/badge.svg?event=push">
      </td>
      <td>
        <img alt="tests" src="https://github.com/nobl9/terraform-provider-nobl9/actions/workflows/unit-tests.yml/badge.svg?event=push">
      </td>
      <td>
        <img alt="vulnerabilities" src="https://github.com/nobl9/terraform-provider-nobl9/actions/workflows/vulns.yml/badge.svg?event=push">
      </td>
    </tr>
  </table>
</div>
<!-- markdownlint-enable line-length html -->

[Nobl9](https://www.nobl9.com/) Terraform Provider.

# Requirements

-	[Terraform](https://www.terraform.io/downloads.html) 0.10.x

# Example

```terraform
terraform {
  required_providers {
    nobl9 = {
      source = "nobl9/nobl9"
      version = "0.23.0"
    }
  }
}

provider "nobl9" {
  client_id = "<CLIENT_ID>"
  client_secret = "<CLIENT_SECRET>"
}

resource "nobl9_project" "test" {
  name = "test"
}

resource "nobl9_service" "test" {
  name    = "test"
  project = "test"
}
```

# Documentation

Generated documentation is located under [docs](./docs) folder.

Developers' documentation sits under [dev-docs](./dev-docs) folder.
