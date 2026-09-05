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

## Verification

- `go generate ./...` per touched example dir; diffed `jsonschema/*.json` and
  `jsonschema_gen.go` against checked-in versions; confirmed only the
  intended files changed (`git status --porcelain` on the whole repo).
- `go build -tags jsonschema ./...` clean except the pre-existing, expected
  `examples/optionality/cmd/proof` false positive (see friction note above
  — not a bug, out of scope).
- `go build ./...`, `go vet ./...`, `go test ./...`, and
  `go run ./internal/cmd/doc-gen -check` all clean (full repo, post-rebase).
