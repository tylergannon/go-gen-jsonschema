# Issue 30 plan review — round 1 (Fable)

Reviewer: Claude (Fable), 2026-07-10
Worktree: `/Users/tyler/src/go-gen-jsonschema-issue-30`, branch `codex/issue-30-plan`, HEAD `13421b6`
Plan under review: `docs/design/issue-30-json-tag-name-plan.md` (untracked in this worktree)

## Exact prompt given

```
Here is my goal:

Read GitHub issue 30 in this repository and produce a correct, minimal,
implementation-ready written plan for fixing omitted JSON tag names so generated
schema property names agree with Go's encoding/json behavior. The work is
planning only: do not implement the fix. Keep issue 30 independent of issues 28
and 29 except for documenting relevant constraints. Respect the repository's
AGENTS.md test requirements and the user's requirement to work only in this
separate worktree.

Review the implementation plan at:
docs/design/issue-30-json-tag-name-plan.md

Use `gh` to read issue 30 and its comments. Inspect the relevant source, tests,
fixtures, generator harness, and repository conventions yourself. Review at the
level of the actual goal, not merely formatting. Look for correctness problems,
hidden dependencies, wrong ownership boundaries, over-engineering, unrequested
features or layers, missing test cases, proof gaps, and simpler implementation
paths that still satisfy the issue.

Do not edit product code or the plan. Write the exact prompt you were given and
your findings to:
ephemeral/reviews/20260710-issue-30-plan-round-1-fable.md

Label every finding using these definitions:

- critical: must fix before proceeding.
- bug: demonstrable incorrect behavior, broken contract, race, or regression.
- design: architecture, boundary, scope, maintainability, or proof issue that is
  materially likely to cause problems.
- nit: small cleanup that should not block progress.

Give file and line references where useful. End with a concise recommended plan
shape and list any evidence you could not verify.
```

## What I verified independently

- **Bug reproduces at HEAD (`13421b6`).** I ran a temporary test (removed
  afterward; worktree returned to its prior state) constructing
  `StructField{Field: &dst.Field{...}}` directly:
  - `json:",omitzero"` → `PropNames() == [""]` (bug present)
  - `json:",omitempty"` → `[""]` (bug present)
  - `json:"max_retries,omitzero"` → `["max_retries"]` (explicit name preserved)
  - untagged → `["MaxRetries"]`
  - `json:""` → `["MaxRetries"]` — already falls back today, because
    `structTag` returns nil when `tag.Value` is empty
    (`internal/syntax/node_wrappers.go:678-681`)
  - `json:"-"` and `json:"-,omitempty"` → `Skip() == true`
- **Root cause is exactly as the issue and plan state.**
  `internal/syntax/node_wrappers.go:648-664`: `case 1` returns
  `tag.Options[0]` unconditionally. The tag parser
  (`github.com/tylergannon/structtag v0.1.0`, `tags.go:112-121`) sets
  `Options = strings.Split(value, ",")`, so `json:",omitzero"` yields
  `Options = ["", "omitzero"]` with `Value = ",omitzero"` (non-empty, so the
  nil-guard in `structTag` does not save us). `Options[0]` can never panic:
  `Value == ""` returns nil before indexing.
- **Ownership claim is correct.** Every property-name consumer in the builder
  goes through `PropNames`: `internal/builder/gen_schema.go:166` (provider
  JSONName), `:1017` (seen-prop collection), `:1171` (template holes), `:1185`
  (object prop rendering), `:1265`, `:1294`, `:1313` (interface props). The
  only other direct `Options[0]` reads are skip checks for `"-"`
  (`node_wrappers.go:702`, `scan_result.go:598`), which don't consume the
  name. Fixing `PropNames` alone covers generation, providers, and interface
  props consistently.
- **Skip separation is correct.** `StructField.Skip`
  (`node_wrappers.go:688-714`) owns exclusion and needs no change.
