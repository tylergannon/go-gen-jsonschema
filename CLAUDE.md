# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

go-gen-jsonschema is a Go code generator that creates JSON Schema definitions from Go types, optimized for LLM function calling (OpenAI, Anthropic). It uses `//go:build jsonschema` build tags to separate schema registration from production code.

## Commands

```bash
# Run all tests
go test ./...

# Run a specific test
go test ./... -run 'TestName'

# Lint (requires: modernize, staticcheck, govulncheck, golangci-lint, goimports)
just lint

# Build the CLI
go build ./gen-jsonschema

# Generate schemas for an example
cd examples/basictypes && go generate ./...
```

Task runner is `just` (justfile), not `make`.

## Architecture

### Two-Phase Generation Pipeline

1. **Phase 1 — JSON schema files**: Scans Go types via AST, generates `.json` files in `jsonschema/` subdirectory
2. **Phase 2 — Go code**: Generates `jsonschema_gen.go` with `embed.FS` for runtime schema access

### Package Layout

- **`gen-jsonschema/`** — CLI entry point with `gen` and `new` subcommands
- **`internal/syntax/`** — AST parsing, package loading (uses `golang.org/x/tools/go/packages` with `jsonschema` build tag), type scanning, comment extraction
- **`internal/builder/`** — Schema generation engine. `SchemaBuilder` orchestrates: type scanning → schema node construction → JSON output → Go code generation
- **`internal/builder/model.go`** — Schema node types: `ObjectNode`, `PropertyNode`, `ArrayNode`, `UnionTypeNode`, `RefNode`, `TemplateHoleNode`
- **`internal/common/`** — Struct tag parsing, helpers
- **Root package** — Public types: `JSONSchema`, `ObjectSchema`, `ParentSchema`, and marker registration functions

### Registration System

Schema types are registered via no-op marker functions in build-tagged `schema.go` files. The scanner reads these as AST call expressions:

- `NewJSONSchemaMethod(T.Schema)` — primary registration
- `WithEnum(T{}.Field)` / `WithStringerEnum(T{}.Field)` — enum fields
- `WithInterface(T{}.Field)` / `WithInterfaceImpls(...)` — union types with discriminators
- `WithStructAccessorMethod(...)` / `WithRenderProviders()` — provider-based template rendering

Legacy API: `NewEnumType[T]()`, `NewInterfaceImpl[I](impls...)`

### Key Patterns

- **Build tags**: `//go:build jsonschema` for registration code, `//go:build !jsonschema` for generated code
- **Discriminators**: Default `"!type"`, overridable per-field with `WithDiscriminator()`
- **Comments → descriptions**: Go doc comments automatically become JSON Schema `description` fields
- **Optional fields**: Detected from `json:"field,omitempty"` tags

### Validation

Every non-rendered schema type gets a generated `ValidateJSON([]byte) error` method (no stub needed). Schemas are compiled once in `init()` using `github.com/santhosh-tekuri/jsonschema/v6`. Rendered/template types are excluded because their schemas depend on runtime values.

### Limitations

- No support for maps, channels, functions, or inline interfaces
- Circular/recursive references are detected and rejected
- External package types limited to `time.Time` (rendered as string with RFC3339 guidance)
- Max nesting depth: 100

## Test Structure

- Unit tests alongside source files (`*_test.go`)
- Integration test fixtures in `internal/builder/testfixtures/` and `internal/builder/test_run/`
- Golden file comparisons via `internal/testutils/golden_file.go`
- Example directories each contain types, registration, and generated output
