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
