# Terraform Plugin Framework Migration

The Nobl9 Terraform Provider was originally implemented with the
[Terraform Plugin SDK](https://github.com/hashicorp/terraform-plugin-sdk),
referred to as the SDK in this document.

The provider is being migrated incrementally to the
[Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework),
referred to as the framework.

## Rationale

The framework provides typed resource models and first-class plan modification
APIs. These capabilities let the provider work with typed planned state,
produce targeted diagnostics before destructive operations,
and avoid operating on generic `map[string]any` values.

## Migration architecture

The two implementations are separated as follows:

- SDK code is located in the [nobl9](../nobl9/) directory.
- Framework code is located in
  [internal/frameworkprovider](../internal/frameworkprovider/).

The SDK and framework have separate provider definitions.
Their provider-level schemas and configuration behavior must remain compatible
while both are served by the same provider binary.
This includes defaults, environment variables, configuration-file loading,
credential validation, error messages, and version reporting.

The [provider entry point](../main.go) combines both implementations with the
[Terraform provider multiplexer](https://github.com/hashicorp/terraform-plugin-mux):

```go
muxServer, err := tf6muxserver.NewMuxServer(
	ctx,
	newSDKProvider(ctx),
	newFrameworkProvider(),
)
```

The framework provider uses Terraform Plugin Protocol v6.
The SDK provider uses protocol v5 and is converted to v6 with `tf5to6server`
before it is passed to the multiplexer.
Protocol v6 requires Terraform 0.15 or newer.

Acceptance tests must instantiate the complete protocol v6 multiplexer,
even when a test exercises only one implementation.
Testing an SDK or framework provider directly does not represent the shipped
provider and can leave resources from the other implementation undefined.
See [the shared acceptance-test setup](../internal/frameworkprovider/provider_test.go).

## Moving a resource from the SDK to the framework

Use the
[Project resource migration](https://github.com/nobl9/terraform-provider-nobl9/pull/425)
for a small example and the
[SLO resource migration](https://github.com/nobl9/terraform-provider-nobl9/pull/455)
for a resource with nested schemas and state compatibility concerns.

For each migrated resource:

1. Inventory the complete SDK schema and its released state representation.
2. Map every attribute, block, validation rule, default, replacement rule,
   diff suppression rule, and import format to framework behavior.
3. Implement the framework model, schema, lifecycle methods, and acceptance
   tests.
4. Add state upgrade logic when released state cannot be decoded by the
   framework schema.
5. Register the resource in the framework provider and remove its SDK
   registration in the same change.
6. Run `make generate` and verify the generated resource documentation.
7. Run the project checks and tests described in
   [DEVELOPMENT.md](./DEVELOPMENT.md).

Migration is not automatically transparent to existing users.
Compatibility depends on preserving configuration syntax,
plan behavior, and state produced by released SDK versions.

## Known caveats

### Blocks and attributes

Blocks in the framework do not expose `Optional` and `Required` fields.
Existing blocks also cannot be changed to nested attributes during migration
without changing the Terraform configuration syntax:

```terraform
block {
  name = "block_name"
}
```

The equivalent nested attribute requires an equals sign:

```terraform
attribute = {
  name = "attribute_name"
}
```

Preserve existing blocks during migration.
Represent a single nested block as a list in the framework model and enforce
its cardinality with validators.

For example, migrate a required single block from this SDK schema:

```go
"time_window": {
    Type:        schema.TypeSet,
    Required:    true,
    MaxItems:    1,
    Elem: &schema.Resource{
        // ...
    },
},
```

To this framework schema:

```go
schema.ListNestedBlock{
    Description: "Time window configuration for the SLO.",
    Validators: []validator.List{
        listvalidator.IsRequired(),
        listvalidator.SizeBetween(1, 1),
    },
    NestedObject: schema.NestedBlockObject{
        // ...
    },
}
```

For an optional single block,
omit `IsRequired` and use `SizeAtMost(1)`.
Do not add a minimum-size validator if the SDK schema allowed an explicitly
empty collection.
That mistake caused the `alert_policies = []` regression fixed in
[PR #490](https://github.com/nobl9/terraform-provider-nobl9/pull/490).

For new fields, prefer nested attributes when backward compatibility does not
require block syntax.

The generated documentation does not include block cardinality validators.
State whether a block is required or optional and document its minimum and
maximum size in the schema description.

### Translate all SDK schema behavior

Framework migration requires an explicit replacement for each SDK behavior.
In particular, audit:

- `ForceNew` and destructive replacement behavior;
- `DiffSuppressFunc` and state normalization;
- `ConflictsWith`, `AtLeastOneOf`, `ExactlyOneOf`, and related constraints;
- `MinItems`, `MaxItems`, and collection uniqueness;
- `ValidateFunc`, defaults, computed values, and sensitive values;
- import ID validation and parsing.

Do not make the framework schema stricter solely because a convenient
validator exists.
Test the exact values accepted by the SDK,
including omitted and explicitly empty values.

Apply relationship validators to the side that produces the clearest
diagnostic.
Applying conflict validators to both sides can emit duplicate diagnostics.
Use a custom validator when built-in validators cannot preserve the intended
contract or diagnostic,
as in [PR #560](https://github.com/nobl9/terraform-provider-nobl9/pull/560).

### Preserve Terraform value semantics

Terraform distinguishes among null, unknown, omitted, empty,
and explicit zero values.
The API can treat several of these values as equivalent,
but the framework still requires the value returned after apply to agree with
the plan.

Use framework `types` or pointers for optional primitive values when their
presence matters.
Use Go primitives only when the value is guaranteed to be known and non-null.

When reading an API response,
preserve the configuration or prior-state representation if the API:

- omits an explicitly empty collection through `omitempty`;
- replaces an empty value with a server-side default;
- cannot distinguish omitted from explicit `false` or zero;
- normalizes a deprecated attribute into its replacement.

Otherwise Terraform can report perpetual drift or an error such as
`Provider produced inconsistent result after apply`.
Examples include the empty alert policy fix in
[PR #490](https://github.com/nobl9/terraform-provider-nobl9/pull/490),
the composite aggregation default fix in
[PR #525](https://github.com/nobl9/terraform-provider-nobl9/pull/525),
and preservation of omitted and explicit Boolean values in
[PR #588](https://github.com/nobl9/terraform-provider-nobl9/pull/588).

Test omitted, explicitly empty, explicit zero or `false`,
and explicit default values separately.

### Preserve state from released SDK versions

The framework decodes prior Terraform state against the resource schema before
normal lifecycle methods can repair it.
Removing an SDK attribute from the framework schema can therefore fail before
`Read` is called.
Changing an attribute to write-only can also fail when an older provider stored
that attribute in state.

Before removing or changing an existing field:

1. Inspect state written by released provider versions.
2. Increment the framework schema version.
3. Implement `resource.ResourceWithUpgradeState`.
4. Transform or remove incompatible raw state before decoding it with the new
   schema.
5. Test an upgrade from the last released SDK-backed version.

The SLO migration initially failed with
`error marshaling prior state: unsupported attribute "attachments"` and
`write-only attributes cannot be read back from the provider`.
[PR #533](https://github.com/nobl9/terraform-provider-nobl9/pull/533)
introduced the corresponding
[SLO state upgrader](../internal/frameworkprovider/slo_resource_upgrader.go).

### Hydrate complete state during import

An import initially contains only the resource identifier.
Decoding that partial state into a framework model can fail when required
attributes use Go primitive types because those types cannot represent null.

The current resource convention is to:

1. Validate and parse the import ID.
2. Fetch the object from the Nobl9 API inside `ImportState`.
3. Convert the object into the complete framework model.
4. Set the complete state.

See the [Project import implementation](../internal/frameworkprovider/project_resource.go)
and [SLO import implementation](../internal/frameworkprovider/slo_resource.go).
Every migrated resource must have an import acceptance test followed by a
no-op plan.

### Define collection identity and ordering

Framework sets determine nested-object identity from the entire object.
They cannot identify an object by one selected field such as a label key.
Changing one field in a nested set can therefore appear as deleting one object
and creating another.

Use a list when domain identity is based on only part of a nested object.
When the API does not preserve order,
sort the API response to match configuration or prior-state order before
writing state.
Use a set only when every element is genuinely unordered and its complete value
defines its identity.

Conversion from a list to a map can silently discard duplicate keys.
Add an explicit uniqueness validator before that conversion.
The shared label implementation combines stable ordering,
set-valued label values,
and label-key uniqueness in
[metadata_labels_block.go](../internal/frameworkprovider/metadata_labels_block.go).
This pattern was introduced during the
[Service migration](https://github.com/nobl9/terraform-provider-nobl9/pull/351)
to replace SDK diff suppression for labels.

Test API responses in a different order from configuration,
updates to one nested object,
and duplicate domain keys.
These cases have caused fixes in
[PR #509](https://github.com/nobl9/terraform-provider-nobl9/pull/509)
and [PR #576](https://github.com/nobl9/terraform-provider-nobl9/pull/576).

### Audit plan modifiers and plan-time API calls

Translate every SDK replacement rule into a framework plan modifier.
Preserve warnings for destructive replacements,
especially when replacing a Project or Service cascades to SLOs and their data.

Use `UseStateForUnknown` only when the computed value cannot change as a result
of the current configuration change.
Reusing an old computed value when another attribute affects it can make the
provider return a value that contradicts the plan,
as happened with Service review cycle status in
[PR #506](https://github.com/nobl9/terraform-provider-nobl9/pull/506).

`ModifyPlan` implementations must handle null destroy plans and unknown values.
The Project, Service, and SLO resources also perform Nobl9 dry-run API requests
during planning.
Plan behavior therefore depends on valid provider configuration,
network availability, and the Nobl9 API.
Keep validation errors actionable and distinguish them from transient API
failures.
See [PR #484](https://github.com/nobl9/terraform-provider-nobl9/pull/484).

If an update maps to multiple API operations,
handle partial success explicitly.
For example, moving an SLO and then applying its updated definition can leave
the SLO in the target Project even when the second operation fails.
The implementation must report that resulting state to the user.
See [PR #466](https://github.com/nobl9/terraform-provider-nobl9/pull/466).

### Keep provider configuration behavior aligned

The SDK and framework provider definitions must do more than expose matching
schema types.
Verify matching behavior for:

- default values and environment-variable names;
- precedence among Terraform configuration, environment variables,
  and the Nobl9 configuration file;
- partial and missing credentials;
- alternate authentication methods;
- endpoint overrides;
- provider version and SDK user-agent reporting.

Validate credentials after the Nobl9 SDK resolves all supported configuration
sources.
Validating only raw Terraform attributes incorrectly rejects credentials loaded
from the Nobl9 configuration file.

Historical fixes in
[PR #366](https://github.com/nobl9/terraform-provider-nobl9/pull/366),
[PR #419](https://github.com/nobl9/terraform-provider-nobl9/pull/419),
[PR #460](https://github.com/nobl9/terraform-provider-nobl9/pull/460),
and [PR #576](https://github.com/nobl9/terraform-provider-nobl9/pull/576)
cover mux-based tests, environment configuration,
version reporting, and resolved credential validation respectively.

## Migration verification checklist

Before removing the SDK implementation of a resource, verify:

- the complete protocol v6 mux initializes successfully;
- create, read, update, delete, and import use the mux in acceptance tests;
- importing an existing object produces a subsequent no-op plan;
- state created by the last SDK-backed release upgrades successfully;
- omitted, null, empty, explicit zero or `false`, and default values remain
  stable after apply and refresh;
- API collection reordering does not produce a plan;
- duplicate domain keys fail before conversion to an API map;
- every old `ForceNew` field still replaces the resource and emits any required
  destructive-operation warning;
- plan modifiers handle create, update, replacement, destroy,
  and no-op plans;
- every provider configuration source works through both implementations;
- generated documentation describes block cardinality that validators cannot
  expose.

Acceptance tests are preferred for migrated behavior because they exercise
Terraform configuration, state, the multiplexer, and the Nobl9 API together.
Follow the testing instructions and permission requirements in
[DEVELOPMENT.md](./DEVELOPMENT.md#testing).
