# Review: `docs/design/issue-29-plan.md` (round 2)

## Exact prompt given

Here is my goal:

Produce an implementation-ready plan for GitHub issue 29 in
`/Users/tyler/src/go-gen-jsonschema`: repair exported/unexported struct-field
type traversal without changing embedded-field behavior; include the coupled
registered-local-interface correction exposed by the visibility fix; ensure
exported cross-package fields load and render real schemas rather than `{}`;
avoid recursively loading renderer-owned leaves such as `time.Time`; cover
newly reachable remote registered-enum dependency behavior; and define
regression, generation, and end-to-end proof strong enough to implement
safely.

Review the revised plan at:

`docs/design/issue-29-plan.md`

This is review round 2 after material revisions. Review the whole plan against
the repository and issue rather than limiting yourself to the prior changes.
Look for correctness problems, hidden dependencies, wrong ownership
boundaries, over-engineered components, unrequested features or layers,
missing tests, migration risk, proof gaps, and simpler implementation paths
that still satisfy the work order. In particular, verify that the remote-name
classification design is correct, the remote-interface non-goal is supported
by the repository contract, the fixture topology can prove the stated failure
modes, and the candidate-commit generation gate is executable as written.

Relevant primary evidence includes GitHub issue 29 and both comments,
`internal/syntax/scan_result.go`, `internal/syntax/node_wrappers.go`,
`internal/builder/gen_schema.go`, syntax and builder fixtures/tests,
`.github/workflows/go.yml`, and the issue 28 prerequisite design documents.

Do not edit product code or the proposed plan. You may inspect the repository
and GitHub. Write the exact prompt you were given and your findings to:

`ephemeral/reviews/202607101200-issue-29-plan-sonnet.md`

Label every finding with exactly one of:

- **critical:** must fix before proceeding.
- **bug:** demonstrable incorrect behavior, broken contract, race, or
  regression.
- **design:** architecture, boundary, scope, maintainability, or proof issue
  that is materially likely to cause problems.
- **nit:** small cleanup that should not block progress.

Use file/line references and concrete evidence when possible. Prefer real
findings over style preferences; do not merely approve the plan. If no
critical, bug, or design findings remain, say so explicitly.

## Method

