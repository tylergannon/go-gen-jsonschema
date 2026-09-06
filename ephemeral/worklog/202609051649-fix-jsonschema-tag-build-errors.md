# Fix `go build -tags jsonschema ./...` compile errors

## Task

User-supplied report of three pre-existing compile errors under
`go build -tags jsonschema ./...`:

1. `examples/test_options/schema.go:24` — `jsonschema.WithDescription(...)`,
   a function that doesn't exist anywhere in the codebase.
2. `examples/iota_global/schema.go:19` — `jsonschema.NewEnumType[Priority]()`
   where `Priority` is `type Priority int` (doesn't satisfy `~string`).
3. `examples/indirecttypes/schema.go` — `Schema()` methods declared on
   pointer-type receivers (`PointerToInt`, `PointerToSimpleInt`,
   `PointerToPerson`), which Go rejects outright.

Baseline `go test ./...` was green before starting (established per repo
CLAUDE.md gate). Reproduced all three errors exactly as reported via
`go build -tags jsonschema ./...`.

## Findings

friction: reproducing the report's build command surfaced a 4th error
(`examples/optionality/cmd/proof/main.go:90: ValidateJSON undefined`) not in
the report -> false positive of the diagnostic method itself, not a bug.
`examples/optionality/cmd/proof` is a consumer program that must build
against the *non*-tagged generated code (`jsonschema_gen.go`, which carries
`//go:build !jsonschema`); blanket-applying `-tags jsonschema` to `./...`
also (wrongly) tags that consumer, hiding `ValidateJSON`. Any CI check added
for this class of bug must scope `-tags jsonschema` to just the packages
containing `//go:build jsonschema` files (e.g. via `go vet -tags jsonschema`
on that package list), not a blanket `./...`.

decision: found `ephemeral/issue-73/finish/result.md` (a prior session's
closeout) explicitly recording that `examples/iota_global`,
`examples/template_rendering`, `examples/self_contained`, and
`examples/test_options` were *deliberately* left on broken/legacy
registration forms as "internal compatibility/limitation-demonstration
fixtures ... documenting a specific legacy edge case, not a style choice."
That characterization no longer holds: `NewEnumType[T]()`'s constraint is
`T ~string` (union_type.go), so `NewEnumType[Priority]()` (int-based) is a
**compile-time** generic-constraint violation — it never reaches the
generation-time panic the comments describe
(`internal/builder/gen_schema.go:454`, "dst.Expr is *dst.Ident, not
*dst.BasicLit"). The demo can't even be run to observe the documented
failure; it just fails to build, for an unrelated/undocumented reason. Since
the "intentional" framing no longer matches what actually happens, and the
user's own request in this session was to fix these compile errors, treated
the prior "leave broken" note as stale and fixed all three instead of
preserving it as-is — see PR description / final summary for the reasoning
surfaced to the user.

discovery: an equivalent fix for the same three files already exists,
unmerged, on branch `codex/issue-61-preview` (commit
`e1e6aaa98299ba2b8fc25460966943a913cd08bd`, "Repair invalid generation-tag
examples"). That branch predates the fluent `Declare(...)` API
(`declare.go`, PR #84) landing on `main`, so its fix uses the legacy
`NewJSONSchemaMethod`/`NewJSONSchemaFunc` forms, and fixes indirecttypes'
pointer-receiver problem with free functions
(`func PointerToIntSchema(PointerToInt) json.RawMessage`) rather than
deleting the types. Did not cherry-pick verbatim.

correction: user pushed back before I applied the indirecttypes free-function
workaround, asking whether these were stale/deprecated features that should
be stripped rather than fixed. Checked: `PointerToInt`/`PointerToSimpleInt`/
`PointerToPerson` are declared *only* to be registered as standalone schema
roots — `ComplexStruct` (the file's actual field-indirection showcase) uses
plain `*SimpleInt`/`*Person` instead, never these named pointer types. So the
free-function workaround would've kept dead, decorative code alive solely to
route around a Go language restriction (no pointer-receiver methods) that
nothing in the example actually needs. Deleted the three types, their
`Schema()` stubs, and their registrations instead of fixing the receiver
issue. Lesson: before patching a compile error in example/demo code, check
whether the broken symbol is actually load-bearing for the example's stated
teaching purpose — it may be scaffolding nobody uses.

friction: mid-task, `origin/main` gained one commit
(`8f5764f`, "Rename module to polytype and split root package (#89)") not
present when research started — import path changed from
`jsonschema "github.com/tylergannon/go-gen-jsonschema"` to plain
`"github.com/tylergannon/polytype"` across the repo. Branch had 0 local
commits (still at the old merge-base) with 2 dirty files, so committed WIP,
rebased onto `origin/main`, resolved the resulting import-path conflicts in
the same two files, and re-verified from scratch (re-read every file this
task touches — line numbers and even import syntax had shifted). Lesson:
when a long research phase precedes edits, re-check `git fetch` / diverged
`origin/main` before trusting anything read earlier in the session, even
within one sitting.

discovery: converting `iota_global`'s `Priority` registration to field-level
`.Enum(Task{}.Priority)` silently drops the description JSON Schema keyword
that a globally-registered enum type would carry (`Priority`'s doc comment,
"Priority represents task priority levels using iota", no longer appears on
the `priority` property). Confirmed this isn't fixable by moving the doc
comment onto the `Task.Priority` field instead — regenerated, still no
description. This matches a *known, already-documented* limitation from
`ephemeral/issue-73/finish/result.md` §4 ("field-level `.Enum` ... is not a
full replacement for package-level `NewEnumType[T]()`"), just observed here
as a description-loss instance rather than the sharing-scope instance that
doc already covered. Left it as-is and documented the trade-off inline in
`examples/iota_global/schema.go` rather than trying to force it — this is a
real, current generator gap, not something a workaround in example code
should paper over.

friction: the generator doesn't prune orphaned generated artifacts.
Deleting the `PointerToInt`/`PointerToSimpleInt`/`PointerToPerson`
registrations left `examples/indirecttypes/jsonschema/PointerTo*.json(.sum)`
on disk — `jsonschema_gen.go` stopped referencing them (correctly), but
`go generate` never deleted the now-dead files, and since the whole
`jsonschema` directory is `//go:embed`-ed, they'd have silently ridden along
in the embed.FS. Deleted them by hand. Worth a follow-up: `gen-jsonschema`
could delete `jsonschema/*.json(.sum)` files with no corresponding
registration after a generate run.

## Fixes applied

- `examples/test_options/schema.go`: dropped the `polytype.WithDescription`
  call (never existed); Team's Go doc comment already supplies the
  description per the project's comment-to-description convention.
- `examples/iota_global/schema.go`: replaced the global
  `NewEnumType[Priority]()` with the fluent
  `polytype.Declare(Task.Schema).Enum(Task{}.Priority)` (matching the fluent
  style `examples/enums_stringmode` already uses for its analogous
  `.StringerEnum` case). Documented the description-loss trade-off inline;
  dropped the stale "THIS WILL PANIC" claim since it no longer describes
  what actually happens (a compile-time generic-constraint error, not a
  generation-time panic).
- `examples/indirecttypes/{types.go,schema.go}`: deleted `PointerToInt`,
  `PointerToSimpleInt`, `PointerToPerson` entirely (types, `Schema()` stubs,
  registrations) rather than working around the pointer-receiver restriction
  — see "correction" above. Removed their orphaned
  `jsonschema/PointerTo*.json(.sum)` artifacts by hand since `go generate`
  doesn't prune them.

## Round 2: adversarial review findings and their fix

Launched an independent agent via `/adversarial-review` (round 1:
`ephemeral/reviews/202609051826-jsonschema-tag-build-fix-round-01.md`) against
the two commits above. Outcome: material findings remain.

decision: finding 1 was correct and I acted on it — deleting the pointer-root
types was wrong. The reviewer cited `internal/builder/testfixtures/entrypoints`
(free-function roots, tested) and `internal/builder/basic_test.go:114-133`
(asserts standalone JSON for named-pointer-type roots in
`testfixtures/indirecttypes`) as proof this is current, exercised
functionality, not dead scaffolding. Verified both citations directly before
acting. Restored `PointerToInt`/`PointerToSimpleInt`/`PointerToPerson` and
registered them via free functions (`func PointerToIntSchema(PointerToInt)
json.RawMessage` + `polytype.Declare(PointerToIntSchema)`), matching the
pattern `Declare`'s own doc comment describes.

friction/discovery: registering via free functions compiled fine under
`-tags jsonschema` and produced correct JSON, but `jsonschema_gen.go` (the
runtime, non-tagged output) emitted **no accessor at all** for these three
types — not a method, not a free function, nothing. Traced this to
`SchemaBuilder.SchemaMethods()` (`internal/builder/gen_schema.go`): it merges
method-root and free-function-root (`SchemaFuncs`) registrations into one
list for the `schemas.go.tmpl` Go-code template, but silently drops any entry
whose receiver's underlying type is a pointer or interface (comment: "filter
out invalid receiver base types"). That's correct for genuine method-root
entries (a real compile could never have produced one), but wrong for
free-function-root entries — the entire point of free-function registration
is to support exactly this case, and `schemas.go.tmpl` had no code path to
emit a free function instead of a method. Confirmed empirically: the
`test7-entrypoints` fixture's existing `FuncType`/`BuilderType` free-function
registrations get silently converted into *methods* in the generated output
(`func (FuncType) FuncTypeSchema()`), which only works because those types
aren't pointer-underlying — the moment the receiver type isn't
method-capable, generation for that entry just vanishes, with no error,
warning, or lint signal.

Fixed at the source: added `SchemaBuilder.SchemaFreeFuncs()` (free-function
roots that specifically *can't* be a method) alongside the now-correctly-named
`SchemaMethods()`, and a new `{{ range .SchemaFreeFuncs }}` block in
`schemas.go.tmpl` that emits a real free function matching the original
registration's name/signature. First attempt regressed
`test2-indirecttypes` (a fixture with the *same* invalid-pointer-receiver bug,
but registered as a genuine method-root, not free-function-root) — I'd
stopped filtering `SchemaMethods` entries entirely instead of only changing
where free-function entries with invalid receivers go. Root cause: I assumed
a method-root entry could never have an invalid receiver base, since the
source stub would have to compile first — false, because the *scanner*
(`internal/syntax`) only does AST-level parsing, not real type-checking, so
it happily records `func (PointerToIntType) Schema()` as a valid
`SchemaMethod` even though it would never survive a real `go build`.
Restored the original drop-on-invalid-receiver behavior for genuine
`SchemaMethods` entries (unchanged from before my edit); only
`SchemaFuncs`-sourced entries with an invalid receiver now get routed to the
new free-function emission path instead of being silently dropped.

discovery: `test2-indirecttypes` (`internal/builder/testfixtures/indirecttypes`)
*is* wired into `TestBasic`'s `cases` list (I misread this earlier and said
it wasn't — it's `testName: "test2-indirecttypes"` at basic_test.go:115) and
its `go build ./...` step (a **real** compile, inside a nested temp module)
should have been failing on `PointerToIntType`'s invalid-receiver-method bug
this whole time. It wasn't, because the *old* filter silently dropped that
entry from codegen too (same bug class, opposite symptom: instead of a
compile error, a silently-missing generated accessor). My fix only repairs
the free-function-root variant of this; the method-root variant in this
fixture is still silently broken exactly as before — untouched, out of
scope, filed as follow-up (see below). Correcting the record: my earlier
claim that this fixture "doesn't do a full type-check, so it doesn't hit this
bug" was wrong on the *reason* (right conclusion, wrong mechanism) — the real
reason it doesn't fail is the silent-drop filter, not an absence of
type-checking in the harness (the harness's `go build ./...` step is a real
compile).

Proof the fix actually works (not just "compiles"): extended
`internal/builder/testfixtures/entrypoints` (already wired into `TestBasic`
as `test7-entrypoints`, which does a real `go mod tidy` / `go generate` /
golden-diff / `go build ./...` / `go test ./...` round-trip in a nested
module) with `PointerFuncType *int` registered via
`polytype.Declare(PointerFuncTypeSchema)`, plus a new
`entrypoints_test.go::TestPointerFuncTypeSchemaCallable` that actually calls
`PointerFuncTypeSchema(nil)` and asserts real JSON content back. This is the
only test in the repo that would have caught the original gap (existing
fluent-declaration tests only do AST-level scanning + JSON string comparison,
never a real `go build`+`go test` of the generated output).

Also fixed finding 2 (no CI/local check builds `-tags jsonschema` source):
added a `just build-tagged` recipe and a matching CI step
(`.github/workflows/go.yml`) that build exactly the `examples/` packages
containing a `//go:build jsonschema` file, as explicit package arguments
(not `./...`) — this is what avoids the `examples/optionality/cmd/proof`
false positive: since that consumer program isn't itself jsonschema-tagged
and nothing in the explicit target list imports it, `go build` never visits
it. Deliberately scoped to `examples/` only, not the whole module: several
`internal/syntax/testfixtures/*` files (e.g. `fluent_field_mismatch`) are
*deliberately* invalid Go used to test scanner error paths, and a handful of
`internal/builder/testfixtures/*`/`test_run/*` dirs are separate nested Go
modules (own `go.mod`) already covered by `TestBasic`'s real
`go build`+`go test` harness — sweeping either into this new check would
either false-positive on intentionally-broken fixtures or duplicate existing
coverage.

Filed two follow-up issues (not fixed here, milestone 1.0):
- tylergannon/polytype#90 — `test2-indirecttypes` has the method-root variant
  of the same invalid-receiver bug and is unexercised in the sense that
  nobody ever looks at its actual generated output; needs the same
  free-function conversion (or the fixture should be deleted as superseded
  by `entrypoints`' pointer-root coverage).
- tylergannon/polytype#91 — `go generate` never deletes orphaned
  `jsonschema/*.json(.sum)` files for removed registrations (hit this
  directly: had to `rm` `examples/indirecttypes/jsonschema/PointerTo*.json`
  by hand after the delete-then-restore cycle above).

Round 2 review launched against the fixed state; see round-02 artifact for
outcome.

## Round 3: TypeScript omission fixed, --validate gap filed as follow-up

Round 2 (`ephemeral/reviews/202609051900-jsonschema-tag-build-fix-round-02.md`)
found `SchemaFreeFuncs()` roots were wired into the schema-accessor template
block only: `TypeDefinitions()` (`internal/builder/typegrammar.go`) walks
`SchemaMethods()` alone, so `--typescript` output silently omits a
free-function pointer/interface root; `--validate` has the same gap plus a
claimed (code-shape-plausible, not independently isolated) unused-import
compile break when such a root is a package's only registration.

correction (self, prompted by user pushback on scope/pace): distinguished
what's actually broken for anything shipping in this PR from a
plausible-but-unexercised gap. Verified the TypeScript claim directly against
the real CLI (`go run ./polytype gen -typescript ... -force` against a
disposable copy of the extended `entrypoints` fixture): confirmed
`PointerFuncType` was missing from `types.ts` before, present after. Fixed
with a 4-line addition to `TypeDefinitions()` (also range over
`SchemaFreeFuncs()`) — cheap, safe, no new API design, since TS lowering only
needs a type name, not a Go accessor shape. Added
`TestTypeDefinitionsIncludesFreeFunctionPointerRoot` in
`internal/builder/typegrammar_adapter_test.go` (in-process, no CLI
round-trip needed) as permanent regression coverage.

For `--validate`: verified `PointerFuncType` gets no generated `ValidateJSON`
in a real `-validate` run (confirmed), but did not independently reproduce
the unused-import compile break, and did not implement a fix — this needs a
real design decision (generate a free-function `ValidateJSON` equivalent, or
reject the combination outright per the project's #80 precedent) that
nothing in the actual repo currently depends on (no example combines a
free-function pointer/interface root with `-validate`). Filed as
tylergannon/polytype#92 rather than expanding this PR further.

## Verification

- `go generate ./...` per touched example dir; diffed `jsonschema/*.json` and
  `jsonschema_gen.go` against checked-in versions; confirmed only the
  intended files changed (`git status --porcelain` on the whole repo).
- `go build -tags jsonschema ./...` clean except the pre-existing, expected
  `examples/optionality/cmd/proof` false positive (see friction note above
  — not a bug, out of scope).
- `go build ./...`, `go vet ./...`, `go test ./...`, and
  `go run ./internal/cmd/doc-gen -check` all clean (full repo, post-rebase).
