# Adversarial review: jsonschema-tag build fixes, round 05

## Review target

The entire current branch `claude/gallant-austin-9fd4e3` at `b5be7aa`,
including the original example repairs, the tagged-build gate, the
free-function-root generation and TypeScript changes, and the responses to
review rounds 01 through 04. I reviewed it against the same authoritative
contracts used in those rounds: repair the reported generation-tag failures
without discarding supported registrations, preserve the documented legacy and
fluent registration APIs, fail explicitly instead of emitting partial or
misleading output, and retain behavioral proof at the claimed acceptance
surface.

## Evidence inspected

- Repository `AGENTS.md`; the public `Declare`, `NewJSONSchemaFunc`,
  `NewJSONSchemaBuilder`, and `NewInterfaceImpl` contracts; README registration
  and TypeScript guidance
- Full `origin/main...HEAD` diff and all six branch commits; scanner marker
  classification, builder root classification, Go template, TypeScript
  lowering, affected examples, generated artifacts, fixture harness, CI
  workflow, and `justfile`
- `ephemeral/worklog/202609051649-fix-jsonschema-tag-build-errors.md` and review
  rounds 01 through 04
- Current GitHub issues #90, #91, and #92
- Fresh `go test ./...`, `just build-tagged`, `go build ./...`, `go vet ./...`,
  `go run ./internal/cmd/doc-gen -check`, and
  `git diff --check origin/main...HEAD`: all exited 0

Round 04's two implementation findings are fixed at their immediate code
paths: legacy registered interfaces are now classified as invalid method
receiver bases, rendered free-function roots are rejected before rendering,
and validation rejects only the non-rendered free-function roots for which it
cannot generate `ValidateJSON`. The original three example packages continue
to compile with the generation tag.

## Findings

### 1. issue — `SchemaFreeFuncs` changes the contract of legacy no-argument schema builders

`NewJSONSchemaBuilder[T]` accepts a `SchemaFunction`, whose declared type is
`func() json.RawMessage` (`union_type.go:13,30-33`). The scanner appends every
such marker to `Scan.SchemaFuncs` without retaining a separate output category
(`internal/syntax/scan_result.go:427-434`). The new `SchemaFreeFuncs()` then
selects every `SchemaFuncs` entry whose named root has a pointer/interface
underlying type, without distinguishing `NewJSONSchemaBuilder` from an actual
one-argument free-function registration
(`internal/builder/gen_schema.go:756-767`). Finally, the new template emits all
of those entries as `func Name(Type) json.RawMessage`
(`internal/builder/schemas.go.tmpl:131-143`).

That signature is valid for `Declare(fn)` and `NewJSONSchemaFunc[T](fn)`, whose
input function takes `T`; it is not valid for `NewJSONSchemaBuilder[T](fn)`.
For example, the supported source
`NewJSONSchemaBuilder[PointerRoot](BuildSchema)`, where
`type PointerRoot *int` and `func BuildSchema() json.RawMessage`, now generates
`func BuildSchema(PointerRoot) json.RawMessage`. Runtime callers using the
registered no-argument API no longer compile. Reusing one builder function for
two invalid-receiver roots is worse: the template emits two package-level
functions with the same name and the generated package fails with a
redeclaration error. This contradicts the repository's statement that
`NewJSONSchemaBuilder[T](fn)` remains supported and is unaffected
(`README.md:544-548`).

Impact: the branch fixes silent loss for one kind of `SchemaFuncs` entry by
silently inventing an incompatible API for another supported kind, and can
produce uncompilable generated Go from valid generation-tag input. Preserve
registration provenance when classifying outputs. Because the ordinary
builder behavior relies on generating a method per root, the safest behavior
for an invalid-receiver `NewJSONSchemaBuilder` registration is an actionable
generation-time rejection unless a separate, collision-safe builder contract
is deliberately designed. Add a fixture covering this exact marker and named
pointer/interface base.

### 2. issue — The retained interface-root test does not perform the compile-and-call proof it claims

`TestFreeFunctionRootForRegisteredInterfaceCompilesAndRuns` says both in its
name and comment that it compiles generated code and calls the resulting
accessor (`internal/builder/validate_free_func_test.go:57-65`). Its body only
runs the generator, reads `jsonschema_gen.go`, and checks for two source
substrings (`internal/builder/validate_free_func_test.go:96-101`). It never runs
`go build` or `go test`, never adds a caller, never invokes `ValueSchema`, and
never inspects the emitted union schema. Because the fixture is created under
`t.TempDir()` as a nested module, the repository's outer `go test ./...` does
not compile or execute that generated package for it.

The worklog says a disposable fixture was independently compiled and called,
but retains no command output or artifact for that assertion. There is also no
interface-root `--validate` case, despite round 04 explicitly requiring that
behavior be asserted; the only validation test uses a named pointer
(`internal/builder/validate_free_func_test.go:29-55`).

Impact: the claimed proof can stay green if the generated interface accessor
has invalid imports or syntax, reads the wrong embedded file, returns the wrong
schema, or stops being callable. Replace the substring-only assertion with a
real generated-module build/test containing a runtime call and schema-content
assertions, and exercise the interface-root validation rejection. Keep the
focused string assertions only as supplementary diagnostics.

## Outcome

material findings remain
