# Worklog: plan for issues #86, #87, #88

Session: planning only (no code changes). Baseline `go test ./...` green at
start (commit 8f5764f).

decision: Execute #86 -> #87 -> #88 sequentially on this one branch, one commit
per issue, single PR closing all three (source: user request "1,2,3 in
sequence but all in one session"; #88 depends on #87).

decision: Plan file lives at `ephemeral/issues-86-87-88/plan.md`; the
execution session should append to this worklog rather than start a new one.

discovery: `tests/typescript/node_modules` is absent in this worktree; the
TypeScript conformance suite needs `npm ci` in `tests/typescript` before
`npm test` can run locally.

discovery: `examples/test_options/schema.go` references
`polytype.WithDescription`, which does not exist, so that example does not
build under the `jsonschema` tag today (see vet check in plan). It must be
either fixed or dropped during #86's "all examples regenerate" step.

discovery: The scanner already rejects non-literal `Discriminator(...)`
arguments (`internal/syntax/fluent_expr.go` parseInterfaceNestedOptions);
#88's `SealedUnion[I](name)` can reuse that literal check and the builder's
existing property-name/collision validation.

discovery: No standalone "linter" subcommand exists (`polytype` has `gen` and
`new` only). The issues' "actual CLI and linter behavior" is read as: the
diagnostics that `polytype gen` emits, exercised through `polytype/main_test.go`
style CLI tests.

open_question: marker-enum type combined with field-level `.StringerEnum` on
the same field. Proposed: diagnostic (mirrors the existing
"cannot combine WithEnum and WithStringerEnum" rule).

## #86 execution

decision: `.StringerEnum` on a field of a marked enum type is allowed and
wins for that field (explicit name mode); no diagnostic. Source: issue #86
says Stringer mode "remains explicit through .StringerEnum and is not
changed", and the `v1_enums_stringmode` fixture / `stringer_enums` example
rely on one type being numeric in one field and named in another.

decision: value-mode enum fields no longer get a field plan at all; the type
sits in `ScanResult.Constants` (the map `NewEnumType` used to fill) and the
existing type-level rendering, typegrammar `named()` path, and TypeScript
projection handle every use (fields, slices, Optional/Nullable, roots). The
two field-plan validations that only existed for `.Enum` (pointer field,
`json:",string"` on a numeric field) are gone with it; they still apply to
`.StringerEnum`.

friction: staticcheck/golangci `unused` (U1000) flags every `enum()` marker
because nothing calls it -> each marker in this repo carries
`//lint:ignore U1000 ...` (which both linters honor); README, skill, and
website enums page tell consumers to do the same. A generator-emitted
reference to silence it automatically would be a follow-up, not a blocker.

discovery: TypeScript projection of a marked enum field is now a `Ref` to
the enum's own definition rather than an inline `Enum` node (same shape
`NewEnumType` already produced); `typegrammar_adapter_test` updated
accordingly. Rendered field descriptions now come from the field comment
where `.Enum` used to emit none.

decision: `examples/enums` now generates with `--validate` so the round-trip
test can prove out-of-set rejection through `ValidateJSON`; value-mode enums
have no generated codec and encoding/json alone accepts any string.

proof: `go test ./...` green; `tests/typescript` `npm test` 11/11 PASS;
`go generate` in every example; `just lint` steps (tidy, modernize, vet,
staticcheck, golangci-lint, goimports) clean; website api/index.md
regenerated with gomarkdoc.

## #87 execution

decision: inference errors are recorded per interface in
`ScanResult.InterfaceDiagnostics` and raised only when a generated schema
reaches the interface (direct field resolution in the builder), so an
unrelated non-sealed interface in the package never fails generation. Source:
issue #87 "the rule applies to the schema being generated".

decision: every disqualifying shape (inherited sealing method, invalid direct
candidate, non-struct candidate, zero variants, embedding-derived sealing
method) is a hard error, not a silent exclusion, matching "report
unsupported shapes clearly rather than silently omitting".