- **Harness mechanics match the plan.** `TestBasic`
  (`internal/builder/basic_test.go:21-186`) copies each fixture to
  `test_run/<name>`, runs `go mod tidy` → `go generate ./...` → golden
  comparison → `go mod tidy` → `go build ./...`. Goldens live in the fixture
  source (e.g. `internal/builder/testfixtures/structs/jsonschema/*.json.golden`)
  and are copied along; `AssertGoldenFile` diffs `<file>` against
  `<file>.golden` (`internal/testutils/golden_file.go:11-23`). The subtest
  name `TestBasic/test4-structs` in the plan's proof command exists
  (`basic_test.go:124`). The fixture is a nested module
  (`testfixtures/structs/go.mod` with a `replace` to the repo root), so it is
  indeed invisible to root `go test ./...`.
- **Baseline is green.** `go test ./...` at repo root passes at HEAD (exit 0),
  including `internal/builder` and the `examples/*` packages.
- **No committed artifact churn.** The only omitted-name JSON tag in the repo
  is `examples/structs/types.go:59` (`ContactInfo \`json:",inline"\``), which
  is an **embedded** field — `PropNames` `case 0` returns early, so the
  planned `case 1` change cannot alter any committed generated output.
- **Issue independence holds.** Issue 29 is the traversal
  exported/unexported reversal in `scan_result.go:skipField` (:583-609) —
  disjoint code path from `PropNames`. Issue 28 consumes this fix but the fix
  itself introduces no Optional/Nullable surface.

## Findings

### 1. design — Plan misses the simplest home for the behavioral `encoding/json` assertion: `examples/structs` is in the root module and already tests generated `Schema()` output

Plan step 3 (`docs/design/issue-30-json-tag-name-plan.md:66-71`) treats the
nested-module problem as forcing a choice between "run the fixture's test
through the harness" and "make the harness run `go test ./...` in generated
fixture modules." Both are workable but heavier than needed:

- `examples/structs` has **no `go.mod`** — it is part of the root module, its
  generated `jsonschema_gen.go` and schemas are committed, and it already
  contains `schema_test.go` making behavioral assertions against generated
  `Schema()` output (e.g. `TestPersonSchemaWithTime`). A test there that
  marshals a populated value with `encoding/json` and compares its key set to
  the generated schema's `properties` keys is discovered by plain
  `go test ./...` — which is precisely the AGENTS.md gate and the plan's own
  completion criterion (line 112). Zero harness changes.
- The harness-modification option changes proof behavior for all nine
  fixtures, and a test file inside `testfixtures/structs` only compiles
  *after* generation (the source fixture has no `jsonschema_gen.go`; only
  build-tagged `schema.go` stubs). That is workable — nothing builds the
  source fixture module directly — but it is a subtle trap for future editors
  and for gopls.

Recommended shape: keep the golden-file fixture coverage in
`testfixtures/structs` (static generation proof) and put the
`encoding/json`-vs-`Schema()` behavioral assertion in `examples/structs`
(runtime proof, root module). Trade-off to acknowledge: examples double as
user-facing docs, so the added type should read as a legitimate example
(e.g. a struct demonstrating `json:",omitzero"`), which — given issue 28 will
document exactly that idiom — it is. If the team rejects mixing regression
proof into examples, then scope the harness `go test` change to the structs
fixture only rather than all fixtures.

### 2. design — Plan omits the AGENTS.md baseline step

AGENTS.md (repo root, `Agents.md`) requires running all tests **before**
starting to establish a baseline. The plan starts at "add a failing unit test"
(step 1, line 32) and only exercises the suite at the end (step 4, line 79).
Add an explicit step 0: `go test ./...` must be green before any edit (I
verified it currently is). Cheap to add, and it is a stated repository
contract the prompt asked the plan to respect.

### 3. nit — Proof command prescribes test names the plan never fixes

