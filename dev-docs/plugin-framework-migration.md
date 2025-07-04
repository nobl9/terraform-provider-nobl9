# Terraform Plugin Framework Migration

Originally Nobl9 Terraform Provider (referred to as "provider") was written
in [Terraform Plugin SDK](https://github.com/hashicorp/terraform-plugin-sdk),
from now on referred to as "SDK".

Currently, the provider is in the process of being rewritten
to use [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework),
from now on referred to as "framework".

## Rationale

The _tldr;_ rationale for this move is simple, the SDK does not have the
features we want, which the framework provides.
Quoting Terraform docs:

> For new provider development it is recommended to investigate
terraform-plugin-framework, which is a reimagined provider SDK
that supports additional capabilities.

The cherry on top of these sought after capabilities is that the framework
is much more elegant and idiomatic, shipping with type safety features, making
it much easier and pleasant to develop in.

## Migration

Thankfully, the process of migrating from SDK to the framework is quite
seamless. It is not only transparent to the end user but it can also be
fragmented to the point where we can rewrite each resource one by one and
allow both the SDK and framework code to function side by side.

The process of migration is well documented, you can read more about it
[here](https://developer.hashicorp.com/terraform/plugin/framework/migrating).

Currently each of the code bases is separated as follows:

- SDK lives where it used to, in [nobl9](../nobl9/) directory.
- Framework code is located at [internal/frameworkprovider](../internal/frameworkprovider/).

Both the SDK and framework have separate provider definitions!
If happen to updated provider definition in one of them but forget
to do that in the other, the tests will fail and the discrepancy
should be detected.

The glue which binds these two together is a
[multiplexer](https://github.com/hashicorp/terraform-plugin-mux),
located at the entrypoint [main.go](../main.go).

```go
muxServer, err := tf5muxserver.NewMuxServer(
	ctx,
	newSDKProvider(Version),
	newFrameworkProvider(Version),
)
```

Since both the SDK and framework providers talk using the same protocol (gRPC),
the multiplexer can delegate each resource to one of these providers.

## Moving resource from SDK to framework

We recommend you take a look at [this PR](https://github.com/nobl9/terraform-provider-nobl9/pull/425).

You will see that the tests look slightly different, that is intentional.
The acceptance tests defined in the SDK code often leave a lot to be desired.
We want to use this rewrite as an excuse to improve them and make them
more readable and better reflect user journeys.

## Know caveats

### Blocks vs Attributes

Blocks in Terraform Plugin Framework are not the same as in the SDK.
Namely, they lack options like `Optional` or `Required`.
In general, they are no longer recommended and should be avoided, however,
we can't really replace them with attributes without
breaking backwards compatibility.

Block looks like this:

```terraform
block {
  name = "block_name"
}
```

Attributes, on the other hand, look like this:

```terraform
attribute = {
  name = "attribute_name"
}
```

The only stylistic difference is the `=` sign in the latter.

Functionally, they are the almost the same, but the framework favors attributes.

So how do we migrate required block from the SDK to the framework?
We need to define each block as a list in the framework model,
mind you, it will always have exactly one item!
On top of that we need to define validators to enforce the size and optionally
mark it as required.

Before:

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

After:

```go
schema.ListNestedBlock{
    Description: "Time window configuration for the SLO.",
    Validators: []validator.List{
        listvalidator.IsRequired(),
        listvalidator.SizeBetween(1, 1),
    },
    NestedObject: schema.NestedBlockObject{
        // ..
    },
},
```

If we wanted to make it optional,
we would simply remove the `IsRequired()` validator and
swap the `SizeBetween(1, 1)` to `SizeAtMost(1)`.
