# Review: sealed-interface slices implementation (fable, round 1)

date: 2026-07-10
reviewer: Claude (fable), independent review round 1
scope: full uncommitted diff vs `origin/main` (7c5f4a6) in
`/Users/tyler/src/.worktrees/go-gen-jsonschema/arrays-sealed-union-investigation`,
plus untracked files (`examples/sealed_interface_slices/`,
`internal/builder/unsupported_interface_containers_test.go`, fixture tests),
judged against the contract in
`ephemeral/worklog/202607101748-arrays-sealed-union-investigation.md`.

## Exact prompt received

```
Here is my goal:

Review the complete current uncommitted implementation in
`/Users/tyler/src/.worktrees/go-gen-jsonschema/arrays-sealed-union-investigation`
against the desired end state recorded in
`ephemeral/worklog/202607101748-arrays-sealed-union-investigation.md`.

The implementation is intended to support exactly direct one-dimensional Go
slice fields (`[]I`) whose element is a registered sealed-interface union, for
both V1 field registration and legacy `NewInterfaceImpl` registration. It must
emit an array schema with the union under `items.anyOf`, generate transactional
element-wise interface decoding with indexed failures, preserve value versus
pointer implementations, and continue rejecting the illegal/deferred container
forms named in the worklog.

Give a broad-based, open-ended review of the current design, implementation,
generated artifacts, tests, and proof. Judge it specifically for:

1. Completeness and correctness.
2. Whether the behavior is correctly and sufficiently proven.
3. Whether it is non-over-engineered and avoids adventitious or unrequested
   features, security theater, robustness belt-and-suspenders, and unit-test
   bloat.
4. Idiomatic Go.

If you have a code review skill, use it. Otherwise review directly. Do not
narrow the review to formatting or the author's stated theory. Inspect the full
diff against current `origin/main`, repository conventions, the acceptance
example, generated output, and relevant implementation paths. Look for
demonstrable bugs, regressions, missing proof, unnecessary layers, scope drift,
wrong component boundaries, leaky abstractions, and simpler correct designs.
Prefer real findings over style preferences and do not propose gold plating.

Do not edit product code. Write the exact prompt you received followed by your
findings to:

`/Users/tyler/src/.worktrees/go-gen-jsonschema/arrays-sealed-union-investigation/ephemeral/reviews/2026071001-sealed-interface-slices-fable.md`

Use file and line references when possible. Label every finding with exactly
one of:

- critical: must fix before proceeding.
- bug: demonstrable incorrect behavior, broken contract, race, or regression.
- design: architecture, boundary, scope, maintainability, or proof issue that
  is materially likely to cause problems.
- nit: small cleanup that should not block progress.

If you find no issue in a category, say so explicitly. End with a concise
assessment of whether the work is ready to document and ship, and identify the
minimum proof still required, if any.
```

## Method

Ran the code-review skill at high effort: 8 parallel finder passes
(line-by-line, removed-behavior, cross-file, reuse, simplification,
efficiency, altitude, repo conventions) over the diff and untracked files,
followed by verification. Where a candidate was decidable by execution I
verified it directly with temporary probe tests against the real builder
(all probe scaffolding was removed afterward; `git status` matches the
pre-review state). Independently re-ran the worklog's arrival proof:
`go test ./...` passes; `gen --target ./examples/sealed_interface_slices
--pretty --no-changes` reports no drift; the legacy fixture
(`TestBasic/test5-interfaces`) and negative container suite pass.

Verdict legend: CONFIRMED = demonstrated by executing code or by directly
quoted code; PLAUSIBLE = consistent with the code but not executed.

## Findings

### 1. bug — outer fields no longer shadow embedded registered-interface fields; generator emits non-compiling code (CONFIRMED)

`internal/builder/gen_schema.go:1530-1548` (new `unseen` handling in
`resolveLocalInterfaceProps`).

On `origin/main`, `seenProps` was advanced for **every** direct-ident field
before the interface check, so an outer field claimed its JSON names and a
same-named field on an embedded struct was skipped (Go/JSON shadowing
semantics). The rewrite advances `seenProps` only after
`resolveRegisteredInterfaceField` returns non-nil — i.e. only registered
interface fields mark names as seen. Non-interface fields now `continue`
at line 1537-1539 without claiming their names.

