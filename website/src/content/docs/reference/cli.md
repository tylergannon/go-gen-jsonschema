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

The CLI generates schemas. Validation and selected JSON/YAML decoding methods
are optional generated capabilities; schema generation does not provide a
general-purpose Go codec.

## Discover and inspect

```sh
go tool gen-jsonschema version --json
go tool gen-jsonschema inspect --json --target ./models Request Response
```

Read `--help` before using these subcommands on an older installation.
`version` reports executable identity and installed capabilities. `inspect`
reports each registered root's schema, JSON encode/decode, validation, and YAML
support separately. Omit type names to inspect all registered roots.

Inspection does not write source/generated/module files or invoke user hooks.
External Go caches may be populated. `--json` emits one versioned JSON result
to stdout; human mode writes to stderr. Exit 0 means supported/help, 2 invalid
request/source, 3 unsupported/unknown, and 1 internal/toolchain error.

See the [machine result contract](https://github.com/tylergannon/go-gen-jsonschema/blob/main/docs/agent-cli.md)
for field definitions and compatibility rules. Both commands support `--help`,
including structured `--json --help`.

## Generate

```text
go tool gen-jsonschema
go tool gen-jsonschema gen [flags]
  -pretty            indent schema JSON
  -target DIR        package to process (default: current directory)
  -no-changes        fail without writing schemas when schema JSON would change
  -force             rewrite unchanged output; incompatible with -no-changes
  --validate         generate validation methods for the selected formats
  --formats MODE     decoding and validation: json (default) or both
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
applies through existing `go generate` directives. It guards schema JSON, but
generation can still update `jsonschema_gen.go` when schemas are unchanged. In
CI, follow generation with `test -z "$(git status --porcelain)"` to verify
tracked and untracked generated files.
