# Review: sealed-interface slices implementation (sonnet, final independent round)

date: 2026-07-10
reviewer: Claude (sonnet), final independent review round
scope: full uncommitted diff vs `origin/main` (7c5f4a6, confirmed via `git merge-base
origin/main HEAD`) in
`/Users/tyler/src/.worktrees/go-gen-jsonschema/arrays-sealed-union-investigation`,
including untracked files (`examples/sealed_interface_slices/`,
`internal/builder/unsupported_interface_containers_test.go`, both fixture test
files), judged against the contract in
`ephemeral/worklog/202607101748-arrays-sealed-union-investigation.md`. This
round runs after the round-1 (fable) and round-2 (sonnet) consensus fixes
recorded in the worklog's "Consensus and release continuation" section were
applied to the working tree.

## Exact prompt received

```
Perform a final independent review of the complete current uncommitted
implementation in
`/Users/tyler/src/.worktrees/go-gen-jsonschema/arrays-sealed-union-investigation`
against the desired end state in
`ephemeral/worklog/202607101748-arrays-sealed-union-investigation.md`.

The supported boundary is exactly direct one-dimensional `[]I` struct fields
whose element is a registered sealed-interface union, through both V1 field
registration and legacy `NewInterfaceImpl`. Schema output, generated decoding,
missing/null/empty semantics, transactionality, indexed errors, illegal shapes,
and both builder fixture paths must agree.

Review the current revised diff broadly for:

1. Completeness and correctness.
2. Whether behavior is correctly and sufficiently proven.
3. Whether the change is non-over-engineered and avoids unrequested
   features, security theater, robustness belt-and-suspenders, and test bloat.
4. Idiomatic Go.

Inspect the full diff and runnable proof yourself. Prefer demonstrable findings
over hypotheticals and do not propose gold plating. Do not edit, stage, format,
generate, or otherwise mutate any repository file except the review artifact
named below. Read-only commands and tests are allowed; do not run commands that
rewrite tracked or generated files.

Write the exact prompt you received followed by labeled findings to:

`/Users/tyler/src/.worktrees/go-gen-jsonschema/arrays-sealed-union-investigation/ephemeral/reviews/2026071003-sealed-interface-slices-sonnet-final.md`

Use file/line references when possible. Labels:

- critical: must fix before proceeding.
- bug: demonstrable incorrect behavior, broken contract, race, or regression.
- design: architecture, boundary, scope, maintainability, or proof issue that
  is materially likely to cause problems.
- nit: small cleanup that should not block progress.

Explicitly state when a category has no findings. End by stating whether only
nits remain and whether the implementation is ready for the authorized skill,
`llms.txt`, README/example-index, and website updates.
```

## Method

Read the worklog end to end, including both prior review artifacts
(`ephemeral/reviews/2026071001-sealed-interface-slices-fable.md` and
`ephemeral/reviews/2026071002-sealed-interface-slices-sonnet.md`) and the
worklog's consensus log recording which findings from each were accepted,
rejected, or deferred, and why. Rather than trusting that log, I independently
re-traced the classifier and decoder by hand against the diff and re-verified
that each accepted fix is actually present and doing what the log claims:

- read the full diff of `internal/builder/gen_schema.go` (483 lines) and
  `internal/builder/schemas.go.tmpl` line by line;
- read `internal/builder/unsupported_interface_containers_test.go` in full and
  ran it with `-v` to confirm all eleven subtests (five V1, six legacy) exist
  and pass, matching the worklog's illegal-shape table;
- read `examples/sealed_interface_slices/{types,schema,schema_test}.go` and
  `internal/builder/testfixtures/interfaces/interfaces_test.go` in full and
  checked each assertion against the worklog's schema/runtime contract
  clauses;
- diffed `internal/builder/testfixtures/v1_interfaces_options/{types,schema,optionality_test}.go`
  against `origin/main` to confirm the round-2 V1-fixture-parity fix (design
  finding 1) landed, and confirmed the resulting golden
  (`jsonschema/Owner.json.golden`) renders `ifs` as `{"type":"array","items":
  {"anyOf":[...]}}`;
