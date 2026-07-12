---
name: go-gen-jsonschema
description: >
  Generate JSON Schemas from Go types with go-gen-jsonschema, optimized for LLM
  function calling and structured output (Anthropic tool_use, OpenAI tools).
  Use this whenever a Go project needs JSON Schema for its structs — defining
  LLM tools, validating LLM output, structured extraction — or when the user
  mentions gen-jsonschema, schema.go stubs, the jsonschema build tag, ValidateJSON,
  or asks to keep generated schemas in sync via go generate, lefthook, or CI.
---

# go-gen-jsonschema

Code generator that turns Go structs into JSON Schema files plus Go accessors
(`Schema()`, optionally `ValidateJSON()`). Built for LLM function calling:
properties are emitted in struct field order (deterministic, prompt-controllable),
`additionalProperties: false`, ordinary and nullable fields required,
`Optional[T]` fields optional, and Go doc comments become the schema
`description` fields.

Import path: `github.com/tylergannon/go-gen-jsonschema` (library markers) and
`github.com/tylergannon/go-gen-jsonschema/gen-jsonschema` (CLI).

## Mental model: two build-tagged files

- `schema.go` — `//go:build jsonschema`. You write this. Panic stubs + marker
  registrations. Compiled only during generation, never in production.
- `jsonschema_gen.go` — `//go:build !jsonschema`. Generated. Real `Schema()`
  and `ValidateJSON()` implementations over an `embed.FS` of `jsonschema/*.json`.

The build tags make them mutually exclusive, so the package always compiles —
before and after generation. Commit all generated outputs: `jsonschema_gen.go`
and the whole `jsonschema/` directory (each `T.json` schema comes with a
`T.json.sum` checksum the tool uses for change detection).

## Setup workflow

1. **Add the tool** (Go 1.24+ tool directive — keeps the version in go.mod so
   every contributor and CI runs the same binary):

   ```bash
   go get -tool github.com/tylergannon/go-gen-jsonschema/gen-jsonschema@latest
   ```

2. **Add the generate directive** to the file defining your types:

   ```go
   //go:generate go tool gen-jsonschema
   ```

   Add `--validate` to also generate `ValidateJSON()` methods (recommended when
   the JSON comes from an LLM): `//go:generate go tool gen-jsonschema --validate`

3. **Create the stub file.** Let the CLI write it (it derives the package name
   and stubs from your flags), then generation runs immediately via `--generate`:

   ```bash
   go tool gen-jsonschema new -out schema.go -methods 'Person=Schema,Address=Schema' --validate --generate
   ```

   Or write `schema.go` by hand — see the example below.

4. **Tidy** (only needed with `--validate`): run `go mod tidy`. The generated
   code imports `github.com/santhosh-tekuri/jsonschema/v6`, which won't be in
   go.sum until you tidy — `go build` fails with a missing-go.sum-entry error
   otherwise.

5. **Verify**: `go build ./...` and `go test ./...` must pass, and a second
   `go generate ./...` must produce no diff (generation is idempotent).

6. **Wire it into commits/CI** so schemas never drift from types — read
   [references/hooks-and-ci.md](references/hooks-and-ci.md) for lefthook and
   GitHub Actions recipes (auto-stage vs fail-on-drift).

## Minimal example

```go
// types.go
package contacts

import jsonschema "github.com/tylergannon/go-gen-jsonschema"

//go:generate go tool gen-jsonschema --validate

// Person is a single contact extracted from the document.
type Person struct {
    // Full legal name, e.g. "Ada Lovelace".
    Name string `json:"name"`

    // Age in whole years at the time of writing.
    Age int `json:"age"`

    // Email address. Omit when not stated in the source text.
    Email jsonschema.Optional[string] `json:"email,omitzero"`

    // Required key; null means no phone number was supplied.
    Phone jsonschema.Nullable[string] `json:"phone"`
}
```

```go
// schema.go
//go:build jsonschema

package contacts

import (
    "encoding/json"
    jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

// Stubs so the package compiles before generation; jsonschema_gen.go
// provides the real implementations.
func (Person) Schema() json.RawMessage     { panic("not implemented") }
func (Person) ValidateJSON(_ []byte) error { panic("not implemented") }

var _ = jsonschema.NewJSONSchemaMethod(Person.Schema)
```

Run `go generate ./...`, then use it:

```go
schema := Person{}.Schema()                      // json.RawMessage for the tool definition
if err := Person{}.ValidateJSON(llmOutput); err != nil {
    // *jsonschema.ValidationError: InstanceLocation, ErrorKind, Causes
}
```

## Doc comments ARE the schema descriptions

Every field doc comment is copied verbatim into that property's `description`,
and the type's doc comment becomes the top-level schema description. The LLM
filling the fields reads these — so write them as instructions to the model,
not as notes to Go maintainers:

- State semantics, format, units, and valid ranges: "RFC3339 timestamp",
  "score from 0.0 to 1.0", "lowercase kebab-case slug".
- Say when to omit an optional field.
- Skip Go implementation trivia ("backed by sync.Map") — it wastes prompt
  tokens and confuses the model.

```go
// Bad:  getter for the ts field, set by the ingest worker
// Good: Time the event occurred, as an RFC3339 string, e.g. "2026-07-09T14:00:00Z".
Timestamp string `json:"timestamp"`
```

## Required vs optional

Ordinary fields and `jsonschema.Nullable[T]` fields are required.
`jsonschema.Optional[T]` fields are omitted from `required` and must use
`json:",omitzero"`. Optional rejects JSON null; Nullable accepts null. Both
preserve present zero and empty values through their `Present` and `Value`
fields. Validate before unmarshaling when missing-vs-null matters, because plain
`json.Unmarshal` cannot distinguish those states for Nullable.

For OpenAI strict Structured Outputs, every property must be required. Use
Nullable for OpenAI's documented required-plus-null pattern. Do not use
Optional in a strict schema because it deliberately removes the property from
`required`.

Wrappers must be complete direct named field types. V1 Optional follows the
ordinary renderer's scalar and named scalar, struct, pointer, array/slice,
supported-ref, and registered-interface paths. V1 Nullable supports scalars,
registered enums, structs, pointers to structs, and structs registered with
`AsRef()`.

## Beyond flat structs

Enums (string consts and iota+Stringer), discriminated unions over interfaces,
custom discriminators, free-function registration, shared `$ref`/`$defs` via
`AsRef()`, and the full CLI/flag reference live in
[references/registration-api.md](references/registration-api.md). Read it
when a type uses enums, interfaces, or you need non-default generation flags.
By default, a struct type referenced from multiple places is inlined at every
call site; add `AsRef()` to its registration to render it once as a `"$ref"`
into `"$defs"` instead.

For concise, source-backed examples of optionality, enums, interface
discriminators, and shared `$defs`, read
[references/examples.md](references/examples.md). The snippets are generated
from compiling examples in this repository and checked for drift by the Go
test suite.

Known limitations (fail fast, don't fight them): no maps or recursive types;
registered interfaces support scalar fields and direct one-dimensional `[]I`
fields, but not fixed arrays, nested slices, named slice containers, or
Optional/Nullable interface slices; external package types are unsupported
except `time.Time`.

## Closeout checklist

- `go generate ./...` runs clean and a second run produces no diff.
- `go build ./...` and `go test ./...` pass.
- Generated `jsonschema/*.json` and `jsonschema_gen.go` are committed.
- Field doc comments read as LLM-facing descriptions.
- A pre-commit hook or CI check guards against schema drift.