Demonstrated with a probe fixture run through the real builder:

```go
type Inner struct { Foo Variant `json:"foo"` }
type Owner struct {
    Inner
    Foo string `json:"foo"` // shadows Inner.Foo in Go and in encoding/json
}
```

New behavior: `customTypes["Owner"]` records the embedded `Inner.Foo`
interface prop, and the generated `Owner.UnmarshalJSON` contains

```go
Foo json.RawMessage `json:"foo"`
...
if __next.Foo, err = __jsonUnmarshal__fixture__Variant(wrapper.Foo); err != nil {
```

where `__next.Foo` resolves to the **outer** `string` field — the generated
file does not compile (`cannot use Variant value as string`). On
`origin/main` this input generated no interface prop for `Owner` at all
(correct: the outer field shadows). This is a regression introduced by the
refactor, orthogonal to the slice feature itself. It is an edge shape
(same-JSON-name shadowing across embedding), and it fails loudly at build
time rather than silently at runtime, hence bug rather than critical.

Fix shape: claim `prop.PropNames()` in `seenProps` for every non-embedded
field (as the old code did for ident fields — claiming for all fields is
even more faithful to encoding/json), before deciding whether the field is
a registered interface.

### 2. design — acceptance example is missing the `//go:generate` directive every sibling example carries (CONFIRMED)

`examples/sealed_interface_slices/types.go:3`.

All other examples with committed generated output start `types.go` with
`//go:generate go run ../../gen-jsonschema/` (e.g.
`examples/basictypes/types.go:3`, `examples/uniontypes/types.go:3`), and
`examples/README.md` documents that regeneration flows through these
directives ("the directives run the generator from this checkout"); repo
CLAUDE.md documents `cd examples/basictypes && go generate ./...`. The new
example has no directive, so `go generate ./...` is a no-op there and it is
the only example whose committed `jsonschema_gen.go`/`jsonschema/` cannot be
refreshed by the documented workflow. This is a functional workflow gap in
the acceptance artifact, not a `./docs` change, so it is not covered by the
documentation-permission deferral. One-line fix plus regeneration check.

(Related, intentionally not filed as a finding: `examples/README.md` does
not list the new example. Public-documentation updates are explicitly
deferred pending authorization per the worklog contract, so the omission is
correct for now, but it belongs on the authorized-docs checklist alongside
README limitation text.)

### 3. design — negative proof does not cover the legacy-registration branch of the classifier (CONFIRMED gap; behavior itself verified correct)

`internal/builder/unsupported_interface_containers_test.go:25-124` vs
`internal/builder/gen_schema.go:1381-1408`.

The committed negative suite exercises fixed array, nested slice, nullable
slice, optional slice, and named slice **only through V1 registration**,
plus one legacy case (top-level named slice). The classifier has two
separate rejection paths — the `v1Configured` branch (lines 1346-1378) and
the legacy/default branch (lines 1381-1408, `registeredInterfaceInExpr`
catch-all plus the named-type-underlying sweep) — and the legacy branch's
field-level rejections are locked by no test. I probed legacy-registered
`[2]I`, `[][]I`, named-slice-field, `Nullable[[]I]`, and `*I` through
`builder.New`: all are rejected today with position-bearing errors, so this
is purely a proof gap, not a behavior gap. But the worklog's own decision
("use one classifier so legacy and V1 cannot diverge") is exactly the
invariant these missing cases would pin down; the two branches are separate
enough code that they can regress independently. Adding legacy-registration
rows to the existing table test is cheap (the harness already supports it).

### 4. design — container rejection is split across three mechanisms, and the renderer-level guard keys on schema node type (PLAUSIBLE, forward-looking)

- `internal/builder/gen_schema.go:732` — generic `renderSchema` ArrayType
  case rejects when the rendered `Items` is a `UnionTypeNode`;
- `gen_schema.go:1397-1399` — classifier catch-all via
  `registeredInterfaceInExpr`;
- `gen_schema.go:1400-1408` — classifier named-type-underlying sweep.

