# Review: `docs/design/issue-29-plan.md` (round 3, final)

Reviewer: Claude (Sonnet), 2026-07-10. Repository at `13421b6`, working tree
otherwise clean except the untracked plan and `ephemeral/reviews/` files
(`git status` re-verified during this review). Baseline `go test ./...`
previously confirmed green on this commit by both prior review rounds; not
re-run here since no product code changed since those runs.

## Exact prompt given

Here is my goal:

Produce an implementation-ready plan for GitHub issue 29 in
`/Users/tyler/src/go-gen-jsonschema`: repair exported/unexported struct-field
type traversal without changing embedded-field behavior; include the coupled
registered-local-interface correction; load and render real cross-package
schemas; treat `time.Time` consistently as a renderer-owned leaf; safely
classify already-loaded remote registered enums; and provide focused,
generation, and repository-wide proof.

Perform the third and final consensus review of:

`docs/design/issue-29-plan.md`

Review the entire plan against the repository and issue, not merely the latest
edits. Look for correctness problems, hidden dependencies, wrong ownership
boundaries, over-engineering, scope drift, missing tests, migration risk,
proof gaps, and simpler paths. Specifically verify that the now-concrete
`test11-traversal` fixture layout really exercises the already-loaded remote-
enum path under either package visitation order, and that sharing
`syntax.IsTimeType` between scanner and renderer respects package boundaries.

Do not edit product code or the plan. You may inspect the repository and
GitHub. Write the exact prompt you were given and your findings to:

`ephemeral/reviews/202607101209-issue-29-plan-sonnet-final.md`

Label every finding with exactly one of:

- **critical:** must fix before proceeding.
- **bug:** demonstrable incorrect behavior, broken contract, race, or
  regression.
- **design:** architecture, boundary, scope, maintainability, or proof issue
  that is materially likely to cause problems.
- **nit:** small cleanup that should not block progress.

Use file/line references and concrete evidence when possible. Prefer real
findings over style preferences. If no critical, bug, or design findings
remain, say so explicitly. This is the third review round, so record any
remaining dissent clearly rather than requesting another round.

## Method

Read the full plan and both prior reviews
(`ephemeral/reviews/202607101143-issue-29-plan-fable.md`,
`ephemeral/reviews/202607101200-issue-29-plan-sonnet.md`) to establish what
changed and why. Re-read issue 29 and both of its comments via `gh issue view
29 --comments`. Independently re-derived, from current source rather than
from the prior reviews' word, every load-bearing claim in the plan, with
particular depth on the two items the prompt called out by name.

### Verifying the `test11-traversal` topology under both visitation orders

Hand-traced `internal/syntax/scan_result.go` end to end for the plan's step-7
fixture (root `TraversalHolder{Remote remotestruct.RemoteStruct, Status
remoteenum.RemoteEnum, When time.Time}`, `remotestruct.RemoteStruct{Status
remoteenum.RemoteEnum}`, `remoteenum.RemoteEnum` registered via
`NewEnumType[RemoteEnum]()`):

- `ScanResult.deps` is `map[string]ScanResult` (`scan_result.go:218`), and
  `newScanResult` (`scan_result.go:269-283`) stores whatever `deps` map it is
  given verbatim — every nested `ScanResult` in one `LoadPackage` call
  (`scan_result.go:252-257`) shares the *same* underlying map object, since
  Go map values are reference types and the map is threaded through
  unchanged (`remote = newScanResult(pkgs[0], r.deps)` at `scan_result.go:640`
  reuses `r.deps`, never a copy). So a write to `r.deps[pkgPath]` from deep
  inside a nested remote resolution is visible to every other `ScanResult` in
  the tree, including the root, immediately.
