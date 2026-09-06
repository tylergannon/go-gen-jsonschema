# Adversarial review: jsonschema-tag build fixes, round 01

## Review target

Commits `e157a3b5797d68646c973fa336e3d0d441e572eb` and
`f61b872767bac5e5bfaf18b6dfa97647e181cf82` on
`claude/gallant-austin-9fd4e3`, reviewed against the stated goal of repairing
the three `jsonschema`-tag compile errors in `examples/test_options`,
`examples/iota_global`, and `examples/indirecttypes`, including the user's
mid-task direction to strip a feature only if it was stale, deprecated, or
canceled.

## Evidence inspected

- `ephemeral/worklog/202609051649-fix-jsonschema-tag-build-errors.md`
- Clean pre-review working tree; `git log -p -2`; full diff from
  `origin/main...HEAD`; prior versions of the affected files
- `declare.go`, `union_type.go`, `internal/syntax/fluent_expr.go`, the fluent
  declaration scanner tests, the indirect-type builder fixture/test, all
  affected example source, generated Go, JSON, and checksum files, and the Go
  CI workflow
- Baseline `go test ./...`: exit 0
- `go build -tags jsonschema ./examples/test_options
  ./examples/iota_global ./examples/indirecttypes`: exit 0
- `go build ./...`, `go vet ./...`, and
  `go run ./internal/cmd/doc-gen -check`: exit 0
- `go build -tags jsonschema ./...`: exit 1 only at
  `examples/optionality/cmd/proof/main.go:90` because the tag excludes the
  generated `ValidateJSON` method, matching the worklog's disclosed
  out-of-scope limitation

The `test_options` removal fixes its undefined symbol and retains Team's
description through its Go doc comment. The `iota_global` registration now
compiles and its generated `priority` field contains enum values `0,1,2,3`.
The worklog accurately discloses the resulting loss of Priority's description.

## Findings

### 1. issue — The indirecttypes change strips a supported root-schema case, not a stale or canceled feature

Commit `f61b872` deletes `PointerToInt`, `PointerToSimpleInt`, and
`PointerToPerson`, their registrations, their runtime `Schema` methods, and
their generated schemas. The worklog calls these types dead because
`ComplexStruct` does not use them as fields and says registering a named
pointer root is "impossible in Go." That conclusion conflates the invalid
method-receiver spelling with the generator feature.

`Declare` explicitly supports a free function taking `T` as its sole argument
(`declare.go:12-20`), and the scanner has a passing free-function-root case
(`internal/syntax/scanner_test.go:250-260`). More importantly, the builder's
indirect-types fixture defines named pointer roots such as
`PointerToIntType` and `PointerToNamedType`
(`internal/builder/testfixtures/indirecttypes/types.go:12-16`), while its main
generation test expressly requires their standalone JSON schemas
(`internal/builder/basic_test.go:114-133`). This is current, exercised
functionality. The public example's own stated purpose is to demonstrate
forms of indirection (`examples/indirecttypes/types.go:5-6`); standalone
registered roots do not need to reappear as `ComplexStruct` fields to be
load-bearing examples.

Impact: the compile error is removed by deleting three public examples and
their generated API/output, reducing example coverage of a feature the
generator still supports. This does not satisfy the user's conditional
"strip if stale/deprecated/canceled" direction because none of those
conditions was established. Restore the named pointer types and register
free-function entrypoints such as `func PointerToIntSchema(PointerToInt)
json.RawMessage` with `polytype.Declare(PointerToIntSchema)`, then regenerate
the deleted artifacts. That fixes Go's receiver restriction without removing
the supported schema roots.

### 2. issue — No automated check protects the exact tagged-build regression

Neither commit adds a test or CI check that compiles generation-tag source.
The current workflow runs `go test ./...`, `go generate ./...`, and then
`go test ./...` again (`.github/workflows/go.yml:29-42`), but never builds
with `-tags jsonschema`. The clean baseline test run demonstrates the gap:
the ordinary suite was green while all three reported tagged-source errors
existed. Generation also does not provide this protection because its AST
pipeline can consume source forms that the Go compiler rejects.

Impact: the same class of undefined marker, invalid generic instantiation, or
illegal receiver can return without failing the repository's required gate.
Add a permanent compile regression check for packages containing
`//go:build jsonschema` registration files. It must scope the package set so
the known `examples/optionality/cmd/proof` consumer/tag incompatibility does
not turn the check into the blanket-command false positive documented above.

## Outcome

material findings remain
