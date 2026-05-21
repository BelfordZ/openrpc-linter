# OpenRPC Linter

[![CI](https://github.com/open-rpc/openrpc-linter/workflows/CI/badge.svg)](https://github.com/open-rpc/openrpc-linter/actions)

Fast, extensible linter for OpenRPC documents.

## Usage

```bash
# Lint with default rules
openrpc-linter lint openrpc.json -r rules.yml

# JSON output
openrpc-linter lint openrpc.json -r rules.yml -f json

# Validate document structure
openrpc-linter validate openrpc.json
```

## Install

```bash
go install github.com/open-rpc/openrpc-linter@latest
```

## Rules

Create a rules `rules.yml` with rules you want to apply:

```yaml
rules:
  method-description:
    description: "Methods must have descriptions"
    given: "$.methods[*].description"
    severity: "error"
    then:
      function: "truthy"
```

The built-in functions currently include:

- `truthy`: require a selected value to be present and non-empty
- `unique`: require every selected value to be distinct across one rule run

`unique` is symmetric with `truthy`: point `given` directly at the values you want to compare. By default duplicates are tracked in a single global bucket per rule run. Set `functionOptions.scope` to a JSONPath to partition duplicates by the longest-matching scope; targets outside every scope match are skipped (not deduped globally).

```yaml
rules:
  unique-method-names:
    description: "Method names must be unique"
    given: "$.methods[*].name"
    severity: "error"
    then:
      function: "unique"

  unique-param-names-per-method:
    description: "Param names should be unique within each method"
    given: "$.methods[*].params[*].name"
    severity: "error"
    then:
      function: "unique"
      functionOptions:
        scope: "$.methods[*]"
```

`unique` also supports `then.functionOptions.ignoreMissing`, which defaults to `true` and only matters for `given` paths whose terminal segment is a field name (so the selector can emit missing-field targets).
