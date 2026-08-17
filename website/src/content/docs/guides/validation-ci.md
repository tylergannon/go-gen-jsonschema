---
title: Validation and CI
description: Validate generated JSON or YAML and prevent committed schemas from drifting.
---

## Generate validation methods

Add `--validate` to both the generation directive and the scaffold command:

```go
//go:generate go tool gen-jsonschema --validate
```

```bash
go tool gen-jsonschema new \
  -out schema.go \
  -methods 'ToolInput=Schema' \
  --validate \
  --generate
go mod tidy
```

The generated `ValidateJSON([]byte) error` compiles the schema once at startup.
Validation covers required fields, types, unknown properties, enum membership,
and nested structure. Schema-validation failures can be inspected as
`*jsonschemav6.ValidationError`; malformed JSON may instead return a parsing
error.

```go
import (
    "errors"
    "log"

    jsonschemav6 "github.com/santhosh-tekuri/jsonschema/v6"
)

func validateToolInput(data []byte) error {
    if err := (ToolInput{}).ValidateJSON(data); err != nil {
        var validationErr *jsonschemav6.ValidationError
        if errors.As(err, &validationErr) {
            log.Printf("invalid field: %s", validationErr.InstanceLocation)
        }
        return err
    }
    return nil
}
```

For YAML input, add `--formats=both` to both commands. Generation adds
`ValidateYAML([]byte) error` and yaml/v4 decoding adapters. YAML is translated
into the JSON data model before validation and unmarshaling, so JSON Schema
property names and `json` tags remain canonical; Go `yaml` struct tags are ignored.
Because yaml/v4 does not pass decoder options into `UnmarshalYAML`,
`yaml.WithKnownFields()` cannot enforce strict fields inside registered types;
use `ValidateYAML` for schema-backed unknown-property rejection. Decode with
`yaml.WithV4Defaults()` to match `ValidateYAML` scalar resolution.

## Fail CI on drift

`JSONSCHEMA_NO_CHANGES` flows through every `go generate` directive and makes
the generator fail without writing schema files when schema JSON would change.
Generation can still update `jsonschema_gen.go` when schemas are unchanged, so
pair the command with a Git diff:

```yaml
- name: Check generated schemas
  run: JSONSCHEMA_NO_CHANGES=1 go generate ./... && test -z "$(git status --porcelain)"
```

For repositories with generators that do not understand
`JSONSCHEMA_NO_CHANGES`, use the broader fallback:

```yaml
- name: Check all generated files
  run: go generate ./... && test -z "$(git status --porcelain)"
```

Run `go test ./...` after the drift check so tests execute against the same
generated state that will be committed.
