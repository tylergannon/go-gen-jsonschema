# Issue 29 implementation review

## Prompt given to me

Here is my goal:

Implement GitHub issue 29 in `/Users/tyler/src/go-gen-jsonschema`: repair
exported/unexported struct-field type discovery, preserve deliberate skip and
embedded-field behavior, resolve registered local interfaces, load concrete
cross-package schemas, keep `time.Time` a renderer-owned leaf, and safely
classify already-loaded remote registered enums.

Perform a final implementation review of the current working tree against:

- `docs/design/issue-29-plan.md`
- the current product-code and test diff (`git diff`)
- untracked `internal/syntax/scan_result_test.go`
- untracked `internal/builder/testfixtures/traversal/`
- generated untracked `internal/builder/test_run/test11-traversal/`

The reviewed plan had already reached consensus. During implementation, the
full test suite exposed two compatibility paths that were dormant behind the
visibility bug: map expressions in syntax-only flattening and named interfaces
registered through v1 builder options. The implementation now treats
`*dst.MapType` and `*dst.InterfaceType` as non-recursive discovery boundaries;
maps remain renderer-rejected, and v1 interface implementations remain owned
by the builder. Review that adjustment especially carefully for hidden
acceptance of unsupported schemas, missed remote dependencies, or a simpler
correct boundary.

Also verify the complete implementation for correctness, regression risk,
scope drift, generated-fixture fidelity, proof gaps, and order-dependent
behavior. The current proof state is:

- focused syntax tests pass;
- `TestBasic/(test5-interfaces|test9-v1-interfaces-options|test11-traversal)`
  passes;
- `go generate ./...` passes;
- `JSONSCHEMA_NO_CHANGES=1 go generate ./...` passes;
- `go test ./...` passes after the compatibility adjustment.

Do not edit product code, tests, fixtures, or the plan. You may inspect the
repository and GitHub. Write the exact prompt you were given and your findings
to:

`ephemeral/reviews/202607101300-issue-29-implementation-sonnet.md`

Label every finding with exactly one of:

- **critical:** must fix before proceeding.
- **bug:** demonstrable incorrect behavior, broken contract, race, or
  regression.
- **design:** architecture, boundary, scope, maintainability, or proof issue
  that is materially likely to cause problems.
- **nit:** small cleanup that should not block progress.

Use file/line references and concrete evidence. Prefer real findings over
style preferences. If you identify a real bug outside Issue 29, place it in a
separate `Out-of-scope bugs` section so it can be classified independently.
If no critical, bug, or design findings remain for Issue 29, say so explicitly.

## Verification performed

- Read `docs/design/issue-29-plan.md` and GitHub issue #29 in full.
- Read the complete `git diff` for all seven modified files and the untracked
  `internal/syntax/scan_result_test.go`, `internal/builder/testfixtures/traversal/**`,
  and `internal/builder/test_run/test11-traversal/**`.
- Traced `resolveTypeExpr`, `resolveTypes`, `loadPackageInternal`, and the new
  `requestType` helper by hand against both possible map-iteration orders for
  the `test11-traversal` fixture (root → `remoteenum` first, and root →
  `remotestruct` → `remoteenum` first), confirming both orders now avoid the
  original zero-`TypeSpec` panic and that the fix is deterministic (not
  order-*lucky*).
- Ran `go build ./...`, `go vet ./...`, `go test ./...` (all pass).
- Ran `go test ./internal/syntax -run TestAlreadyLoadedRegisteredEnumResolves -count=200` and
  `go test ./internal/builder -run TestBasic/test11-traversal -count=15` to
  rule out order-dependent flakiness (all pass).
- Diffed `internal/builder/test_run/test11-traversal/jsonschema/TraversalHolder.json`
  against both golden files byte-for-byte (`diff` exit 0 against both).
- Independently ran `go generate ./...` followed by `git status --porcelain`
  (clean — no drift) and `JSONSCHEMA_NO_CHANGES=1 go generate ./...` (exit 0),
  reproducing the stated proof state myself rather than trusting it.
- Grepped for any other literal `"time"`/`"Time"` identity checks outside
  `syntax.IsTimeType` to confirm there is exactly one canonical definition.
- Traced how `test5-interfaces`, `test9-v1-interfaces-options`, and the new
  `typescanner` `ResolutionHolder` fixture each exercise a *different* part of
  the classification logic (registered-interface-by-name vs. the new
  `*dst.InterfaceType` structural boundary), to make sure "tests pass" wasn't
  hiding an untested branch.
- Confirmed no tracked file was modified by any verification command I ran.

No product code, tests, fixtures, or the plan were edited.

## Findings

No **critical** or **bug** findings for Issue 29.

### design

**1. The new `*dst.MapType`/`*dst.InterfaceType` discovery boundary is unconditionally permissive, including for the explicit "consumer-only-registered remote interface" non-goal, and its correctness depends entirely on an unenforced coupling to `gen_schema.go`'s render-time rejection.**

