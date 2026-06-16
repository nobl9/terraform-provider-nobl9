# AGENTS.md

This repository contains Terraform provider implementation for Nobl9 platform resources.
Do not use this file as a replacement for the project documentation.

Read the existing docs before changing behavior, tests, or release logic:

- [dev-docs/DEVELOPMENT.md](./dev-docs/DEVELOPMENT.md) for development workflow,
  Makefile behavior, CI, validation tests, e2e tests, code generation, and
  dependencies.
- [README.md](./README.md) for user-facing purpose and usage.
- [dev-docs/RELEASE.md](./dev-docs/RELEASE.md) for release automation details.
- [dev-docs/plugin-framework-migration.md](./dev-docs/plugin-framework-migration.md)
  for instructions when migrating [legacy SDK resources](./nobl9/)
  to the [new framework library](./internal/frameworkprovider/).

If a workflow is documented there, follow the existing doc instead of adding
a second version here.

## Migration status

This repository is in the middle of migration between the legacy SDK,
located under `./nobl9` and the new provider framework located under
`./internal/frameworkprovider`.
Depending on which one you're working under, adhere to the respective standards.
Read more about it in [./dev-docs/plugin-framework-migration.md](./dev-docs/plugin-framework-migration.md).

## Testing

Write end-to-end acceptance tests over unit tests by default.

Unit tests are acceptable for narrow internal logic, edge cases,
or failure branches that cannot be exercised through acceptance tests without brittle
setup or excessive external state.
If you choose unit-only coverage for a behavior change, state the reason in
the PR or handoff notes.

Acceptance tests talk to the Nobl9 platform API.
Do not run them without explicit user permission.

Before writing or modifying acceptance tests, read:

- [dev-docs/DEVELOPMENT.md](./dev-docs/DEVELOPMENT.md#testing)
- sample existing tests to follow the established style and practices

## Code standards

Follow existing package layout, command patterns, and test style before adding
new abstractions.

Do not edit generated files directly.
If generated output is stale, update the source definitions and run
`make generate`, then verify with `make check/generate`.

### Shell

Use the Makefile targets instead of calling tools directly.
To inspect available targets, run: `make help`.
The CI workflows under [.github/workflows](./.github/workflows/) use the same
Makefile targets, so treat them as the local source of verification commands.

## Pull requests

When creating or updating a pull request description,
follow guidelines and template defined in
[.github/pull_request_template.md](./.github/pull_request_template.md).
Do not introduce new sections, subsections for the existing ones are acceptable.
Fill only the sections that apply, remove template instructions from the final
description, and remove the `## Release Notes` section entirely when the change
does not need release notes.
Always ask the use for for `## Motivation`,
unless you already know it from a spec or ticket.
Be vigilant of any breaking changes and document them in `## Breaking Changes` section.

## Verification

Always verify changes with project targets before claiming completion.
For Markdown-only changes, run `make check/markdown` at minimum.

If a command cannot be run locally, report the exact command and exact error.
Do not replace failed verification with assumptions.
