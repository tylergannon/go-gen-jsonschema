# Issue 30 plan review — round 2 (Claude Sonnet)

Reviewer: Claude (Sonnet), 2026-07-10
Worktree: `/Users/tyler/src/go-gen-jsonschema-issue-30`, branch `codex/issue-30-plan`, HEAD `13421b6`
Plan under review: `docs/design/issue-30-json-tag-name-plan.md` (untracked in this worktree)
Round-1 review: `ephemeral/reviews/20260710-issue-30-plan-round-1-fable.md`

## Exact prompt given

```
Here is my goal:

Read GitHub issue 30 in this repository and converge on a correct, minimal,
implementation-ready written plan for fixing omitted JSON tag names so generated
schema property names agree with Go's encoding/json behavior. The work is
planning only: do not implement the fix. Keep issue 30 independent of issues 28
and 29 except for documenting relevant constraints. Respect the repository's
AGENTS.md test requirements and the user's separate-worktree requirement.

This is review round 2. Review the revised plan at:
docs/design/issue-30-json-tag-name-plan.md

The round-1 Fable review is at:
ephemeral/reviews/20260710-issue-30-plan-round-1-fable.md

Confirm whether the revision actually resolves the round-1 design findings,
then perform a broad review of the whole revised plan. Use `gh` to read issue 30
and inspect the relevant source, tests, generated examples, fixtures, and
repository conventions yourself. Look for correctness problems, hidden
dependencies, wrong ownership boundaries, over-engineering, unrequested
features or layers, missing test cases, proof gaps, and simpler implementation
paths. Do not reduce the review to checking the previous findings.

Do not edit product code or the plan. Write the exact prompt you were given and
your findings to:
ephemeral/reviews/20260710-issue-30-plan-round-2-sonnet.md

Label every finding using these definitions:

- critical: must fix before proceeding.
- bug: demonstrable incorrect behavior, broken contract, race, or regression.
- design: architecture, boundary, scope, maintainability, or proof issue that is
  materially likely to cause problems.
- nit: small cleanup that should not block progress.

Give file and line references where useful. End with a consensus verdict,
including any unresolved dissent and whether another review round is warranted.
```

## What I verified independently

- **Baseline is green.** `git status` at HEAD `13421b6` (branch
  `codex/issue-30-plan`, tracking `origin/main`) shows only the untracked plan
  doc and `ephemeral/reviews/`; no product code changed since round 1.
- **Bug still reproduces at HEAD**, confirmed by re-reading
  `internal/syntax/node_wrappers.go:648-664` (unchanged since round 1):
  `case 1` returns `tag.Options[0]` unconditionally when a JSON tag exists,
  so `json:",omitzero"`/`json:",omitempty"` still yield `[""]`.
- **Ownership claim re-verified independently**, not just re-checked against
  round 1. Grepped every `PropNames()` call site
  (`internal/builder/gen_schema.go:166,1017,1171,1185,1265,1294,1313`) and
  every direct `Options[0]` read in the repo
  (`node_wrappers.go:654,702`; `scan_result.go:598`). The only non-`PropNames`
  reads are `Skip()`-style `"-"` checks, which don't consume the name. Fixing
  `PropNames` alone is still sufficient and correctly scoped.
- **Issue 30's own text** (`gh issue view 30`) matches the plan's stated
  root cause, expected behavior, and acceptance proof verbatim; the plan does
  not drift from the issue.
- **Issue 28's comment on issue 30** points to `docs/design/issue-28-optional-plan.md`
  et al. for context only; it does not impose additional requirements on issue
  30's fix, consistent with the plan's independence goal.
- **`examples/structs` and `internal/builder/testfixtures/structs` conventions**
  read directly: `examples/structs/schema.go`, `types.go`, `schema_test.go`
  (existing `TestPersonSchemaWithTime`, `TestPersonJSONMarshalUnmarshal`
  patterns) and `testfixtures/structs/struct_types.go`, `schema.go`,
  `internal/builder/basic_test.go:122-131` (`TestBasic/test4-structs` file
  list, golden-file mechanics). The plan's steps 3 and 4 match these
  conventions precisely — a new type in `testfixtures/structs` needs only a
  `struct_types.go` addition, a `schema.go` registration, and a new
  `jsonschema/<Name>.json` entry + `.golden` in the `test4-structs` file list;
  a new type in `examples/structs` needs a `types.go`/`schema.go` addition and
  a `go:generate` re-run, and is picked up by root `go test ./...` with no
  harness change, exactly as the plan states.
