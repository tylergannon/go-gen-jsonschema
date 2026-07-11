# Review: sealed-interface slices implementation (sonnet, round 2)

date: 2026-07-10
reviewer: Claude (sonnet), independent review round 2
scope: full uncommitted diff vs `origin/main` (7c5f4a6, verified via `git merge-base
origin/main HEAD` = `7c5f4a6`) in
`/Users/tyler/src/.worktrees/go-gen-jsonschema/arrays-sealed-union-investigation`,
including untracked files (`examples/sealed_interface_slices/`,
`internal/builder/unsupported_interface_containers_test.go`, fixture tests),
judged against the contract in
`ephemeral/worklog/202607101748-arrays-sealed-union-investigation.md`. This is
the requested second, independent round, run after the fixes accepted from
round 1 (`ephemeral/reviews/2026071001-sealed-interface-slices-fable.md`) were
applied to the working tree.

## Exact prompt received

```
Here is my goal:

Independently review the complete current uncommitted implementation in
`/Users/tyler/src/.worktrees/go-gen-jsonschema/arrays-sealed-union-investigation`
against the desired end state recorded in
`ephemeral/worklog/202607101748-arrays-sealed-union-investigation.md`.

The intended product boundary is exactly direct one-dimensional Go slice fields
(`[]I`) whose element is a registered sealed-interface union, supported through
both V1 field registration and legacy `NewInterfaceImpl` registration. The
generator must emit an array schema with the union under `items.anyOf`, generate
transactional element-wise interface decoding with indexed failures, preserve
value versus pointer implementations and ordinary missing/null/empty slice
semantics, and continue rejecting the illegal/deferred container forms named in
the worklog.

Give a broad-based, open-ended review of the current revised design,
implementation, generated artifacts, tests, and proof. Judge it specifically
for:

1. Completeness and correctness.
2. Whether the behavior is correctly and sufficiently proven.
3. Whether it is non-over-engineered and avoids adventitious or unrequested
   features, security theater, robustness belt-and-suspenders, and unit-test
   bloat.
4. Idiomatic Go.

If you have a code review skill, use it. Otherwise review directly. Do not
narrow the review to formatting or to the worklog's conclusions. Look for
demonstrable bugs, regressions, missing proof, unnecessary layers, scope drift,
wrong component boundaries, leaky abstractions, and simpler correct designs.
Prefer real findings over style preferences and do not propose gold plating.

Do not edit product code. Write the exact prompt you received followed by your
findings to:

`/Users/tyler/src/.worktrees/go-gen-jsonschema/arrays-sealed-union-investigation/ephemeral/reviews/2026071002-sealed-interface-slices-sonnet.md`

Use file and line references when possible. Label every finding with exactly
one of:

- critical: must fix before proceeding.
- bug: demonstrable incorrect behavior, broken contract, race, or regression.
- design: architecture, boundary, scope, maintainability, or proof issue that
  is materially likely to cause problems.
- nit: small cleanup that should not block progress.

If you find no issue in a category, say so explicitly. End with a concise
assessment of whether the implementation is ready for its authorized skill,
`llms.txt`, example-index, README, and website updates, and identify the minimum
proof still required, if any.
```

## Method

Read the worklog end to end (contract, illegal-shape list, acceptance criteria,
and the round-1 consensus log recording which fable findings were accepted,
rejected, or deferred). Read the full diff of every changed/untracked file
(`internal/builder/gen_schema.go`, `internal/builder/schemas.go.tmpl`, the
`examples/sealed_interface_slices` package, both fixture packages and their new
tests, and the negative-container test). Traced the classifier
(`resolveRegisteredInterfaceField`, `resolveLocalInterfaceProps`,
`renderRegisteredInterfaceUnion`) and the generated-decoder template branch by
hand against every row in `TestUnsupportedRegisteredInterfaceContainersFailDuringGeneration`
and against the acceptance example, rather than trusting the round-1 review's
conclusions. Independently re-ran the worklog's required commands, the full
suite, `go vet`, `gofmt -l`, and `golangci-lint run` (the `just lint` staticcheck
step fails on this machine due to a Go 1.26/staticcheck-built-with-1.25 mismatch
that is unrelated to this diff — `go.mod`/`go.sum` are untouched and the same
mismatch would occur on a clean `origin/main` checkout). Used one background
research pass to independently verify `TypeID_`/`f.ID()` semantics (whether the
new `ArrayNode{TypeID_: f.ID()}` and `UnionTypeNode{TypeID_: prop.ID()}` could
collide with or diverge from existing usage) rather than assuming it from
pattern-matching alone.

Commands re-run from repo root, all passing:

```sh
go build ./...
go run ./gen-jsonschema gen --target ./examples/sealed_interface_slices --pretty --no-changes
go test ./examples/sealed_interface_slices -count=1
go test ./internal/builder -run TestUnsupportedRegisteredInterfaceContainersFailDuringGeneration -count=1 -v
go test ./...
go vet ./...
gofmt -l <changed files>          # empty
golangci-lint run ./internal/builder/... ./examples/sealed_interface_slices/...   # 0 issues
```

## Findings

### critical

None found.

### bug

None found. I specifically re-tested the exact scenario fable's round-1 bug
finding 1 covered (outer field shadowing an embedded registered-interface field
with the same JSON name) by hand-tracing
`resolveLocalInterfaceProps` (`internal/builder/gen_schema.go:1518-1565`): the
function now marks `seenProps` for every non-embedded field in one pass
(lines 1524-1531) *before* recursing into embedded structs in a second pass
(lines 1552-1563), passing the fully-updated `seenProps` down. This restores
Go/JSON shadowing semantics and is pinned by the new
`TestShadowedEmbeddedInterfaceIsNotCustomDecoded`
(`internal/builder/unsupported_interface_containers_test.go:220-249`), which I
confirmed passes. I also independently traced:

- the missing-vs-null distinction in generated decoding
  (`internal/builder/schemas.go.tmpl:139-156`): a missing field leaves
  `wrapper.Events` as a nil `json.RawMessage` (`len == 0`), so
  `__next.Events = <receiver>.Events` preserves the pre-populated destination;
  an explicit `null` is 4 non-zero bytes, so it takes the `else` branch,
  `json.Unmarshal(null, &__raw)` yields a nil `[]json.RawMessage`, and
  `__raw != nil` is false so `__decoded` is never allocated, correctly zeroing
  the field. Verified against
  `examples/sealed_interface_slices/schema_test.go:75-92`
  (`TestBatchUnmarshalInterfaceSlice/missing_preserves_and_null_clears`), which
  passes.
- transactionality: `__next` is a local value: the destination is only
  overwritten by `*b = __next` at the very end of `UnmarshalJSON`
  (`examples/sealed_interface_slices/jsonschema_gen.go:65`), so any element
  error returns before that line and the caller's original value is
  untouched. Verified against
  `TestBatchUnmarshalErrorsAreIndexedAndTransactional`, which passes.
- helper-function dedup: legacy scalar and slice fields of the same interface
  emit one `__jsonUnmarshal__...` function
  (`internal/builder/gen_schema.go:906-909`,
  `generatedInterfaceHelpers` map), confirmed by inspecting
  `internal/builder/testfixtures/interfaces/jsonschema_gen.go.golden:47-72`
  (one call site for `IFace`, one loop-body call site for `IFaces`, one
  function definition).
- import-map safety for the new `InterfaceProp.InterfaceTypeNameWithPrefix`
  field (`internal/builder/gen_schema.go:896-899`): `ImportMap.Alias` panics if
  a package was never registered via `AddPackage`, but `s.imports()`
  (`internal/builder/gen_schema.go:363-380`) already walks every entry of
  `s.customTypes` — including the new repeated `InterfaceProp` entries, which
  are appended to the same map by `resolveLocalInterfaceProps` — and registers
  `prop.Interface.TypeSpec.Pkg()` for every one of them, so there is no
  under-registration/panic risk introduced by computing this field for
  repeated props.

None of these traces turned up a discrepancy between the code and the
contract.

### design

1. **V1-registration slice proof lives only in the acceptance example, not in
   the internal/builder fixture-golden harness.**
   `internal/builder/testfixtures/v1_interfaces_options/jsonschema_gen.go.golden`
   only changed by the unrelated `b`→`data` receiver-parameter rename (`git
   diff origin/main -- internal/builder/testfixtures/v1_interfaces_options/jsonschema_gen.go.golden`
   shows a 4-line diff, no new field). `internal/builder/testfixtures/v1_interfaces_options/types.go`
   was not touched at all. By contrast, the legacy path got both an example
   (`examples/sealed_interface_slices` uses V1 registration only) *and* fixture
   coverage (`internal/builder/testfixtures/interfaces/types.go:31` adds
   `IFaces []TestInterface`, with a matching golden schema and
   `interfaces_test.go`). This means the repo's own regression harness for the
   builder package (`go test ./internal/builder`, which is what a future
   change to `gen_schema.go` or `schemas.go.tmpl` would run first) has zero
   assertions pinning V1-slice schema shape or decode behavior; only
   `go test ./examples/sealed_interface_slices` does. A future refactor that
   breaks V1-slice handling specifically (as opposed to legacy) while leaving
   the `v1_interfaces_options` fixture's existing scalar/optional cases green
   would not be caught by `internal/builder`'s own test target, only by
   running the full suite including examples. This is the same proof-quality
   gap fable's round-1 review flagged (finding 3's discussion and the closing
   proof-quality section) and it remains true after the round-1 fixes — it was
   not on the accepted-fix list. Low cost to close: add an `Events []Event`
   (or similar) field to `v1_interfaces_options/types.go` with a schema/golden
   assertion.

