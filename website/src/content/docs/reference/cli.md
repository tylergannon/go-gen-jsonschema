---
title: CLI reference
description: Commands and flags supported by gen-jsonschema.
---

When installed with the Go tool directive, invoke the CLI as
`go tool gen-jsonschema`.

## Generate

```text
go tool gen-jsonschema
go tool gen-jsonschema gen [flags]
  -pretty            indent schema JSON
  -target DIR        package to process (default: current directory)
  -no-changes        fail without writing when output would change
  -force             rewrite unchanged output; incompatible with -no-changes
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

Any non-empty `JSONSCHEMA_NO_CHANGES` value is equivalent to `-no-changes`.
This is the recommended form for hooks and CI because it applies through
existing `go generate` directives.