- **No other omitted-name JSON tags exist in the repo** (confirmed by
  grep for `omitzero`/`omitempty`); every existing use pairs an explicit name
  with the option, so the fix cannot alter any already-committed generated
  artifact (matches round-1's finding, re-verified).
- **Go 1.26.5 is installed** (`go version`) and `go.mod` declares `go 1.26`,
  so `omitzero` is available to both production and test/fixture code.
- **`.sum` sidecar files are a real generator output.** `writeSchema`
  (`internal/builder/gen_schema.go:754-768`) writes `<file>.sum` next to every
  `.json`/`.json.tmpl` schema file it emits. `examples/structs/jsonschema/`
  has committed `.sum` files for each existing type (e.g. `Address.json.sum`);
  `testfixtures/structs/jsonschema/` has none, because those files are
  produced transiently in `test_run/` during `TestBasic` and only the listed
  `.json` files are golden-compared (`basic_test.go:57-63`). This confirms
  the plan's two proof locations behave differently by design, not by
  omission.

## Round-1 finding resolution check

| # | Round-1 finding | Resolved? | Evidence |
|---|---|---|---|
| 1 | Put static proof in `testfixtures/structs`, behavioral proof in `examples/structs`, not a harness change | **Yes** | Plan step 3 (lines 62-75) and step 4 (lines 77-91) match this split exactly. |
| 2 | Missing AGENTS.md baseline step | **Yes** | Plan step 0 (lines 32-40) added, with the observed-green baseline commit pinned (`13421b6`). |
| 3 | Proof command names a test the plan never pins | **Partially — recurs elsewhere, see Finding 1 below** | Step 1 now names `TestStructFieldPropNames` (line 42) and step 5's first command matches it (line 96). But step 5's *third* command (line 98) still names a test (`SchemaPropertyNamesMatchJSON`) that step 4 never pins. |
| 4 | `json:""` characterization case | **Yes** | Added to the step 1 case list (line 50). |
| 5 | Name the two pre-existing divergences to prevent scope creep | **Yes** | Both now appear as explicit Non-goals (lines 113-117): grouped-field tag-ignoring and `json:"-,omitempty"`. |
| 6 | Fixture-module test route needs testify or stdlib-only note | **Moot, correctly so** | The plan never adds a `go test` file inside `testfixtures/structs`; it only adds a golden `.json` fixture (step 3) and puts the only new test in root-module `examples/structs` (step 4), sidestepping the concern entirely. |

Round 1's findings were all either fully resolved or rendered moot by the
chosen approach, with one exception (#3) whose underlying pattern reappears
in a new spot — see Finding 1.

## New findings (round 2)

### 1. nit — Step 5's third proof command names a test that step 4 never pins

`docs/design/issue-30-json-tag-name-plan.md:98`:

```
go test -run 'SchemaPropertyNamesMatchJSON' -count=1 -v ./examples/structs
```

Step 4 (lines 77-91), which specifies the new `examples/structs` behavioral
assertion, never names the test function — it only describes what the test
must do ("marshal a populated value with `encoding/json`, parse the generated
`Schema()` document, and compare the two property-key sets directly"). This is
the same class of gap round-1 finding #3 flagged for step 1/step 5's first
command, and that instance was fixed by naming `TestStructFieldPropNames` in
step 1. The fix wasn't generalized: the newly added step 4/step 5 pairing has
the identical defect. As written, an implementer can satisfy step 4 with a
test named anything (e.g. `TestStructSchemaPropertyNamesMatchEncodingJSON`,
following the existing file's naming style in
`examples/structs/schema_test.go:10,52,108,136,176`) and the step 5 proof
command would match zero tests, silently passing `go test -run` with no
tests executed rather than failing loudly.

Fix: name the test in step 4 (e.g. `TestStructSchemaPropertyNamesMatchJSON`,
matching the file's existing `TestXxx` convention) and use that exact name in
step 5.

### 2. nit — The fix introduces a new sibling-name collision surface that isn't in the test matrix

`internal/builder/gen_schema.go:1011-1020` (`renderStructProps`) only uses
`SeenProps` to suppress an embedded field's property when the *outer* struct
already claimed that name; it does not detect or reject two *sibling* fields
at the same struct level producing the same `PropNames()` result (props are
simply appended into `ObjectPropSet`, e.g. `gen_schema.go:1185-1191`).

Before this fix, two omitted-name fields at the same level would both
silently collide on `""`. After this fix, an omitted-name field's fallback
(its own Go identifier) can newly collide with a *sibling field's already-
explicit* JSON name, e.g.:

```go
type T struct {
    Retries   string `json:"MaxRetries"`
    MaxRetries int   `json:",omitzero"`
}
```

Both fields now resolve to `"MaxRetries"`. This collision could not occur
before the fix (the omitted-name field always mapped to `""`), so the fix
changes — not just corrects — the collision surface. This is a narrow edge
case, not required by the issue's acceptance criteria, and the existing
collision-handling gap is pre-existing and out of scope to fix here. Given
the prompt's ask to look for "missing test cases," it's worth a one-line
mention (in Risks or Non-goals) that same-level name collisions, including
ones newly reachable through this fix, remain unhandled and untested — so a
future report against this exact scenario isn't mistaken for a regression
from this change.

### 3. nit — Step 4 doesn't name the `.json.sum` sidecar file the generator will also produce

`internal/builder/gen_schema.go:768` writes a `<file>.json.sum` next to every
generated `.json`/`.json.tmpl` schema, and `examples/structs/jsonschema/`
already has one committed per existing type (e.g. `Address.json.sum`,
confirmed via `find`/`git ls-files`). Step 4 (lines 84-87) says "run the
existing generator so the schema and generated Go code are updated" but
doesn't name the `.sum` file explicitly. This is a natural byproduct of
"run the generator" and low risk, but since the plan is otherwise precise
about exactly which files each step touches (step 3 names `.golden` files
explicitly), naming the `.sum` file too would prevent a partial `git add`
that leaves a stray untracked/modified `.sum` file after implementation.

## Findings not raised (explicitly checked and found sound)

- **The minimal fix is correctly scoped and there's no simpler path.** The
  described change (`case 1`: return the parsed name only when non-empty,
  otherwise fall through to the existing exported-name loop) is a ~3-line
  diff at the single correct ownership boundary. Re-reading
  `node_wrappers.go:648-664`, the fallthrough already works today when
  `JSONTag()` returns nil (untagged fields reach the loop below the switch);
  adding an emptiness check to the tagged branch requires no new control
  flow and cannot introduce a fallthrough bug.
- **No unrequested layers or features.** Nothing in the plan touches the
  struct-tag parser, `Skip()`, required/optional semantics, or issue 28/29
  surfaces. Confirmed by re-reading the full diff surface the plan proposes
  against the actual call graph, not just re-trusting round 1's claim.
- **Baseline/worktree/AGENTS.md instructions are honored.** `Agents.md`
  (repo root) requires a pre-work `go test ./...` baseline and a completed
  `go test ./...` at the end; plan step 0 and step 5's final command satisfy
  both. The worktree at `/Users/tyler/src/go-gen-jsonschema-issue-30` is a
  real separate checkout on its own branch, consistent with "the user's
  separate-worktree requirement."
- **No committed generated-artifact churn from the fix itself.** Still true:
  the only omitted-name tag in the repo today is the embedded
  `ContactInfo` field (`examples/structs/types.go:59`), which
  `PropNames`' `case 0` handles and this fix does not touch.

## Consensus verdict

**Approve to proceed, with two cheap nits worth folding in before
implementation starts (not blocking):**

- Finding 1 (pin the `examples/structs` test name in step 4) directly closes
  the reopened instance of round-1 finding #3 and should be fixed for
  consistency — it's the same one-line class of fix already applied
  elsewhere in this same plan.
- Findings 2 and 3 are genuinely minor (an edge case out of the issue's
  acceptance criteria, and a byproduct file that "run the generator" already
  covers implicitly) and can be folded in opportunistically or left as-is
  without materially increasing risk.

No critical, bug, or design findings against this revision. All five
substantive round-1 findings are resolved or rendered moot by the chosen
approach; the one recurring pattern (unpinned test name) is downgraded to a
nit here because it repeats a known-cheap fix rather than introducing a new
class of risk. The production change itself remains correct, minimal, and at
the right ownership boundary, and the two-location proof strategy (static
golden fixture + root-module behavioral test) is sound and matches existing
repository conventions exactly.

**No further full review round is warranted.** If the two nits are folded in,
this plan is implementation-ready as written. No unresolved dissent with the
round-1 review.