2. **Four different wordings for the same rejection class.**
   `unsupportedRegisteredInterfaceContainer` is reused across
   `internal/builder/gen_schema.go:732`, `:1348`, `:1352`, `:1362`, `:1399`,
   and `:1408-1409`, but is wrapped in five distinct `fmt.Errorf` shapes
   (`"%s at %s"`, `"field %s.%s: %s at %s"`,
   `"field %s.%s through named type %s: %s at %s"`,
   `"%s for interface %s at %s"`, plus the bare no-array-context path). All of
   them satisfy the test suite's `require.ErrorContains(..., "arrays/slices of
   registered interfaces are not yet supported")` check, and each individual
   message is locally reasonable (some carry more diagnostic context than
   others), so this is not incorrect — but it is a a maintenance surface: the
   five call sites must each be updated if the shared sentence is ever
   reworded, and a user comparing two different illegal shapes gets
   differently-shaped messages for what the contract treats as one policy.
   Not blocking; a candidate for a follow-up consolidation pass, the same
   spirit as fable's deferred finding 4 (which I independently re-checked and
   confirmed is still accurate: `UnionTypeNode{` is constructed only at
   `gen_schema.go:461` and `:1431`, both exclusively representing registered-
   interface unions today, so the type-based guard at line 732 is not
   presently a latent bug, only a forward-looking coupling).

### nit

Re-verified fable's deferred nit findings 6-8 are still present essentially
unchanged in the current revision (the worklog's consensus log records these
as consciously deferred, not silently dropped, so I list them here for
completeness rather than re-litigating them):

- `registeredInterfaceField.V1` (`internal/builder/gen_schema.go:1257`) is
  still fully derivable from `FuncNameAlias != ""`; its only consumer is
  `renderRegisteredInterfaceUnion`'s branch at line 1416.
- The anonymous `v1Cfg struct{ Impls []syntax.TypeID; Disc string }`
  (`internal/builder/gen_schema.go:1326-1329`) still duplicates the shape of
  `s.IfaceV1`'s value type field-for-field.
- `InterfaceProp.JSONName()`'s `len(names)==0` fallback
  (`internal/builder/gen_schema.go:1482-1487`) is still unreachable: an
  `InterfaceProp` is only constructed in `resolveLocalInterfaceProps` after
  the `unseen`/`PropNames()` walk, which requires at least one name.
