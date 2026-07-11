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

## Generate

```text
go tool gen-jsonschema
go tool gen-jsonschema gen [flags]
  -pretty            indent schema JSON
  -target DIR        package to process (default: current directory)
  -no-changes        fail without writing schemas when schema JSON would change
  -force             rewrite unchanged output; incompatible with -no-changes
  -num-test-samples N  accepted for compatibility; currently has no effect
  --validate         generate ValidateJSON methods
```

The command without a subcommand is equivalent to `gen`.

## Scaffold a registration file

```text
go tool gen-jsonschema new [flags]
  -out FILE          output path; empty or -- writes to stdout
  -pkg NAME          package name override for stdout mode
  -methods LIST      required comma-separated Type=Method entries
  --validate         include ValidateJSON stubs
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
