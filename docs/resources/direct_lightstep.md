---
page_title: "nobl9_direct_lightstep Resource - terraform-provider-nobl9"
description: |-
  Lightstep Direct | Nobl9 Documentation https://docs.nobl9.com/Sources/lightstep#lightstep-direct.
---

# nobl9_direct_lightstep (Resource)

Lightstep is an observability platform that enables distributed tracing, that can be used to rapidly pinpoint the causes of failures and poor performance across the deeply complex dependencies among services, teams, and workloads in modern production systems. Nobl9 integration with Lightstep enables organizations to establish service level objectives from performance data captured through distributed traces in the Lightstep platform. Nobl9 connects with Lightstep to collect SLI measurements and compare them to SLO targets.

For more information, refer to [Lightstep Direct | Nobl9 Documentation](https://docs.nobl9.com/Sources/lightstep#lightstep-direct).

## Example Usage

```terraform
resource "nobl9_direct_lightstep" "test-lightstep" {
  name                   = "test-lightstep"
  project                = "terraform"
  description            = "desc"
  source_of              = ["Metrics", "Services"]
  lightstep_organization = "acme"
  lightstep_project      = "project1"
  app_token              = "secret"
  historical_data_retrieval {
    default_duration {
      unit  = "Day"
      value = 0
    }
    max_duration {
      unit  = "Day"
      value = 30
    }
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `lightstep_organization` (String) Organization name registered in Lightstep.
- `lightstep_project` (String) Name of the Lightstep project.
- `name` (String) Unique name of the resource, must conform to the naming convention from [DNS RFC1123](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names).
- `project` (String) Name of the Nobl9 project the resource sits in, must conform to the naming convention from [DNS RFC1123](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names).
- `source_of` (List of String) Source of Metrics and/or Services

### Optional

- `app_token` (String, Sensitive) [required] | Lightstep App Token.
- `description` (String) Optional description of the resource. Here, you can add details about who is responsible for the integration (team/owner) or the purpose of creating it.
- `display_name` (String) User-friendly display name of the resource.
- `historical_data_retrieval` (Block List, Max: 1) [Replay configuration documentation](https://docs.nobl9.com/Features/replay) (see [below for nested schema](#nestedblock--historical_data_retrieval))

### Read-Only

- `id` (String) The ID of this resource.
- `status` (String) Status of the created direct.

<a id="nestedblock--historical_data_retrieval"></a>
### Nested Schema for `historical_data_retrieval`

Required:

- `default_duration` (Block List, Min: 1) Used by default for any SLOs connected to this data source. (see [below for nested schema](#nestedblock--historical_data_retrieval--default_duration))
- `max_duration` (Block List, Min: 1) Defines the maximum period for which data can be retrieved (see [below for nested schema](#nestedblock--historical_data_retrieval--max_duration))

<a id="nestedblock--historical_data_retrieval--default_duration"></a>
### Nested Schema for `historical_data_retrieval.default_duration`

Required:

- `unit` (String) Must be one of Minute, Hour, or Day.
- `value` (Number) Must be an integer greater than or equal to 0


<a id="nestedblock--historical_data_retrieval--max_duration"></a>
### Nested Schema for `historical_data_retrieval.max_duration`

Required:

- `unit` (String) Must be one of Minute, Hour, or Day.
- `value` (Number) Must be an integer greater than or equal to 0

## Nobl9 Official Documentation

https://docs.nobl9.com/