Step 4's first command (line 76) filters on
`TestStructField.*PropNames|TestPropNames`, but step 1 never pins the test
function names. Either name the test in step 1 (e.g. adopt the issue's
`TestPropNamesFallsBackToGoNameForEmptyJSONName`) or loosen the `-run`
pattern. As written, a faithful implementer can satisfy step 1 and have the
step 4 command match nothing.

### 4. nit — `json:""` is already handled; worth one characterization case

Because `structTag` returns nil for an empty tag value
(`node_wrappers.go:678-681`), `json:""` already falls back to the Go name
today (verified). The planned non-empty-name check makes the empty-value and
empty-name paths uniform, which is good — but adding `json:""` to the step 1
table costs one line and pins the invariant.

### 5. nit — Known, deliberately-unfixed divergences could be named in Non-goals

Two pre-existing divergences from `encoding/json` sit adjacent to the changed
line and are correctly left untouched, but naming them prevents scope creep
and reviewer re-litigation:

- `json:"-,omitempty"` / `json:"-,"`: `encoding/json` treats only the exact
  tag `"-"` as "ignore"; a `-` name with options is the literal property name
  `-`. This repo's `Skip()` (`node_wrappers.go:702`) skips on `Options[0] ==
  "-"` regardless of options (verified).
- Grouped fields (`A, B int \`json:"x"\``): `PropNames` ignores the tag
  entirely for multi-name fields (`case 1` only). The plan already says
  grouped behavior is out of scope (line 91); listing the tag-on-grouped-field
  case explicitly would make the boundary crisp, and a one-line
  characterization test is cheap.

### 6. nit — Fixture-test option needs testify in the fixture module

If the fixture-module test route is chosen despite finding 1, note that
`testfixtures/structs/go.mod` has no test dependencies today; the harness's
pre-existing `go mod tidy` (basic_test.go:42) will absorb this, but the plan
should say the test must use only stdlib or accept the go.sum churn.

No critical or bug findings against the plan: the proposed production change
(empty-name fallback in `PropNames` `case 1`, no option special-casing) is
correct, minimal, at the right ownership boundary, and cannot disturb any
committed generated artifact.

## Recommended plan shape

1. **Step 0**: `go test ./...` baseline (AGENTS.md), from this worktree.
2. **Unit regression (red first)**: table-driven test in
   `internal/syntax/node_wrappers_test.go` (new file, internal `package
   syntax`, matching existing convention) with pinned test name(s); cases:
   `,omitzero`, `,omitempty`, explicit name, untagged, `json:""`, `json:"-"`
   via `Skip`, plus optional one-line grouped/embedded characterization.
3. **Fix**: `PropNames` `case 1` — use the parsed name only when non-empty;
   otherwise fall through to the exported-name loop. (~3 lines.)
4. **Generation proof**: dedicated struct in `testfixtures/structs` +
   `schema.go` registration + new entry in `TestBasic` test4 file list +
   committed `.golden`.
5. **Behavioral proof**: `encoding/json` key-set vs generated `Schema()`
   properties comparison in `examples/structs/schema_test.go` (root module —
   runs under plain `go test ./...`; no harness changes). Fall back to a
   structs-fixture-scoped harness `go test` only if examples are ruled out.
6. **Close-out**: the plan's step 4 command ladder, with the `-run` pattern
   matched to the actual test names; confirm no empty key, explicit name
   preserved, `-` excluded, key sets equal.

## Evidence I could not verify

- That `go test ./...` inside generated fixture modules produces empty stderr
  across all nine fixtures (relevant only to the plan's harness-modification
  option; I did not modify the harness to try it).
- The issue comment's claim that the four referenced `docs/design` /
  `ephemeral/worklog` links resolve on `main` — the issue-28 docs exist on
  this branch, but I did not check `main` on the remote.
- Exact `encoding/json` output equality for the future fixture type (the type
  does not exist yet); I verified the mechanism and the individual name-
  resolution behaviors it depends on, at HEAD, via the temporary test above.