- confirmed the round-1 shadowing fix
  (`TestShadowedEmbeddedInterfaceIsNotCustomDecoded`,
  `internal/builder/unsupported_interface_containers_test.go:222-255`) and the
  `//go:generate` directive
  (`examples/sealed_interface_slices/types.go:5`) are both present;
- grepped for every `UnionTypeNode{` construction site
  (`internal/builder/gen_schema.go:461,1431`) to re-verify the still-deferred
  "renderer guard keys on schema node type" observation (fable finding 4 /
  sonnet design finding 2) remains accurate and non-blocking, rather than
  re-asserting it from the prior review's word;
- confirmed helper-function dedup by grepping the legacy fixture golden for
  `^func __jsonUnmarshal` (exactly one match for `TestInterface`, used by both
  the scalar `iface` and slice `ifaces` fields).

Independently re-ran, from the repository root:

```sh
go build ./...
go run ./gen-jsonschema gen --target ./examples/sealed_interface_slices --pretty --no-changes
go test ./...
go test ./internal/builder -run TestUnsupportedRegisteredInterfaceContainersFailDuringGeneration -count=1 -v
go vet ./...
gofmt -l <every changed .go file>
git diff --check
```

All passed; `--no-changes` reported no drift; `gofmt -l` and `git diff --check`
produced no output. No file was edited except this review artifact.

## Findings

### critical

None found.

### bug

None found. I specifically re-verified, rather than assumed, the two behaviors
the prior rounds fixed and flagged as easy to regress:

- **Outer-field shadowing** (`internal/builder/gen_schema.go:1524-1531`):
  `resolveLocalInterfaceProps` claims `seenProps` for every non-embedded
  field's JSON names in one pass before classification, then recurses into
  embedded structs afterward, so an outer field still shadows a same-named
  embedded interface field. `TestShadowedEmbeddedInterfaceIsNotCustomDecoded`
  asserts `builder.customTypes["Owner"]` is empty for exactly this shape, and
  I ran it directly — it passes.
- **Missing-vs-null slice semantics**
  (`internal/builder/schemas.go.tmpl:139-156`): a missing `events` key leaves
  `wrapper.Events` as a zero-length `json.RawMessage`, taking the
  `len(...) == 0` branch that preserves the pre-populated destination; an
  explicit `"events":null` is 4 non-zero bytes, so it unmarshals into a nil
  `[]json.RawMessage`, and `__raw != nil` gates the `make` call, so
  `__decoded` stays nil and the field is cleared. I traced this by hand and
  confirmed it against `TestBatchUnmarshalInterfaceSlice/missing_preserves_and_null_clears`,
  which passes.

I also independently traced (not merely re-ran) the remaining runtime
contract clauses and found them all correctly implemented: empty-array
decodes to a non-nil empty slice (`schema_test.go:46-54`, confirmed against
the template's `make([]T, len(__raw))` on a non-nil-but-empty `__raw`);
value/pointer implementations decode and re-marshal in order
(`schema_test.go:56-73`); indexed errors identify the correct
`field[index]` for both missing and unknown discriminators, in both the V1
(`events[1]`) and legacy (`ifaces[1]`) paths, and destinations are left
unmodified on failure because `__next` is a local value only assigned to
`*receiver` on the final unconditional `return nil` line.

### design

None found beyond what the worklog's consensus log already recorded as
knowingly deferred (see below) — I re-verified each deferred item rather than
re-stating it on faith, and none has become a live problem in this revision:

- **V1/legacy fixture parity** (round-2 design finding 1) is now closed:
  `internal/builder/testfixtures/v1_interfaces_options/types.go` has
  `IFaces []IFace \`json:"ifs"\`` with matching V1 registration in
  `schema.go`, a golden schema asserting `ifs.items.anyOf`, and
  `TestInterfaceSliceDecode` in `optionality_test.go`. `go test
  ./internal/builder` now pins V1-slice behavior independently of the
  standalone example, closing the gap both prior rounds flagged.
- **Five wordings of one rejection sentence** (fable finding 4 / sonnet
  design finding 2) is still true — `unsupportedRegisteredInterfaceContainer`
  is wrapped in five distinct `fmt.Errorf` shapes across
  `gen_schema.go:733,1349,1354,1362,1384,1399,1408` — but every wording
  still contains the one contracted substring the test suite and the
  worklog require, so this remains a maintenance-surface observation, not a
  behavior defect. Consciously deferred in the worklog's round-1 consensus
  log; I concur it is not blocking.
- **Renderer guard keyed on schema-node type** (fable finding 4): the
  `_, isUnion := schema.Items.(UnionTypeNode)` check at
  `gen_schema.go:732` remains correct today only because `UnionTypeNode{`
  is still constructed at exactly two sites (`gen_schema.go:461` and
  `:1431`), both exclusively for registered-interface unions — I re-grepped
  this rather than trusting the prior count. Still forward-looking, not a
  present bug.

### nit

Re-verified fable's deferred nit findings 6-8 (also re-confirmed by sonnet's
round 2) are unchanged in this revision — `registeredInterfaceField.V1` is
still derivable from `FuncNameAlias != ""`, the anonymous `v1Cfg` struct still
duplicates `s.IfaceV1`'s value type, `InterfaceProp.JSONName()`'s
`len(names)==0` branch is still unreachable, and `renderRegisteredInterfaceUnion`'s
V1 branch is still a near-copy of `mapInterface`'s assembly loop. These were
explicitly and knowingly deferred by the user in the round-1 consensus log as
cleanup-scope, not defects; I have nothing to add to that list.