`internal/syntax/scan_result.go:494-503`:

```go
case *dst.MapType:
    // Maps are rejected by the renderer. ...
    return nil
case *dst.InterfaceType:
    // Named interfaces registered through v1 schema options remain in
    // LocalNamedTypes. ...
    return nil
```

This case fires for *any* `*dst.InterfaceType` node reached during discovery,
not only ones backed by a v1-registered local interface. Concretely: for a
type declared in a *remote* package as `type Foo interface{...}` with no
`NewInterfaceImpl`/`NewEnumType` marker call in that package's own
`schema.go` (i.e. an interface registered only by the *consuming* package —
the scenario the plan's Non-goals section names explicitly: "Adding support
for a field whose type is an interface declared in another package and
registered only by the consuming package"), the remote package's own
`ScanResult` classifies `Foo` into `LocalNamedTypes` (not `Interfaces`, since
it never sees its own registration — see the classification in
`loadPackageInternal` around scan_result.go:429-440). When that name is
queued via `requestType` (scan_result.go:647-663) and popped in
`resolveTypes` (scan_result.go:619-624), `resolveTypeExpr` is invoked directly
on the interface's own `*dst.InterfaceType` body — which is exactly the new
no-op case. Discovery therefore now silently *succeeds* for a shape the plan
says is out of scope, where before this diff it would have hit the
`default: return fmt.Errorf("unhandled expression %s", ...)` branch and
failed loudly at scan time.

I confirmed this doesn't currently produce a *wrong* schema: `gen_schema.go:732-737`
still unconditionally rejects both shapes at render time
(`mapType/chanType not allowed`, `interface types are not supported`), so a
field that reaches this boundary without a v1/legacy interface match still
errors — just later (render-time) instead of earlier (scan-time), with an
arguably clearer message. I also confirmed this is a *necessary* change, not
a gratuitous one: `internal/syntax/flatten_types_internal_test.go`'s
`TestFlattenTypes` walks `internal/syntax/testfixtures/structtype`, whose
`ExperimentRun`/`ArrayOfSuperStruct` types have real `map[string]...` fields
that are now reached by the fixed exported-field traversal, and this
pre-existing test exercises exactly this map path today.

So the boundary works, but it is wider than the plan's stated intent ("named
interfaces used by the v1 builder"), and its safety is an implicit,
un-tested cross-file invariant: `resolveTypeExpr`'s permissiveness is only
correct *because* `renderSchema` independently still rejects the same node
kinds. Nothing enforces that these two switches stay in sync — a future
change that makes rendering even slightly more permissive for one of these
node kinds (e.g. to support empty/marker interfaces, or a `map[string]any`
special case) would silently combine with this already-permissive scan
boundary to accept schemas neither file's author intended, with no failing
test to catch it. There is also no focused syntax-level unit test that pins
`resolveTypeExpr` returning `nil` (not an error) for a bare `*dst.MapType` or
`*dst.InterfaceType` node directly (as `TestIsTimeType` does for the time
boundary) — coverage is entirely indirect via `TestFlattenTypes` and
`test9-v1-interfaces-options`.

Suggested tightening (not required to land the fix as-is, since it's proven
safe today): make the comments state the actual invariant ("this returns
unclassified/undiscovered, not supported — rendering independently rejects
these node kinds; do not treat `nil` here as validation"), and/or add a
one-line focused test asserting `resolveTypeExpr` on a synthetic
`*dst.MapType`/`*dst.InterfaceType` expression returns `nil` so the boundary
is pinned the same way `IsTimeType` is.

## No other design or bug findings

Everything else I checked lines up correctly with the plan and does not
regress existing behavior:

- `skipField`, `StructField.Skip`, and `StructField.PropNames` now share
  `isExportedFieldName`/`hasExportedFieldName`/`fieldJSONIgnored` helpers
  (`internal/syntax/node_wrappers.go:664-712`, `scan_result.go:594-609`) using
  `token.IsExported`, and each caller keeps its distinct policy: `skipField`
  still skips `jsonschema:"ref=..."` fields (scan_result.go:606-609) while
  `StructField.Skip` does not, and both preserve the original embedded-field
  rule. This matches plan item 1 exactly and is a genuine improvement over the
  old `unicode.IsUpper(rune(name[0]))` check (which only inspected the first
  *byte*, not the first rune, of the identifier).
- The local-identifier classification in `resolveTypeExpr`
  (scan_result.go:519-538) now follows the exact order specified in plan item
  2: basic type → local named type (cycle-checked) → registered
  enum/constant → registered interface → position-bearing undeclared-local
  error. Verified via `TestRegisteredInterfaceIdentifierResolves` and
  `TestUnknownLocalIdentifierFails` in `scan_result_test.go`, both of which I
  traced against the new `ResolutionHolder` fixture in
  `internal/syntax/testfixtures/typescanner/local_func_defs.go:19-24`.
- `syntax.IsTimeType` (`type_id.go:5-9`) is the single canonical definition;
  I grepped the whole tree and found no other `"time"`/`"Time"` literal
  checks — both `resolveTypeExpr`'s `*dst.Ident` and `*dst.SelectorExpr`
  branches (scan_result.go:540-556) and `renderSchema`
  (`gen_schema.go:670`) call it, satisfying plan item 3's "same canonical
  type identity" requirement.
- The `requestType` helper (scan_result.go:647-663) is now the single
  classifier used both by `loadPackageInternal`'s fresh-load path
  (scan_result.go:475-479) and `resolveTypes`'s already-loaded dependency
  path (scan_result.go:624-629), per plan item 4. I hand-traced both possible
  map-iteration orders for `test11-traversal`'s root → `remotestruct` →
  `remoteenum` topology and confirmed the old bug (`remote.LocalNamedTypes[typeName]`
  indexed for a name that only exists in `remote.Constants`, producing a zero
  `TypeSpec` whose `.Concrete` is a typed-nil `*dst.TypeSpec`, causing a nil
  pointer panic in the old code's `ts.Concrete.Name.Name` check on the next
  dequeue) is fixed in both orders, and `TestAlreadyLoadedRegisteredEnumResolves`
  exercises the already-loaded branch directly and deterministically rather
  than depending on map order.
- The unreachable `*dst.SliceExpr` branch is removed (plan item 5); slice
  types are correctly represented as `*dst.ArrayType` and remain handled.
  Generic `IndexExpr`, channels, functions, and other unhandled expressions
  still fall through to the pre-existing explicit `unhandled expression`
  error — no new silent acceptance for those shapes.
- `internal/builder/testfixtures/traversal` exactly matches the plan's
  specified layout (root `TraversalHolder` with `Remote`/`Status`/`When`
  fields, `remoteenum` and `remotestruct` packages, `remotestruct` itself
  depending on `remoteenum`). The golden
  `jsonschema/TraversalHolder.json.golden` shows the remote struct rendered
  with real properties (not `{}`), the remote enum with its actual values,
  and `time.Time` still rendered as the RFC3339 string leaf — I diffed the
  generated `internal/builder/test_run/test11-traversal/jsonschema/TraversalHolder.json`
  against both golden copies and they are byte-identical.
- Ran the full stated proof sequence myself (not just trusting the prompt):
  `go build ./...`, `go vet ./...`, `go test ./...`, `go generate ./...` +
  `git status --porcelain` (clean), and `JSONSCHEMA_NO_CHANGES=1 go generate ./...`
  (exit 0) all pass on this working tree.
- Ran the order-sensitive tests under repetition
  (`TestAlreadyLoadedRegisteredEnumResolves` ×200,
  `TestBasic/test11-traversal` ×15) with no failures, since Go's map
  iteration order is randomized per-process and a single green run wouldn't
  rule out order-dependent flakiness.

## Nits

- `internal/builder/test_run/test11-traversal/` is left untracked, but
  `git ls-files internal/builder/test_run` shows the repo's existing
  convention is to commit the generated `test_run/testN-*` directories for
  every other `TestBasic` case (173 tracked files under `test_run/` today).
  Whether or not that convention is itself a good idea, this new case is
  currently inconsistent with it — worth a deliberate decision (track it for
  consistency, or stop tracking the others) rather than an accidental gap.
- The `pkgPath` computation in the `*dst.SelectorExpr` branch
  (scan_result.go:548-561) computes the same `xIdent.Path == ""` fallback
  twice — once to build `pkgPath` for the `IsTimeType` check, once again via
  `xIdent.Path != ""` for the `addType` call. Functionally correct and
  covered by `TestIsTimeType`, but could be collapsed to one `pkgPath`
  variable used in both places.

## Out-of-scope bugs

- `internal/syntax/scan_result.go:296-312` (`resolveTypeLocal`, untouched by
  this diff): on the "type not found" error path it does
  `for typeName := range r.LocalNamedTypes { fmt.Println(typeName) }` followed
  by `fmt.Println(string(debug.Stack()))` — debug scaffolding that prints an
  arbitrary-order key dump and a full stack trace to stdout on an ordinary
  error return, rather than through the returned `error` or a logger. This
  predates issue 29 (not present in the diff) and isn't newly exercised by
  it, but it's a real production code smell — including a `runtime/debug`
  stack dump on every failure of this path — worth its own cleanup ticket.

## Bottom line

No critical or bug findings for Issue 29. One **design** finding: the
`*dst.MapType`/`*dst.InterfaceType` discovery boundary is correct and proven
by tests today, but broader than the plan's stated scope and safe only
because of an implicit, untested coupling to render-time rejection in a
different file — worth tightening the comment and/or adding one pinning test,
but not blocking. Two nits, and one pre-existing out-of-scope code-quality
issue unrelated to this change.