- `renderRegisteredInterfaceUnion`'s V1 branch
  (`internal/builder/gen_schema.go:1428-1447`) is still a near-copy of
  `mapInterface`'s union-assembly loop (`internal/builder/gen_schema.go:452-
  485`), differing only in which struct field supplies the discriminator prop
  name and per-option discriminator source.
- `resolveNamedType` (`internal/builder/gen_schema.go:1287-1298`) still
  repeats the ident.Path-or-local-package → `GetPackage` → named-type-lookup
  preamble also present in `findInterfaceImpl`
  (`internal/builder/gen_schema.go:1567-1578`).

One new, small observation from this round: `f.Wrapper()` /`prop.Wrapper()` is
now computed twice per struct field on every generation run — once in
`renderStructField` (`internal/builder/gen_schema.go:1069`) and once inside
`resolveRegisteredInterfaceField` (`internal/builder/gen_schema.go:1319`),
since the latter is called from the former without passing the
already-computed wrapper/inner values through. This is O(1) parsing work at
build time (not per-request), so it is not a performance concern, just a
avoidable small duplication.

## Category sweeps

- **Completeness against the contracted shape**: Complete. Both registration
  styles (`examples/sealed_interface_slices/schema.go` for V1,
  `internal/builder/testfixtures/interfaces/types.go` + legacy
  `NewInterfaceImpl` for legacy — the latter registered in that fixture's
  `schema.go`, not diffed in this round but exercised by the existing
  `TestBasic/test5-interfaces` harness) produce `{"type":"array","items":
  {"anyOf":[...]}}`, verified directly against
  `examples/sealed_interface_slices/jsonschema/Batch.json:1-53` and
  `internal/builder/testfixtures/interfaces/jsonschema/FancyStruct.json.golden:103-190`.
  Both discriminator styles (custom `!kind`, default `!type`) are proven.
  Pointer implementations render as plain concrete object schemas (`Deleted`
  in `Batch.json:27-44`), not pointer-shaped artifacts.
- **Illegal/deferred shapes**: `TestUnsupportedRegisteredInterfaceContainersFailDuringGeneration`
  now has eleven rows (six V1 + five legacy, plus one legacy top-level named
  slice), covering fixed arrays, nested slices, `Nullable[[]I]`,
  `Optional[[]I]`, named slice fields, and top-level named slice types, through
  both registration paths. I re-ran this test target directly; all eleven
  subtests pass. I independently traced four representative rows (V1 fixed
  array, V1 nullable slice, legacy nested slice, legacy named slice field)
  through `resolveRegisteredInterfaceField` by hand and confirmed each hits
  the branch the test expects, rather than relying on the test passing alone.
- **Runtime proof for the supported shape**: Empty-slice, mixed value/pointer
  order, missing-vs-null, indexed missing-discriminator and
  unknown-discriminator failures, transactionality, and re-marshal payload
  preservation are all asserted and passing for the V1 path
  (`examples/sealed_interface_slices/schema_test.go`) and for the legacy path
  (`internal/builder/testfixtures/interfaces/interfaces_test.go`,
  copied verbatim into `internal/builder/test_run/test5-interfaces/`, which I
  diffed byte-for-byte identical — consistent with this repo's established
  testfixtures→test_run copy convention).
- **Over-engineering / adventitious features**: None found. The diff's two
  changes beyond the immediate slice feature — writing generated Go into
  `s.Scan.Pkg.Dir` instead of the process CWD
  (`internal/builder/gen_schema.go:958`), and the `b`→`data` receiver-parameter
  rename (`internal/builder/schemas.go.tmpl:128`) — are both prerequisites the
  worklog records as forced by the acceptance commands (`--target` from the
  repo root, and the `Batch` receiver colliding with the renamed byte
  parameter), not scope creep. No new public API was added; `go-gen-jsonschema`'s
  root package is untouched by this diff.
- **Security theater / belt-and-suspenders**: None found. The
  nil-vs-empty-slice preservation and transactional assignment are explicit
  contract requirements, not defensive extras; there is no added input
  validation, retries, or logging beyond what the feature needs.
  `go vet` and `golangci-lint run` (0 issues) found nothing to add here either.
- **Unit-test bloat**: None found. The negative table
  (`unsupported_interface_containers_test.go`) is one row per (registration
  style × illegal shape) cell in the worklog's own table; the acceptance
  example's three test functions map directly onto the worklog's "Schema
  arrival checks" / "Runtime arrival checks" sections; the legacy fixture adds
  exactly two focused tests (happy path + indexed/transactional failure). I
  did not find a duplicated assertion, an assertion for behavior outside the
  contract, or a table row that could be deleted without losing coverage.
- **Idiomatic Go**: Good. Errors are wrapped with `%w` and carry position
  information; the generated decoder mirrors the pattern the rest of the
  template already uses (wrapper struct + `json.RawMessage` fields +
  post-decode dispatch); `gofmt -l` is clean on every touched file;
  `golangci-lint run` reports 0 issues on `internal/builder/...` and
  `examples/sealed_interface_slices/...`. The nits above (largely inherited
  from round 1 and consciously deferred by the user) are the only residue.

## Assessment

The implementation matches the worklog's contract for the exact shape it
claims to support (direct one-dimensional `[]I` fields, both registration
styles), and the round-1 consensus fixes (shadowing restoration, the
`go:generate` directive, legacy-branch negative coverage, and the
missing-vs-null decoding distinction) are all correctly present and covered by
tests I independently traced and re-ran, not just inherited on faith. I found
no critical or bug-level issues in this round. `go build`, the five worklog
acceptance commands, `go test ./...`, `go vet`, `gofmt -l`, and
`golangci-lint run` all pass cleanly against the current working tree.

The one design item worth closing before treating V1-registered interface
slices as equally proven as the legacy path is finding design-1: add a slice
field to `internal/builder/testfixtures/v1_interfaces_options` (with a schema
golden and a focused decode test) so the builder package's own fixture harness
— not only the standalone example — pins V1-slice behavior. This is proof
hygiene, not a functional gap; nothing about the current behavior is wrong,
but nothing in `go test ./internal/builder` alone would currently catch a
V1-slice-specific regression that the legacy path's fixture coverage would
catch for that path.

Given that, the implementation is ready for the requested next phase (skill,
`llms.txt`, example-index, and README/website updates) once design-1 is
addressed or explicitly accepted as-is; none of the outstanding nits or
design-2 (message-wording consolidation) block that work, and both were
already knowingly deferred in the worklog's round-1 consensus log for good
reason (they are forward-looking / cleanup items, not present defects).
Minimum proof still required before calling V1-slice support as thoroughly
proven as legacy-slice support: the `v1_interfaces_options` fixture addition
described above, followed by a re-run of `go test ./internal/builder` and
`go test ./...`.