Read the plan in full, then independently verified every falsifiable claim
against the repository at HEAD (`13421b6`, `go test ./...` green, matching
the plan's stated baseline), issue 29 and its two comments via `gh issue
view 29 --comments`, and `docs/design/issue-28-fable-review.md` /
`issue-28-optional-plan.md`. Traced the two concrete defects (visibility
inversion, interface-condition inversion) directly in
`internal/syntax/scan_result.go`, traced `renderSchema`'s `time.Time` special
case and external-package fallback in `internal/builder/gen_schema.go`, and
hand-simulated the "already-loaded remote dependency" code path against both
possible Go map iteration orders to confirm the claimed panic is real and
reachable under the fixture topology the plan proposes. Cross-checked fixture
and test names (`test2-indirecttypes`, `test3-enums`, `test5-interfaces`)
against `internal/builder/basic_test.go`, and compared the proof commands
against `.github/workflows/go.yml`.

## Findings

No critical or bug findings. Every specific, falsifiable technical claim in
the plan checked out against the repository (see "Verification notes"
below). Two low-severity design/nit items remain; neither should block
implementation.

### design: fixture topology for the already-loaded remote-enum test requires new subpackages the plan doesn't name

The plan's step 7 / design decision #4 correctly specifies the *conceptual*
topology needed to reproduce the already-loaded-dependency panic
deterministically: "a direct remote enum field plus a remote struct whose
traversal reaches the same enum package." I hand-traced `resolveTypes` in
`internal/syntax/scan_result.go:611-648` under both possible iteration orders
of `r.remoteTypes` and confirmed this topology panics either way (whichever
package — the direct-enum package or the struct's package — gets added to
the shared `r.deps` map second hits the buggy `if remote, ok :=
r.deps[pkgPath]; ok` branch and enqueues a zero `TypeSpec` from
`remote.LocalNamedTypes[typeName]`, since enum/interface names are never
populated into `LocalNamedTypes`, per `loadPackageInternal` at
`scan_result.go:435-446`). So the *design* is sound and does prove the
stated failure mode.

What's missing is the concrete realization: neither `test2-indirecttypes`
nor `test3-enums` currently has a fixture subpackage that is both
struct-bearing and enum-bearing in the required shape.
`indirecttypes/indirectsubpkg/types.go` contains only named-type
definitions, no struct types. `enums/enumsremote/enums.go` contains only the
registered enum, no struct types. Building the required topology means the
implementer must add at least one new struct-bearing subpackage (plus,
depending on which base fixture is chosen, possibly a new enum subpackage)
inside whichever fixture module is extended — none of which the plan names.
Given how granular the rest of the plan is (exact line numbers, exact
classification order, exact helper contracts), this is the one place an
implementer could construct a topology that fails to actually exercise the
already-loaded branch (e.g., by putting both references inside a single
resolve pass such that the shared package never becomes "already loaded"
before the second reference is seen) without realizing the regression wasn't
proven. Recommend the plan name the target fixture and sketch the two-package
shape (e.g., "add `indirecttypes/indirectstructpkg` with a struct field of
type `indirectsubpkg`-sibling enum, referenced independently as a second
holder field") so implementation and review can check the topology matches
the mechanism it's meant to prove.

### nit: `time.Time` canonical identity is asserted by convention only, not enforced

Design decision #3 requires "the predicate and the renderer's special case
must use the same canonical type identity." The renderer's existing check is
a literal string match at `internal/builder/gen_schema.go:670` (`node.Path ==
"time" && node.Name == "Time"`), and the plan's new traversal-side predicate
would live in `internal/syntax`, a separate package. Nothing proposed in the
plan makes the two checks share a single source of truth (e.g., an exported
`syntax.IsRendererOwnedLeaf(path, name)` used from both packages, or a
shared constant pair) — the guarantee is "keep two literals in sync by
convention." This is low-risk today (there's exactly one such leaf type),
but worth a one-line note in the plan or a shared helper so a future second
renderer-owned leaf (or a rename) can't silently desync the two checks. Not
blocking.

## Verification notes (supporting evidence for "no critical/bug findings")

- **Visibility inversion** — `internal/syntax/scan_result.go:583-609`
  (`skipField`) confirmed exactly as described: `IndexFunc` searches for a
  *lowercase* rune and returns `true` (skip) when none is found, i.e. skips
  exported fields and traverses unexported ones. Matches issue 29's repro
  verbatim.
- **Interface-condition inversion** — `scan_result.go:517-546`
  (`resolveTypeExpr`'s `*dst.Ident` branch), confirmed the `!ok` at line 526
  is backwards: returns success when the name is *not* a registered
  interface, and falls through to the position-bearing undeclared-type error
  when it *is* registered. Matches the issue's second comment and its
  `TestRegisteredInterfaceIdentifierResolves` repro against the
  `typescanner` fixture, which does contain `MarkerInterface` at
  `internal/syntax/testfixtures/typescanner/local_func_defs.go:5-7`.
- **Dormant registered-interface field** —
  `internal/builder/testfixtures/interfaces/types.go:29`
  (`FancyStruct.IFace TestInterface`) is exported and its type is a
  registered local interface; under the current bug this field is skipped
  before traversal ever reaches the (also inverted) interface branch,
  confirming the plan's claim that the existing fixture's golden generation
  doesn't currently exercise this path. `v1_interfaces_options/types.go:17`
  (`Owner.IF IFace`) is a second, independent instance of the same shape.
- **Already-loaded remote-name classification bug** — traced in full above;
  real and reachable, not hypothetical. `TypeSpec.Concrete` is a `*dst.TypeSpec`
  (`node_wrappers.go:62-65`), so a zero-value `TypeSpec` dequeued in
  `resolveTypes` (`scan_result.go:616-625`) dereferences a nil pointer at
  `ts.Concrete.Name.Name`, i.e. this is a real panic, not merely "wrong data."
- **`time.Time` leaf handling** — confirmed
  `internal/builder/gen_schema.go:670-682` special-cases `time.Time` at
  render time but `resolveTypeExpr` has no equivalent exclusion before
  `remoteTypes.addType` at `scan_result.go:544-546` and the `SelectorExpr`
  fallback at `scan_result.go:547-558`. No fixture currently exercises an
  exported `time.Time` struct field, so this is genuinely untested today,
  matching the plan's framing.
- **Remote-interface non-goal is grounded in the actual registry contract**
  — `Interfaces` is keyed by bare local name and populated only from
  `r.Pkg`'s own type declarations (`loadPackageInternal`,
  `scan_result.go:435-446`); a remote-declared, remote-registered interface
  reached through a field would resolve via `LocalNamedTypes` in the *remote*
  package (since the remote package's own `Interfaces` map wouldn't contain
  it unless registered there), landing on `*dst.InterfaceType` in
  `resolveTypeExpr`'s `default` case, which already errors
  (`scan_result.go:561-563`, "unhandled expression"). The plan's non-goal
  correctly leaves this as an explicit error path rather than silently
  mishandling it.
- **Dead `SliceExpr` branch** — `dst.SliceExpr` models slice *expressions*
  (`a[lo:hi]`), never type syntax; `resolveTypeExpr` only ever runs over type
  positions, so `scan_result.go:504-505` is confirmed unreachable. Safe to
  remove.
- **`StructField.Skip()` already correct** —
  `internal/syntax/node_wrappers.go:688-714` already uses
  `unicode.IsUpper` correctly and already does not skip
  `jsonschema:"ref=..."` fields, matching the plan's claim that only
  `skipField` (not `Skip()`) has the inverted logic, and that the two skip
  policies must stay distinct (embedded-field handling differs, ref handling
  differs).
- **`PropNames()`'s single-name branch bypasses the exported check** — real,
  at `node_wrappers.go:648-664` (the `case 1` branch returns the JSON tag
  name without checking export status), but every call site
  (`gen_schema.go:1014-1023`, `node_wrappers.go:329,394`) gates on
  `Skip()` first, which is already correct, so this is currently dead in
  practice. The plan's inclusion of `PropNames` in the shared
  named-visibility helper (design decision #1) is defense-in-depth, not a
  fix for a live bug — consistent with how the plan frames it.
- **Candidate-commit gate matches CI exactly** — `.github/workflows/go.yml`
  runs, in order: `go test ./...`, `go generate ./...`, `test -z "$(git
  status --porcelain)"`, `JSONSCHEMA_NO_CHANGES=1 go generate ./...`, `go
  test ./...`. The plan's proof-commands block reproduces this sequence
  verbatim (plus the focused test/golden runs first), and the "commit
  before running the clean-tree check" instruction is required for the gate
  to isolate codegen-caused diffs from pre-existing working-tree state —
  correct as written.
- **Fixture/test names are real** — `test1` through `test6` in
  `internal/builder/basic_test.go:79-150` include exactly
  `test2-indirecttypes`, `test3-enums`, and `test5-interfaces` as named in
  the plan's proof commands.
- **Skip-field / non-goal scope boundary with issue 30** — confirmed no
  `Optional[T]`/`Nullable[T]` types exist anywhere in the repository yet
  (`grep` for `Optional[`/`Nullable[` across all non-test `.go` files
  returned nothing), so the plan's non-goal framing of them as prospective
  issue-28 work is accurate, and `json:",omitzero"`'s empty-property-name
  bug (flagged in `issue-28-fable-review.md` §6.3 as
  `PropNames()`/`node_wrappers.go:648-655`) is correctly left to issue 30
  per the plan's non-goals, not silently dropped.
- **Round-1 feedback fully absorbed** — `issue-28-fable-review.md` §6.1-6.4
  (interface-inversion detonation, `time.Time` leaf, dead `SliceExpr`,
  `Skip()`/`PropNames()`/`skipField` divergence) each map one-to-one onto
  the plan's design decisions #2, #3, #5, and #1 respectively. No material
  round-1 finding was missed or only partially incorporated.

## Conclusion

No critical or bug findings. One design finding (fixture topology for the
already-loaded remote-enum regression needs a named, concrete package
layout) and one nit (no shared source of truth for the `time.Time` canonical
identity across packages) remain; neither is a correctness defect in the
plan as reasoned, and neither should block implementation starting. The
remote-name classification design, the remote-interface non-goal, and the
candidate-commit generation gate all check out against the repository
exactly as the plan states.