decision: variant order is go/types scope order (sorted by type name), not
declaration order. The `interfaces` fixture golden reordered accordingly
(PointerToTestInterface now first). Deterministic; not a wire change.

decision: the discriminator payload collision check lives in
`mapInterface` (JSON path) and rejects any variant payload property equal to
the discriminator property outright. The TypeScript grammar's older
"narrowing" allowance is stricter now on the JSON side; `typegrammar` test
fixture keeps `Created.Kind` as `json:"kind"` and no longer collides since
the discriminator is `type`.

decision: `union_codec` fixture rewritten around one inferred `Event` union:
per-field custom discriminators, custom wire values, the `Empty` (empty
discriminator) and `Unregistered` cases no longer exist as concepts and were
dropped; custom-hook coordination now uses `"type"`/type names.

decision: `TestLegacyDuplicateDerivedDiscriminatorRejectedBeforeWriting`
deleted (cross-package impls sharing a name cannot occur under same-package
inference). `TestLegacyHelpersUseResolvedPackageIdentity` kept: each remote
package's interface is inferred without a schema.go.

discovery: generated owner marshalers emit properties in sorted key order,
so `type` lands after `name` where `!kind` used to land first; the
`sealed_interface_slices` re-marshal assertion is now order-independent.

discovery: the TypeScript conformance "added variant" claim now toggles
`func (Renamed) event() {}` in types.go instead of an `Impl(...)` line; the
Unicode wire-value case (`create"雪`) had no inferred equivalent and is
covered by the projection lane only.

proof: `go test ./...` green; `npm test` 11/11 PASS; all examples
regenerate; optionality proof transcript regenerated; lint clean; website
api/index.md regenerated.

## #88 execution

decision: `SealedUnion[I](name)` markers are collected in the scanner's
marker loop and applied after type declarations are classified, so the
same-package rule, duplicates, literal shape, and property-name validation
are scanner diagnostics, while "not sealed" / "not an interface" reuse the
recorded inference diagnostic for the target.

decision: property-name validation is nonempty + valid UTF-8 (the only
validation the removed `Discriminator(...)` option performed was
"string literal"; the typegrammar validator adds the nonempty/UTF-8 rule).
`"!kind"` keeps working.

decision: `union_codec` fixture now declares `SealedUnion[Event]("!kind")`
so its custom-hook coordination, two-owner reuse (Envelope and Nested), and
value+pointer round trips all run on a custom property; the TypeScript
conformance fixture likewise restores `"!kind"`.

doc_bug: `ENUM_OPTIONS_TODO.md` (root) described the removed
WithEnum/NewEnumType pattern as broken -> deleted; the marker supersedes
everything it tracked.

proof: `go test ./...` green; `npm test` 11/11 PASS; every example
regenerates; goldens refreshed in place for union_codec and
v1_interfaces_options; lint clean; skill examples and website api/index.md
regenerated. origin/main had not moved during the session, so no rebase was
needed.

## Follow-up: enum marker and staticcheck U1000

correction: user rejected the `//lint:ignore U1000` approach for the enum
marker ("nothing calls it" read as untested; the marker IS covered by
builder, CLI, example round-trip, and TypeScript tests, but the lint
directive was still friction).

discovery: staticcheck marks a method as used when it satisfies the method
set of any *used* interface in the same package, so one
`var _ interface{ enum() } = *new(T)` clears every enum() marker in that
package. A constraint declared in polytype cannot do this: an unexported
method name is package-scoped, so `polytype.Enum interface{ enum() }` can
never be satisfied from another package, and build-tagged registrations are
invisible to the linter anyway.

decision: the generator now emits `var _ interface{ enum() } = *new(T)` for
every marked enum type in the generated package (schemas.go.tmpl,
`EnumMarkers` in template data). All lint directives removed. Packages that
never generate (remote enum fixtures, scanner fixtures, the TypeScript
fixture that is generated in a temp copy) carry the same one-line assertion
by hand, which is the documented instruction for consumers' shared enum
packages.

proof: staticcheck and golangci-lint clean with zero directives; go test
./...; npm test 11/11; all examples regenerated; goldens refreshed for
interfaces and v1_enums_stringmode.

