---
title: Validation and CI
description: Validate generated JSON and prevent committed schemas from drifting.
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
and nested structure.

```go
if err := ToolInput{}.ValidateJSON(data); err != nil {
    var validationErr *jsonschema.ValidationError
    if errors.As(err, &validationErr) {
        log.Printf("invalid field: %s", validationErr.InstanceLocation)
    }
    return err
}
```

## Fail CI on drift

`JSONSCHEMA_NO_CHANGES` flows through every `go generate` directive and makes
the generator fail without writing when output would change:

```yaml
- name: Check generated schemas
  run: JSONSCHEMA_NO_CHANGES=1 go generate ./...
```

For repositories with other generators, use the broader fallback:

```yaml
- name: Check all generated files
  run: go generate ./... && git diff --exit-code
```

Run `go test ./...` after the drift check so tests execute against the same
generated state that will be committed.