## Category sweeps with no findings

- **Completeness against the contracted shape**: both registration styles
  (V1 via `examples/sealed_interface_slices` and
  `internal/builder/testfixtures/v1_interfaces_options`, legacy via
  `internal/builder/testfixtures/interfaces`) now produce
  `{"type":"array","items":{"anyOf":[...]}}` with the correct discriminator
  style, verified directly against three separate golden/schema files, not
  just one.
- **Illegal/deferred shapes**: all eleven rows of
  `TestUnsupportedRegisteredInterfaceContainersFailDuringGeneration` (fixed
  array, nested slice, nullable slice, optional slice, named slice field —
  each through both V1 and legacy registration — plus the legacy top-level
  named slice) pass and each error contains the source position and the
  contracted sentence.
- **Efficiency**: none. Classification is O(1) map lookups plus a bounded
  walk of one field's type expression, at generation time only.
- **Adventitious features / scope drift**: none. The `--target`-directory
  write fix (`gen_schema.go:958`) and the `b`→`data` receiver-parameter
  rename (`schemas.go.tmpl:128`) are both forced by the acceptance commands
  and recorded as such in the worklog; no new public API.
- **Security theater / belt-and-suspenders**: none. Nil-vs-empty preservation
  and transactional assignment are explicit contract requirements.
- **Test bloat**: none. Every test function maps to a specific contract
  clause; I did not find a redundant assertion or a table row whose removal
  would not lose coverage.
- **Idiomatic Go**: clean. `go vet ./...`, `gofmt -l` on every changed file,
  and `git diff --check` all produced no output.
- **Documentation boundary**: `git diff --name-only origin/main -- docs` is
  empty; `README.md`, `examples/README.md`, and
  `examples/uniontypes/types.go` are unchanged in this diff, consistent with
  the worklog's statement that public documentation is deferred pending
  separate authorization.

## Proof quality against the contract

Re-ran every command the worklog lists as required, plus `go vet`, `gofmt
-l`, and `git diff --check`, all independently and all passing. The negative
suite, the V1 acceptance example, the V1 builder fixture, and the legacy
builder fixture together now exercise every schema-arrival, runtime-arrival,
and negative-generator clause in the worklog's acceptance checklist at least
once, and the two registration styles are proven in fully independent test
files rather than one style riding on the other's coverage.

## Assessment

I found no critical or bug-level issues, and no design issue beyond what the
worklog's own consensus log already records as knowingly and reasonably
deferred (the error-message-wording consolidation and the schema-node-type
renderer guard, both forward-looking, both non-blocking today). The nits are
unchanged pre-existing cleanup items the user already declined to expand
scope for. Only nits remain, and they were already knowingly deferred rather
than newly discovered in this round.

The implementation is ready for the authorized next phase — the source-backed
skill, `llms.txt`, README/example-index, and website updates — with no
additional proof required beyond what has already been produced.
