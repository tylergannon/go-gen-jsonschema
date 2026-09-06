# Adversarial review: jsonschema-tag build fixes, round 02

## Review target

The entire current branch `claude/gallant-austin-9fd4e3` at `6be98ba`,
including the original fixes in `e157a3b`/`f61b872` and the round-01 response
in `6be98ba`, reviewed against the same authoritative goal: repair the three
reported `jsonschema`-tag compile errors in `examples/test_options`,
`examples/iota_global`, and `examples/indirecttypes` without stripping a
supported feature, and add proof that prevents recurrence.

## Evidence inspected

- Repository `AGENTS.md`/`CLAUDE.md`, `Declare`'s public contract in
  `declare.go`, and the full `origin/main...HEAD` diff and commit history
- `ephemeral/worklog/202609051649-fix-jsonschema-tag-build-errors.md` and
  `ephemeral/reviews/202609051826-jsonschema-tag-build-fix-round-01.md`
- All affected example source and generated artifacts; `.github/workflows/go.yml`;
  `justfile`; the scanner's method/free-function classification; builder root
  mapping, type-grammar lowering, Go-code template, and entrypoint fixtures
- Baseline `go test ./...`: exit 0
- `just build-tagged`, `go build ./...`, `go vet ./...`,
  `go run ./internal/cmd/doc-gen -check`, and `git diff --check`: exit 0
- A disposable copy of `internal/builder/testfixtures/entrypoints`, generated
  with the current CLI using `--typescript ... --typescript-barrel -force`:
  exit 0, but `types.ts` and `index.ts` contained `MethodType`, `FuncType`, and
  `BuilderType` only; registered `PointerFuncType` was absent
- The same disposable fixture generated with `--validate -force`: the output
  contained `PointerFuncTypeSchema`, but its compiled-schema initialization
  and generated `ValidateJSON` methods covered only the other three roots

Round 01's two findings are fixed at their original acceptance surfaces. The
three public generation-tag packages compile, the named-pointer roots and
their runtime accessors are restored, the new runtime test calls the generated
free function, and local/CI tagged-build gates now cover the reported class.

## Findings

### 1. issue — The new free-function-root path is wired only into one Go-template block, so other requested outputs silently lose the registered root

`SchemaFreeFuncs()` now preserves free-function registrations whose named
type has a pointer/interface underlying type (`internal/builder/gen_schema.go:749-760`),
but the rest of the generator still treats `SchemaMethods()` as the complete
root list. In particular, TypeScript lowering iterates only
`SchemaMethods()` (`internal/builder/typegrammar.go:17-40`), and validation's
presence check, compiled-schema variables/initialization, and validators do
the same (`internal/builder/gen_schema.go:661-669` and
`internal/builder/schemas.go.tmpl:61-98,145-176`). The added proof calls only
the ordinary generated schema accessor
(`internal/builder/testfixtures/entrypoints/entrypoints_test.go:8-27`), so it
cannot catch either omission.

This is observable with the checked-in entrypoint fixture. Generating it with
`--typescript` succeeds but omits `PointerFuncType` from both generated TS
files, even though the same registration emits its JSON Schema and Go
accessor. Generating with `--validate` likewise emits no validation surface
for that root. Worse, for a package whose only registration is such a free
function, `HasNonRenderedTypes()` is false while `Validate` still adds the
`bytes` and `jsonschema/v6` imports, so the generated Go file is
uncompilable due to unused imports.

Impact: the implementation now presents named-pointer free-function roots as
supported, but silently drops them from an explicitly requested output and
can break `--validate` generation. Establish one complete registered-root
classification and use it for TypeScript and mode preflight as well as schema
accessors. Because Go cannot attach `ValidateJSON`/`RenderedSchema` methods to
a named pointer or interface base, those modes must either provide an explicit
free-function contract or reject the unsupported combination clearly; silent
omission is not a valid result. Extend the real fixture proof to request and
inspect these modes.

## Outcome

material findings remain