## Rebase onto main (#94)

context: origin/main advanced by 1af8982 ("Fix jsonschema-tag compile
errors; harden free-function schema roots (#94)", 44 files) after this
branch's base (8f5764f). Rebased the branch's four commits
(c88e6db/84fad8a/b09ae78/0ebe397) onto origin/main with `git rebase
origin/main`, resolving conflicts in place and continuing with
`GIT_EDITOR=true git rebase --continue` (no squashing).

conflicts (2 files, both in the first commit, c88e6db):

- `examples/iota_global/schema.go`: main's #94 fix and this branch's
  c88e6db independently repaired the same pre-existing breakage (Priority,
  a pure iota enum, could not be registered under the old API). Main's fix
  used the fluent `Declare(Task.Schema).Enum(Task{}.Priority)` field-level
  form and documented a known trade-off: field-level enum registration
  loses Priority's doc comment as the schema description. This branch's
  fix uses the `func (Priority) enum()` marker instead, which has no such
  trade-off. Resolved by keeping this branch's version (marker-based,
  `var _ = polytype.NewJSONSchemaMethod(Task.Schema)` with no enum call).
- `examples/test_options/schema.go`: same shape of conflict. Main's #94
  fix replaced a broken `polytype.WithDescription(...)` option call on
  Team's registration with a plain call relying on Team's doc comment, and
  left Task/WorkItem enum registration commented out (Severity/WeekDay
  can't be global enums under the old API). This branch's c88e6db already
  made the identical Team fix (doc-comment-only) as a side effect of
  migrating everything to marker-based enums, and additionally registers
  Task and WorkItem directly since Status/Priority/Severity/WeekDay now
  each carry their own `enum()` marker. Resolved by keeping this branch's
  version in full.

Commits 84fad8a, b09ae78, and 0ebe397 applied without conflict.

tests ported/dropped: none. Main's #94 commit added no new tests that
exercised a removed API (`internal/builder/validate_free_func_test.go` and
its other new coverage target free-function schema roots and pointer-type
registration, both orthogonal to enum/union registration style, and both
still pass unmodified after the rebase).

post-rebase regeneration: `go build ./...` clean; `go generate ./...` in
every `examples/*/` directory; regenerating `examples/iota_global`
produced a real diff in `jsonschema/Task.json`/`.json.sum` (the description
comes back now that Priority's enum-ness is marker-based, not field-level)
committed separately as 8e4e675; no other example regenerated with a diff.
`examples/optionality/cmd/proof` transcript is byte-identical to
`proof/expected.json`; `go run ./cmd/proof` exits 0. `TestBasic` reported
no golden mismatches, so no builder fixture goldens needed refreshing.
`go generate .` at the repo root produced no diff.

gate results:
- `go test ./...`: all packages `ok` (internal/builder 28.7s, others
  sub-second to a few seconds), zero failures.
- `cd tests/typescript && npm ci --ignore-scripts --no-audit --no-fund &&
  npm test`: 11/11 PASS lines (fresh backend projection, CLI relative
  barrel, JSON Schema/TS enum parity, deterministic regeneration, stale-TS
  detection, barrel lifecycle, user-owned-output preservation, strict TS
  positive/negative/narrowing/enum/reference/composition, discriminated
  union exhaustiveness, sealed-interface exhaustiveness break, generated
  Go consumer build+test).
- `go vet ./...`: clean.
- `staticcheck ./...`: clean.
- `golangci-lint run ./...`: `0 issues.`

No lint directives were added anywhere during the rebase or the follow-up
regeneration commit.

final state: `git log --oneline -6` on
`claude/issues-86-87-88-plan-20be51` after the rebase and the regeneration
commit:

```
8e4e675 Regenerate iota_global Task schema after rebase onto main
515ab94 Reference each enum marker from generated code instead of lint directives
d530fea Declare a custom discriminator per sealed union with SealedUnion[I](name)
bf55de1 Infer sealed union membership from the sealing method; drop non-sealed unions
b657dd8 Infer enum types from a func (T) enum() marker method
1af8982 Fix jsonschema-tag compile errors; harden free-function schema roots (#94)
```

Pushed with `git push --force-with-lease origin claude/issues-86-87-88-plan-20be51`.

## CI generated-files drift

PR #93 `test-and-generate` failed at "Verify generated files are current"
(`test -z "$(git status --porcelain)"` after `go generate ./...`).

drifted files (reproduced locally; two consecutive `go generate ./...`
runs were byte-identical, so this is stale output, not non-determinism):
- `internal/builder/messages/jsonschema/GeneratedTestResponse.json`
- `internal/builder/messages/jsonschema/GeneratedTestResponse.json.sum`
- `internal/builder/messages/jsonschema_gen.go`

root cause: bf55de1 removed the explicit
`NewInterfaceImpl[AssertionValue](AssertNumericValue{}, AssertStringValue{},
AssertBoolValue{}, AssertType{}, AssertArrayLength{})` from
`internal/builder/messages/schema.go`, so the variants are now inferred by
`inferSealedUnion` (`internal/syntax/sealed.go`), which walks
`pkg.Scope().Names()` and emits them in sorted name order
(AssertArrayLength, AssertBoolValue, AssertNumericValue, AssertStringValue,
AssertType) instead of the old declaration order. That reorders the `anyOf`
branches, the marshal/unmarshal `switch` cases, and the content hash in the
generated helper names. The checked-in output was never refreshed because
the rebase verification above ran `go generate .` (root package only);
the `//go:generate` in `internal/builder/messages/assertions.go` is only
reached by `go generate ./...`. Nothing in `go test ./...` writes to
`internal/builder/messages/`. No example, `.json.sum`, or
`skills/polytype/references/examples.md` drifted.

fix: committed the regenerated artifacts unchanged (8682882, amended to add
this section). No generator change, no `.gitignore` change, no CI change.

new hook (`lefthook.yml`): `pre-push` gains `generate`, which runs
`go generate ./...` and fails, printing `git status --porcelain`, if the
tree is dirty afterwards. The group is `piped: true` with priorities
`generate`=1, `test`=2 so `generate` runs first and `test` is skipped on
failure; it is deliberately not in a `parallel: true` group because
`generate` rewrites files that `go test` reads. Note: a bare
`lefthook run pre-push` skips both commands with "no matching push files"
when HEAD equals the remote; use `--force` for a manual run.

hook verification:
- against the stale HEAD (before the fix), `lefthook run pre-push --force`:
  `generate` failed with `go generate output is out of date:` followed by the
  three `internal/builder/messages/...` paths; `test (skip) broken pipe`;
  exit 1.
- after the commit, clean tree: `✓ generate (5.11 seconds)`,
  `✓ test (8.68 seconds)`, exit 0, tree still clean.
- deliberately stale (edited a doc comment in
  `internal/builder/messages/assertions.go`): `generate` failed listing
  `assertions.go`, `GeneratedTestResponse.json`, and `.json.sum`; exit 1;
  restored with `git checkout -- internal/builder/messages`.

exact CI sequence from a clean tree, all exit 0: `git status --porcelain`
empty; `go test ./...` all `ok`; `go generate ./...`;
`test -z "$(git status --porcelain)"`; `JSONSCHEMA_NO_CHANGES=1 go generate
./...`; `go test ./...` all `ok`. `go vet ./...` clean, `staticcheck ./...`
clean, `golangci-lint run ./...` `0 issues.`; no lint directives added.
`tests/typescript` npm suite not rerun: no file under `examples/` and no
generator source changed.

## 2026-09-05 rc.8 follow-ups (branch `claude/rc8-followups`, from origin/main 12359fd = v1.0.0-rc.7)

Scope: (1) point the TypeScript+codec install/version prose at `v1.0.0-rc.8`
(the release that will follow this PR); (2) spell the generated enum-marker
assertion as `_ interface{ enum() } = <FirstConstant>` instead of
`*new(T)`.

decision: the version sentence now reads "requires `v1.0.0-rc.8` or newer:
`v1.0.0-rc.4` includes TypeScript declarations but predates generated owner
codecs, and releases before `v1.0.0-rc.7` predate the marker-based enum and
sealed-union registration" (README.md, docs/tutorial.mdx, llms.txt,
website getting-started.mdx, skills/polytype/SKILL.md). Rationale: the
`enum()` marker / `SealedUnion` API the docs show only exists from rc.7, so
rc.5/rc.6 pins would not compile against the shown registration code.
`grep -rn 'rc\.5'` (excluding node_modules/.git/ephemeral) is now empty.

decision: `schemaTemplateData.EnumMarkers` is `[]EnumMarker{TypeName,
Constant}` built by `(*SchemaBuilder).enumMarkers()` in
`internal/builder/gen_schema.go`: types sorted by name (deterministic),
constant = `Scan.Constants[type].Values[0].Name`. `ResolveEnum`
(`internal/syntax/enums.go`) walks `Pkg().Syntax` files then decls then
specs then names, so `Values[0]` is source declaration order; the zero-
constant case is rejected in `scan_result.go` before the map is populated,
so `Values[0]` always exists. Unexported first constants are fine (same
package); covered by `TestEnumMarkerAssertionUsesFirstConstant`.

doc_bug: README/enums.md/registration-api.md/llms.txt described the
assertion as `*new(T)` and said "once per package" -> now describe the
first-constant form, note the RHS must be a value of the marked type (value
receiver), and say one line per marked type.

friction: TestBasic golden files have no update flag; `AssertGoldenFile` is
diff-only and `t.Fatalf`s on the outer `t`, so one mismatch aborts the
remaining cases -> refreshed by looping: copy
`test_run/<case>/jsonschema_gen.go` over
`testfixtures/<fixture>/jsonschema_gen.go.golden` when the diff is confined
to the marker block, re-run, repeat. Only `interfaces` and
`v1_enums_stringmode` goldens carried markers; passed on iteration 1.
`test_run/` is tracked and was rewritten by the test run (test3/4/5/10).

friction: `just lint` runs `find . -name '*.go' -exec goimports -w` and
`modernize -fix ./...`, which reformatted 21 tracked historical files under
`ephemeral/` (codec-integration-consumer, typescript-generation proof runs)
-> reverted with `git checkout -- ephemeral/`; lint changed nothing else.
Consider excluding `ephemeral/` from the lint `find`.

Hand-written assertions updated to the first constant: tests/typescript
fixture (`Ready`, `Low`), internal/syntax comments fixture (`StringType1`),
typescanner fixtures (`Val1` x2), builder enumsremote (`EnumVal1`),
traversal remoteenum (`RemoteEnumFirst`). Historical `ephemeral/` proof
consumers intentionally left on `*new(T)`.

Proof (all from this worktree, clean tree after commit):
- `go test ./... -count=1`: all packages `ok`, no FAIL.
- `go generate ./...` run twice; second run left `git status --porcelain`
  identical (33 files changed, all intended). Examples regenerated:
  enums, iota_global, ref_types, self_contained, stringer_enums,
  template_rendering, test_options; internal/builder/messages unchanged
  (no marked enums).
- `just build-tagged`: exit 0.
- `just lint`: vet/staticcheck/govulncheck clean, golangci-lint `0 issues.`
- `cd tests/typescript && npm ci && npm test`: 11 PASS lines; proof retained
  at `ephemeral/typescript-generation/proof/run-dyHogF` (committed, 29
  files, its consumer/jsonschema_gen.go shows `= Low` / `= Ready`).
- `grep -rn '\*new(' --include='*.golden' --include='*_gen.go'` outside
  ephemeral: empty.

Out of scope, noted: website getting-started.mdx line ~81 still says
"`.Interface` and `.StringerEnum` registrations automatically generate ...
codecs" although `.Interface` was removed in rc.7 (doc_bug: stale API
name -> should read `.StringerEnum` registrations and inferred sealed
unions).

## Follow-up: stale removed-API prose in docs (rc.8 followups branch)

Task: fix the getting-started.mdx sentence flagged above, then sweep .md/.mdx/.txt
for `.Interface(`, `WithInterface`, `NewInterfaceImpl`, `NewEnumType`, `WithEnum`,
`.Enum(`, `Discriminator(` presented as current API. Docs only; no Go or
generated files touched.

Commands:
- `go test ./...` baseline before edits: all ok.
- `grep -rnE '\.Interface\(|WithInterface|NewInterfaceImpl|NewEnumType|WithEnum|\.Enum\(|Discriminator\(' --include='*.md' --include='*.mdx' --include='*.txt' --exclude-dir=node_modules --exclude-dir=.git --exclude-dir=ephemeral .`
  plus a broader pass for backticked `.Interface` / `.Enum` without a paren
  (the getting-started sentence itself only matched the broader form).
- Checked `internal/builder/gen_schema.go` (`adapted := config.UseStringer && underlying == enumUnderlyingInteger`):
  only sealed-union fields and `.StringerEnum` integer fields produce owner
  codecs; marked enums in value mode do not. Wording was chosen to match that.

Changed: website/src/content/docs/getting-started.mdx,
website/src/content/docs/reference/cli.md, docs/tutorial.mdx, llms.txt (one
`.Enum` mention), AGENTS.md (registration list/root package/discriminator
lines described the removed fluent API as current), docs/design/v1-roadmap.md
(disposition said `WithEnum` and interface options "are implemented").

Left alone, deliberately: migration notes in README.md, llms.txt,
features/enums.md, features/interfaces.md, skills/polytype/references/registration-api.md
(they describe the APIs as removed); docs/internal-dev-notes.md,
docs/design/issue-29-plan.md, .agent/memory/current.mdx, prompts/description.md
(historical snapshots / stale prompt artifact, headered as such or not user
docs). Flagged for a separate change: docs/spec/v1.md still specifies
`Impl("created", Created{})` explicit wire values and "legacy derived names"
in the #57 union decisions; reconciling the contract with rc.7 semantics is a
spec amendment, not a prose fix.

## Follow-up: v1 spec amendment and lint scope (branch claude/spec-and-lint-fixes)

Task: (1) amend docs/spec/v1.md so the union/enum contract states the rc.7
semantics (inferred sealed unions, type-name wire values, `SealedUnion[I]`,
`func (T) enum()` marker); (2) stop `just lint` reformatting tracked files
under ephemeral/. Branched from origin/main 622db5e (v1.0.0-rc.8).

Commands:
- `go test ./...` baseline: all ok.
- `find ephemeral -name '*.go' -exec goimports -l {} \;` before the change:
  21 files would be rewritten, all under ephemeral/. Zero elsewhere.
- Every Go lint step already used `./...`, which stops at ephemeral's nested
  go.mod files; only the `find . -name '*.go' -exec goimports -w` reached in.
  No .golangci.yml or staticcheck.conf exists, so nothing else to exclude.
- New find: `find . \( -path ./.git -o -path ./ephemeral -o -name node_modules \) -prune -o -name '*.go' -exec goimports -w {} +`
  selects the same 312 files as the explicit-exclusion count, none under
  ephemeral/ or website/.
- Gate: `just lint` then `git status --porcelain` -> only docs/spec/v1.md
  and justfile; `go generate ./...` -> tree unchanged; `go test ./...` ok.

Spec edits (all in place, with a dated amendment note after the status
paragraph): semantic-equality bullet, enum/interface capability rows,
generated-method ownership paragraph and its `Impl("created", ...)` example,
all of "Union decisions for #57", the custom-hook table (`"type":"Created"`),
the enum-decisions opening, and the 1.x compatibility line.

doc_bug: docs/spec/v1.md capability row linked internal/builder/interface_options_test.go, deleted in #93 -> now links sealed_union_test.go and sealed_union_discriminator_test.go.
decision: comments belong above a just recipe, not inside its body; just echoes body comment lines as if they were commands, so the lint output showed the comment text.
friction: zsh parses a bare `=====` token in a command line as a command name ("==== not found") -> quote separators in shell one-liners.