- Root's `resolveTypes` (`scan_result.go:611-648`) iterates
  `r.remoteTypes`, a plain Go map, in nondeterministic order. It contains
  exactly two keys here: `remotestruct` → `{RemoteStruct}` and `remoteenum` →
  `{RemoteEnum}` (added by two `addType` calls from the two struct-field
  idents, `scan_result.go:545`).
  - **Order A (`remotestruct` first):** not yet in `r.deps` → fresh-load
    branch (`scan_result.go:637-644`): `remotestruct` is loaded and scanned.
    Its own `RemoteStruct.Status remoteenum.RemoteEnum` field causes its
    *own* `resolveTypeExpr` to call `addType("remoteenum", "RemoteEnum")` on
    its own `remoteTypes`, and its own trailing `resolveTypes` loop
    (`scan_result.go:627-646`, now running one level down) finds `remoteenum`
    not yet in the *shared* `r.deps` and fresh-loads it, storing it into
    `r.deps["remoteenum"]` (shared). Control returns to the root loop, which
    now visits `remoteenum` for the *second* logical time: `r.deps["remoteenum"]`
    already exists → **already-loaded branch** (`scan_result.go:628-633`):
    `remote.LocalNamedTypes["RemoteEnum"]` misses (registered enums are
    excluded from `LocalNamedTypes`, `scan_result.go:435-446`, landing in
    `Constants` instead), so today's code appends a zero-value `TypeSpec{}`
    to `remote.resolveQueue`; `resolveTypes` dequeues it and dereferences
    `ts.Concrete.Name.Name` (`scan_result.go:619`) on a nil `*dst.TypeSpec`
    (`TypeSpec.Concrete` is `T`/`*dst.TypeSpec`, `node_wrappers.go:17`,
    `267`) → panic. This is the **root-level** manifestation.
  - **Order B (`remoteenum` first):** not yet in `r.deps` → fresh-load:
    `remoteenum` loads cleanly (its `RemoteEnum` typesToMap entry resolves via
    the `Constants` branch, `scan_result.go:479-481`, `continue`s, nothing
    queued), and is stored into shared `r.deps["remoteenum"]`. Root then
    visits `remotestruct`: not yet in `r.deps` → fresh-load: `remotestruct`
    is scanned, and *its* `Status` field again calls
    `addType("remoteenum", "RemoteEnum")` on *its own* `remoteTypes`. Its own
    trailing `resolveTypes` loop now checks the *shared* `r.deps` for
    `remoteenum` — already present (added by root a moment earlier) — so it
    takes the **already-loaded branch** itself, hitting the identical zero-
    `TypeSpec` panic. This is the **nested (`remotestruct`-scope)**
    manifestation of the same defect.

  Both orders panic; only the call frame in which the already-loaded branch
  fires differs. This confirms the plan's fixture topology is not merely
  "plausibly sufficient" but mechanically forces the already-loaded path to
  fire deterministically regardless of `range` order over `r.remoteTypes`,
  exactly as design decision #4 and step 7 claim, and exactly what round 2's
  design finding asked the plan to make concrete. The round-2 finding is
  fully resolved, not just reworded.
- Also confirmed the fixture's structural feasibility against the existing
  fixture-authoring convention: subpackages like `remoteenum`/`remotestruct`
  do **not** need their own `go.mod` — `internal/builder/testfixtures/enums`
  and its `enumsremote` subpackage, and `internal/builder/testfixtures/indirecttypes`
  and its `indirectsubpkg` subpackage, both live under one fixture-root
  `go.mod` today, matching what step 7 asks for. `remotestruct` needs no
  `schema.go` (plain non-entrypoint types are discovered purely through
  traversal, as `indirectsubpkg` proves); `remoteenum` needs exactly the
  `NewEnumType[RemoteEnum]()` registration the plan specifies, matching
  `enumsremote/schema.go`'s existing pattern.

### Verifying `syntax.IsTimeType` respects package boundaries

- `internal/builder/gen_schema.go:24` and `internal/builder/builder.go:8` and
  `internal/builder/model.go:9` already import
  `github.com/tylergannon/go-gen-jsonschema/internal/syntax`; `internal/syntax`
  imports nothing from `internal/builder` (confirmed by inspection of
  `internal/syntax/*.go` imports). Adding `syntax.IsTimeType(path, name string) bool`
  and calling it from `internal/builder/gen_schema.go`'s `renderSchema`
  (`gen_schema.go:670`, replacing the literal `node.Path == "time" &&
  node.Name == "Time"`) is strictly downstream of the existing dependency
  direction (`builder` → `syntax`); it cannot create a cycle and does not
  require `syntax` to know anything about rendering.
- Confirmed the two call sites the plan says must share the predicate are the
  only two remote-registration sites in the scanner: the `*dst.Ident` branch
  (`scan_result.go:544-546`) and the `*dst.SelectorExpr` fallback
  (`scan_result.go:547-558`). `addTypeByID` (`scan_result.go:402`, driven by
  `NewInterfaceImpl` marker-call parsing) is a third call site but can never
  receive `time.Time`, since it only ever receives interface-impl type
  arguments parsed from source, not field types — the plan is right not to
  touch it.
- Confirmed by the existing, working `time.Time` special case
  (`gen_schema.go:670-682`, exercised today by `examples/structs/types.go:51`
  and shipped in commit `f1da0b2`) that `dst`'s decorator already resolves a
  package-qualified stdlib field type like `time.Time` to a `*dst.Ident` with
  `Path == "time"`, `Name == "Time"` — the same shape `resolveTypeExpr`'s
  `*dst.Ident` branch already switches on for every other cross-package type
  (`scan_result.go:517-546`). So a single `syntax.IsTimeType(path, name)`
  helper genuinely can serve both the scan-time `*dst.Ident` branch and the
  render-time switch with one canonical identity check, with no adapter or
  boundary-crossing type needed — the "respects package boundaries" concern
  is satisfied by construction, not merely by convention as round 2 flagged.

### Other re-verification (whole-plan pass, not limited to the two callouts)

- Re-confirmed both underlying defects directly against current source:
  `skipField`'s inverted lowercase-search at `scan_result.go:583-591`, and
  the local-ident interface-condition inversion at `scan_result.go:523-529`
  (`!ok` on `r.Interfaces[expr.Name]` returns success when *not* registered,
  errors when registered) — unchanged since rounds 1-2, still exactly as the
  plan describes.
