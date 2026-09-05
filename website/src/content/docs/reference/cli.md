---
title: CLI reference
description: Commands and flags supported by gen-jsonschema.
---

## Installation

Use Go's tool directive to pin the generator version in `go.mod`:

```bash
go get -tool github.com/tylergannon/go-gen-jsonschema/gen-jsonschema@latest
```

Invoke the pinned CLI as `go tool gen-jsonschema`.

The CLI generates schemas. Validation and selected JSON codecs/YAML input methods
are optional generated capabilities; schema generation does not provide a
general-purpose Go codec.

## Generate

```text
go tool gen-jsonschema
go tool gen-jsonschema gen [flags]
  -pretty            indent schema JSON
  -target DIR        package to process (default: current directory)
  -no-changes        fail without writing when schemas or requested TypeScript output would change
  -force             rewrite unchanged output; incompatible with -no-changes
  --validate         generate validation methods for the selected formats
  --formats MODE     decoding and validation: json (default) or both
  --typescript DIR   generate structural TypeScript declarations in DIR
  --typescript-barrel
                     also generate index.ts type-only exports; requires --typescript
```

The command without a subcommand is equivalent to `gen`.

## Scaffold a registration file

```text
go tool gen-jsonschema new [flags]
  -out FILE          output path; empty or -- writes to stdout
  -pkg NAME          package name override for stdout mode
  -methods LIST      required comma-separated Type=Method entries
  --validate         include validation stubs for the selected formats
  --formats MODE     validation stubs: json (default) or both
  --generate         run go generate ./... after writing
```

Example:

```bash
go tool gen-jsonschema new \
  -out schema.go \
  -methods 'Person=Schema,Address=Schema' \
  --validate \
  --generate
```

## Environment

Any non-empty `JSONSCHEMA_NO_CHANGES` value is equivalent to `-no-changes` and
applies through existing `go generate` directives. It guards schema JSON and
every requested TypeScript artifact without writing those destinations, but
generation can still update `jsonschema_gen.go` when they are unchanged. In CI,
follow generation with `test -z "$(git status --porcelain)"` to verify tracked
and untracked generated files.

## Go codecs and TypeScript declarations

One directive can generate validation, Go owner codecs selected by field
registrations, and TypeScript declarations:

```go
//go:generate go tool gen-jsonschema --validate --typescript web/src/generated --typescript-barrel
```

`WithInterface` and `WithStringerEnum` registrations cause the containing Go
struct's JSON methods to be generated automatically; there is no codec flag.
The TypeScript output provides static declarations only, with no runtime decoder
or validator. Validate untrusted values in the TypeScript application, and call
the generated Go `ValidateJSON` method before `json.Unmarshal`. Issue
[#71](https://github.com/tylergannon/go-gen-jsonschema/issues/71) tracks broader
executed Go/JavaScript transport proof.