Verified interplay: the line-732 guard is unreachable for approved direct
`[]I` fields (those are consumed in `renderStructField` and never reach the
generic array renderer), and it is what actually rejects the top-level named
slice; the 1400-1408 sweep rejects the same named-slice shape when it
appears as a field, duplicating the guard with a different (better) message.
Two costs: (a) the same policy lives in three places that must be updated
together when the container rules change — the sweep at 1400-1408 could be
deleted today at the price of a less specific error, or kept as the single
site with the renderer guard as an assertion; (b) the line-732 check tests
the rendered node type rather than the AST shape, so the day any
non-registered-interface construct renders to `UnionTypeNode` (it is
semantically a generic anyOf node), every array of it is wrongly rejected
with an interface-specific message. Not a today-bug; a boundary worth one
consolidation pass before the container rules grow (fixed arrays and named
aliases are already named as follow-ups in the worklog).

### 5. design — misleading diagnostic for non-array unsupported containers (CONFIRMED)

`internal/builder/gen_schema.go:1397-1399`.

The catch-all fires for **any** remaining shape containing a registered
interface — pointer (`*I`), map value, inline struct field — and always
emits `arrays/slices of registered interfaces are not yet supported`.
Probe-confirmed: a legacy `Values *Variant` field fails with
"arrays/slices of registered interfaces are not yet supported for interface
Variant at …". `origin/main` said "found registered interface type Variant
in an unsupported location", which was accurate for these shapes. Rejection
correctness is preserved (verified for `*I`, map, inline struct via the
catch-all and the renderer's existing map/interface rejections); only the
diagnostic regressed, and it now points users at the wrong fix. Cheap
repair: keep the old generic message for non-array shapes and reserve the
arrays/slices message for shapes where `containsArrayType` is true (the
helper already exists and is currently used only to flavor the V1 message
at line 1348).

### 6. nit — classifier bookkeeping is more stateful than it needs to be

`internal/builder/gen_schema.go:1251-1343`.

- `registeredInterfaceField.V1` is derivable state: it is true exactly when
  `FuncNameAlias != ""` (both set only in the `v1Configured` branch). Its
  single consumer is the branch in `renderRegisteredInterfaceUnion:1413`.
  Keeping the pair in lockstep is a small standing hazard; drop the flag or
  drop the alias-emptiness convention, not both.
- The anonymous `v1Cfg struct{ Impls []syntax.TypeID; Disc string }`
  (1327-1330) re-declares the value type of `s.IfaceV1` field-for-field;
  a named config type used in both places removes the shadow copy and the
  parallel `v1GoField`/`v1Configured` vars.
- `InterfaceProp.JSONName()`'s `len(names)==0` fallback
  (gen_schema.go:1479-1485) is unreachable: props are only constructed after
  the `unseen` check, which requires at least one prop name.

### 7. nit — small duplications against existing machinery

- `renderRegisteredInterfaceUnion`'s V1 branch (gen_schema.go:1428-1444) is
  a near-copy of `mapInterface`'s union-assembly loop (452-485), differing
  only in `DiscriminatorPropName`/`TypeID_` and per-option `Discriminator`
  assignment. This predates the diff in spirit (the old inline V1 loop was
  the same copy); the refactor was the natural moment to fold both into one
  union builder, and next time the assembly rules change two sites must move
  together.
- `resolveNamedType` (1287-1298) repeats the ident.Path-or-local-package →
  `GetPackage` → named-type lookup preamble that `findInterfaceImpl`
  (1567-1579) and `resolveEmbeddedType` already contain. It does replace two
  previous inline copies, so it is a net wash; a single shared resolver
  would finish the job.

### 8. nit — generated-identifier hygiene is only partially closed by the `b` → `data` rename

`internal/builder/schemas.go.tmpl:128-149`, `gen_schema.go:903`.

The rename fixes the receiver-initial collision for type names starting
with "B" (and, since every other local is now multi-character or
`__`-prefixed, the whole single-letter-receiver class). But the template
still declares bare `Alias` and `Wrapper` types and a `wrapper` local, so a
user struct (or interface field) named `Alias` or `Wrapper` still produces
non-compiling generated code. Pre-existing, out of contract scope, and the
fix taken is proportionate — noted so the remaining class is a known
follow-up (`__`-prefix the synthetic type names next time the template is
touched). Similarly microscopic: `JSONName` is interpolated directly into
the `fmt.Errorf` format string at schemas.go.tmpl:149, so a JSON property
name containing `%` would garble the error text (not a crash).

## Category sweeps with no findings

- **critical**: none found.
- **Efficiency**: none. The double classification per field
  (renderStructField + resolveLocalInterfaceProps) is O(1) map lookups plus
  a walk of the field's own type expression, at build time; the generated
  decoder adds one guarded allocation and per-element dispatch, which is
  inherent to discriminator decoding. Verified `findInterfaceImpl` is map
  lookups, not scans.
- **Adventitious features / scope drift**: none. The diff implements
  exactly the contracted shape; the two changes beyond it (the
  `--target`-directory write fix at gen_schema.go:958, making the Go output
  land in `s.Scan.Pkg.Dir` consistent with `RenderSchemas` at line 966, and
  the `b`→`data` rename) were both forced by the acceptance commands and are
  recorded in the worklog. No new public API, matching the contract.
- **Security theater / belt-and-suspenders**: none. The nil-vs-empty slice
  preservation and transactional assignment are contract requirements, not
  defensive extras.
- **Test bloat**: none. Tests map one-to-one onto contract clauses. The
  byte-identical fixture test files under `testfixtures/interfaces/` and
  `test_run/test5-interfaces/` follow the repo's established
  copy-into-test_run convention (the harness `RemoveAll`s and `CopyDir`s
  `test_run/<name>` from `testfixtures/` on every run, and
  `test9-v1-interfaces-options/optionality_test.go` is committed precedent),
  so I do not flag them.
- **Cross-file breakage from the WriteFile path change**: none found. `gen`
  mains run via `go:generate` with CWD == package dir (identical result);
  the CLI passes `--target` without chdir; the acceptance command from the
  repo root exercises the new path. The helper-dedup map
  (gen_schema.go:906-909) is correct and necessary for the scalar+slice
  same-interface case; the theoretical collision of two same-named
  interfaces in same-named packages produced a compile error before the
  dedup (duplicate function) and still produces a compile error after it
  (type mismatch), so no silent regression.
- **Idiomatic Go**: good overall — table-driven parallel tests, `%w`
  wrapping, position-bearing errors, `slices.Clone` before mutating shared
  map values, generated output kept `gofmt`/`goimports` clean. The nits in
  6-7 are the residue.

## Proof quality against the contract

Independently re-verified: schema arrival checks (array + `items.anyOf`,
both discriminator styles, no property-level anyOf, pointer impl as concrete
object schema), runtime arrival checks (empty/mixed/order, indexed
missing/unknown discriminator at `events[1]` and `ifaces[1]`,
transactionality, re-marshal payloads, null/missing → nil), negative
generator proof for the six committed shapes, `--no-changes` drift check,
and full `go test ./...`. All pass as claimed in the worklog. The V1 slice
runtime proof lives in the acceptance example and the legacy proof in the
`interfaces` fixture, which together cover both registration styles — the
split is reasonable, though the `v1_interfaces_options` fixture itself
gained no slice field, so V1-slice coverage rides entirely on the example.

## Assessment

The core feature is complete, correct for the contracted shape, tightly
scoped, and unusually well-proven: both registration styles, both
discriminator styles, value/pointer identity, indexed transactional
decoding, nil-vs-empty behavior, and drift-free regeneration are all
executable proof, and the classifier-unification design decision is sound.
The work is close to ready, but not ready to document and ship as-is:

1. Fix finding 1 (restore seen-name claiming for non-interface fields) —
   it is a real regression that makes the generator emit non-compiling code
   for a previously-handled input, and the fix is small.
2. Fix finding 2 (add the `//go:generate` directive and re-verify
   `--no-changes`) — trivial, keeps the acceptance example maintainable by
   the documented workflow.

Minimum proof still required:

- a regression test for finding 1 (outer field shadowing an embedded
  registered-interface field with the same JSON name generates no interface
  prop / still compiles);
- legacy-registration rows (fixed array, nested slice, named slice field,
  nullable slice) in `TestUnsupportedRegisteredInterfaceContainersFail
  DuringGeneration` per finding 3;
- re-run of the worklog's five required commands after the above.

Findings 4-8 are non-blocking: 4 and 5 are worth a small consolidation pass
before the follow-up container work (fixed arrays, named aliases) begins;
6-8 are cleanup that can ride along with any later touch of these files.
