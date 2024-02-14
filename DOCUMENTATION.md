# Documentation

[Official Terraform Provider Documentation](https://www.terraform.io/registry/providers/docs)

This section provides all the information needed to work with
`terraform-provider-nobl9` documentation.

## Tool

Documentation is created using the
[tfplugindocs](https://github.com/hashicorp/terraform-plugin-docs) tool.

## Which files should I change?

- Update (if needed) templates available under `templates/` directory
  - Each resource has a separate template file, e.g. `templates/resources/slo.md.tmpl`
  - We use generic templates for index and resource pages:
    `templates/index.md.tmpl` and `templates/resources.md.tmpl`.
  - Use Data Fields supported by [tfplugindocs](https://github.com/hashicorp/terraform-plugin-docs).
- Update (if needed) examples available under `examples/` directory
  - Make sure that all examples are working with the latest version of the provider
  - `examples/provider/provider.tf` is the example that will be rendered
    on the main page on
    [provider documentation](https://registry.terraform.io/providers/nobl9/nobl9/latest/docs#schema).
  - `examples/resources/<resource_name>/resource.tf` will be rendered for every
    resource on their documentation page, e.g.
    <!-- markdownlint-disable MD034 -->
    https://registry.terraform.io/providers/nobl9/nobl9/latest/docs/resources/slo.
    <!-- markdownlint-enable MD034 -->
- Update (if needed) `"description"` field for a resource,
  e.g. in `nobl9/resource_slo.go`:

  ```go
  Schema: map[string]*schema.Schema{
    ...
    "description":  "Your new description"
    ...
  }
  ```

  - This description will be rendered on the documentation page
    for the changed resource.
- Do not touch anything under `docs/` directory.

## How to generate docs?

You need to have [Go](https://go.dev/) installed on your machine, then run:

```sh
go generate
```