- Re-confirmed `internal/builder/testfixtures/interfaces/types.go:29`
  (`FancyStruct.IFace TestInterface`) is exported, and its current golden
  (`FancyStruct.json`) already renders the `iface` union fully — because the
  *render*-side field walk uses the already-correct `StructField.Skip()`
  (`node_wrappers.go:688-714`) and the interface impls are discovered via
  `NewInterfaceImpl` marker-call parsing (`scan_result.go:388-404`)
  independent of struct-field traversal, not because scan-time traversal of
  the field currently works. This confirms the plan's core coupling claim
  from a different angle than either prior round used: today `IFace` renders
  correctly *despite* being invisible to `resolveTypeExpr`, so fixing
  `skipField` alone is what newly exposes the still-inverted interface
  branch on this exact fixture, which is why "one implementation unit" is
  correct.
- Re-confirmed `.github/workflows/go.yml:25-33` is unchanged since round 2's
  verification and still matches the plan's proof-commands block verbatim
  (`go test ./...` → `go generate ./...` → clean-tree check →
  `JSONSCHEMA_NO_CHANGES=1 go generate ./...` → `go test ./...`).
- Re-confirmed `TestBasic`'s case table (`internal/builder/basic_test.go:78-177`)
  currently ends at `test9-v1-interfaces-options`, so `test11-traversal` does
  not collide with any live case name, and the table's `t.Run(tc.testName,
  ...)` (`basic_test.go:181`) pattern makes the plan's `-run
  'TestBasic/(test5-interfaces|test11-traversal)'` proof command syntactically
  valid once the case exists.
- Checked the Non-goals and the remote-interface boundary (round 1 finding
  #3) once more against `internal/syntax/testfixtures/typescanner`: still
  correctly scoped out — a field typed as a *remote*, consumer-registered
  interface (`scannersubpkg.MarkerInterface`, registered only in
  `typescanner/calls.go`) is not touched by decision #4's requested-remote-
  name classifier, because that classifier only recognizes an interface that
  is registered *in the remote package's own* `Interfaces` map
  (`scan_result.go:396`, populated while scanning that package's own type
  declarations), not one registered by a third package. The plan's Non-goals
  section still states this accurately and the mechanism genuinely can't
  silently start working for that shape.

## Findings

No **critical**, **bug**, or **design** findings remain. One pre-existing,
plan-unrelated repository observation and no new nits arising from this
round's plan text.

### Dissent carried forward from rounds 1-2: none

Both prior rounds' design findings were concretely resolved, not merely
addressed in prose:

- Round 1's design finding #1 (proof commands must include the repo's own
  generation gate) — present verbatim in the current proof-commands block
  and still matches CI exactly (re-verified above).
- Round 1's design finding #2 / round 2's sole design finding (fixture
  topology for the already-loaded remote-enum panic needs a named, concrete
  package layout that actually proves the mechanism) — resolved by the
  current step 7 `test11-traversal` layout, and this round's independent
  trace confirms the topology forces the bug under *both* map-iteration
  orders, not just one. This was the single largest open risk carried into
  round 3, and it holds up under adversarial re-derivation.
- Round 1's design finding #3 (remote registered interfaces as field types) —
  still correctly scoped to Non-goals; re-verified the registry-keying reason
  it can't silently misbehave.
- Round 2's nit (`time.Time` canonical identity shared by convention only) —
  resolved by the `syntax.IsTimeType(path, name)` helper design, and this
  round confirms the two call sites the plan targets are exhaustive and the
  dependency direction is sound.
- Round 1's remaining nits (`SelectorExpr` fallback coverage, `token.IsExported`,
  routing `PropNames` through the shared helper, back-filling real test
  names) are all still present in the plan text and unchanged; none are
  disputed by this review.

### nit: an unrelated, pre-existing numbering gap in the fixture table (not a plan defect)

`internal/builder/basic_test.go`'s `TestBasic` case table currently runs
`test1`…`test9` with no `test10` entry, yet
`internal/builder/test_run/test10-v1-enums-stringmode/` is tracked in git
(`git ls-files` confirms `go.mod`, `types.go`, `schema.go`, `jsonschema_gen.go`
under that path) alongside an untested, unregistered
`internal/builder/testfixtures/v1_enums_stringmode` fixture. This predates
the issue 29 plan entirely and has nothing to do with any change the plan
proposes — flagging only so the implementer isn't confused if they notice
`test11-traversal` skips a number that looks reserved. No action requested;
not a finding against the plan.

## Verdict

Implementation is ready to begin. Across three independent review rounds, no
critical or bug finding was ever raised against this plan, and every design
finding raised in rounds 1-2 has been resolved with concrete, independently-
re-verifiable mechanism — most notably the `test11-traversal` fixture
topology, which this round confirmed by hand-tracing the shared `deps` map
mutation to panic under both possible `range` orders over `r.remoteTypes`,
and the `syntax.IsTimeType` sharing, which this round confirmed crosses the
`builder`→`syntax` dependency edge in the already-existing direction with no
adapter needed. No further review round is requested.
