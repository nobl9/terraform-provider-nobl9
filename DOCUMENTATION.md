# Documentation

[Official Terraform Provider Documentation](https://www.terraform.io/registry/providers/docs)

This section provides all the information needed to work with `terraform-provider-nobl9` documentation.

## Tool

Documentation is created using the [tfplugindocs](https://github.com/hashicorp/terraform-plugin-docs) tool.

## How to generate docs

- Update (if needed) templates available undes `templates/` directory
  - Each resource has a separate template file, ex. `templates/resources/slo.md.tmpl`
  - We use generic templates for index and resource pages: `templates/index.md.tmpl` and `templates/resources.md.tmpl` 
- Do not touch anything under `docs/` directory.
- 

The `examples` and `templates` directories are used to generate the docs in the `docs` folder.
