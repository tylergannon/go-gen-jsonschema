# Issue 32 contract review (Opus, independent)

Status: findings for checkpoint-1 review. **Not an approval.**
Date: 2026-07-10
Reviewer scope: issue #32 as an untrusted proposal, plus
`docs/design/issue-32-checkpoint-plan.md`. Read-only; no product code or plan
was modified.

## Exact prompt I was given

> Here is my goal:
>
> Re-research GitHub issue #32 as an untrusted proposal and define a safe,
> reviewable implementation sequence for explicit absent and nullable Go field
> semantics. The result must culminate in real behavioral proof through the public
> generator CLI, compiled generated consumer code, JSON marshal/unmarshal,
> generated validation, and—if available—the live OpenAI Structured Outputs API.
> The existing issue body's large checklist and implementation prescriptions are
> not authority.
>
> Review `docs/design/issue-32-checkpoint-plan.md` before implementation. Inspect
> issue #32, current source, tests, existing issue-28 design artifacts, and git
> history as useful. Look for incorrect semantics, unnecessary API or scope,
> missing compatibility constraints, wrong ownership boundaries, hidden decoder
> problems, unsafe checkpoint ordering, and proof that would still be fake or
> insufficient. Prefer a smaller truthful vertical slice over accepting the issue
> body wholesale.
>
> Do not modify product code or the plan. Write the exact prompt you were given
> and your findings to:
>
> `ephemeral/reviews/2026071014-issue-32-contract-opus.md`
>
> Label every finding using exactly these definitions:
>
> - **critical:** must fix before proceeding.
> - **bug:** demonstrable incorrect behavior, broken contract, race, or regression.
> - **design:** architecture, boundary, scope, maintainability, or proof issue
>   that is materially likely to cause problems.
> - **nit:** small cleanup that should not block progress.
>
> Do not approve the plan. Produce concrete findings, or explicitly state that
> you found none at a given severity.

## What I verified firsthand (not disputed)

These parts of the plan check out against the current tree at
`efab995` (`fix: repair exported field traversal (#35)`):

- Requiredness derives from `jsonschema:"optional"` via
  `StructField.Required()` (`internal/syntax/node_wrappers.go:639`) and lands as
  `ObjectProp.Optional = !f.Required()`
  (`internal/builder/gen_schema.go:1189`); the `required` array is emitted by
  excluding `Optional` props (`internal/builder/model.go:180-196`). ✔
- Property schema is built *before* `ObjectProp` construction, so the
  direct-struct-field boundary is a coherent place to classify wrappers. ✔
- Generated `ValidateJSON` compiles the emitted schema once in `init()` and
  validates raw JSON bytes (`internal/builder/schemas.go.tmpl:37-119`), so it can
  enforce Nullable required-key presence *at the JSON layer*, independent of Go
  decode. ✔
- Custom `UnmarshalJSON` is generated only for "special" types carrying
  registered-interface fields (`schemas.go.tmpl:124-146`); ordinary fields use
  stock `encoding/json`. ✔
- `JSONSCHEMA_NO_CHANGES=1` is a real idempotence guard
  (`internal/builder/builder.go:57`, README §, `.github/workflows/go.yml:33`), so
  the proof-sequence step using it is valid. ✔
- The scaffolded (untracked) `examples/optionality` currently demonstrates the
  **real** present-zero-vs-absent presence-loss bug through the actual generator
  (`--validate`) and generated `ValidateJSON`. That is honest red evidence of the
  problem — it proves the motivation, not the solution. ✔

Note: the branch has **no commits ahead of `main`**. The checkpoint plan, the
worklog, and `examples/optionality/` are all untracked. There is no product
change to review yet; this review is of the contract and sequence only.

---

## Findings

### critical

**C1 — The plan's "Current generator boundary" misstates existing capability;
generic field types are unhandled in BOTH field walkers, so the first slice and
its ownership boundary are wrongly scoped.**

The plan asserts (checkpoint plan, "Current generator boundary"):

> The syntax layer already preserves generic `IndexExpr` shape and canonical
> import paths, including aliases.

That is true only of the *registration-argument* parser
(`internal/syntax/scan_expr.go:42,105`), which reads marker call expressions
like `NewJSONSchemaMethod(...)`. It is **not** true of the two code paths that
actually walk a struct field's type:

- Discovery/dependency resolution: `ScanResult.resolveTypeExpr`
  (`internal/syntax/scan_result.go:486-568`) has cases for `ParenExpr`,
  `StarExpr`, `ArrayType`, `MapType`, `InterfaceType`, `StructType`, `Ident`,
  `SelectorExpr`, `BasicLit` — and **no `*dst.IndexExpr` case**. A generic field
  type falls to `default` → `return fmt.Errorf("unhandled expression %s", …)`.
- Rendering: `SchemaBuilder.renderSchema`
  (`internal/builder/gen_schema.go:656-736`) likewise has **no `*dst.IndexExpr`
  case**.

`#35` just *activated* exported-field traversal (previously dormant), so this is
now on the hot path: today, any registered struct with an exported field of any
generic type — including the proposed `Optional[T]`/`Nullable[T]` — fails
generation before rendering. Decision #13's "normalize a recognized direct field
once and send the inner expression through existing rendering" therefore rests on
a false premise: "existing rendering" does not accept the field position, and
classification must be added to **both** independent walkers or discovery errors
before rendering ever runs. The real ownership boundary is "every struct-field
type walker," not a single point. Correct the boundary description and the
first-slice scope before checkpoint 2/3 is planned against it.

### bug

None demonstrable beyond the unimplemented-traversal behavior folded into C1. I
looked specifically for a broken contract in the current tree (requiredness,
`omitzero`/`IsZero`, validation opt-in, interface decode) and did not find a
standalone regression to report at this severity. Stated explicitly per the
instructions.

### design

**D1 — Nullable-object strict compatibility is asserted but unverified, and may
be false.** Issue decision #9 emits Nullable structs as `anyOf:[<inlined
object>, {"type":"null"}]`, and decision #12 claims "ordinary plus supported
Nullable fields can remain strict-compatible." The compact scalar form
`type:["integer","null"]` is OpenAI-documented; the `anyOf:[object,null]` form
for objects is **not** confirmed under OpenAI strict mode. If OpenAI rejects a
null branch in a property-level `anyOf`, decision #12 is false for the object
case and Nullable structs are silently non-strict. The plan's proof contract
must treat Nullable-*object* strictness as a distinct live-probe case with a
documented fallback (`type:["object","null"]` with inlined properties), not
assume the scalar result generalizes. (The issue-28 Fable review §3.3 reached the
same conclusion for refs.)

**D2 — Anthropic is a named target but appears nowhere in the plan's
compatibility surface.** `CLAUDE.md` states the generator is "optimized for LLM
function calling (OpenAI, Anthropic)." Anthropic strict tool-use caps
`anyOf`/type-array usage (documented ~16-parameter budget). A Nullable-heavy
schema can hit a *provider* ceiling that has nothing to do with OpenAI's
"all-required" rule. The plan and proof contract discuss only OpenAI strictness.
At minimum the skill/README selection guidance and the proof matrix should state
that nullable unions are not free on every provider (issue-28 Fable §6.6).

**D3 — Checkpoint 2's "commit a consumer that initially fails because the
wrappers do not exist" is a repo-wide hazard as written.** A committed
`examples/optionality` package that references non-existent `jsonschema.Optional`
does not fail one test — it fails to compile, which breaks `go build ./...`,
`go test ./...`, and CI for the entire branch, blocking all other checkpoints.
"Red" here must mean an isolated failure surface: a separate module (as
`internal/builder/test_run/*` already do), an expected-generation-error
assertion, or a driver that asserts today's behavior — not a non-compiling
in-module consumer. (The current scaffold sidesteps this correctly by exercising
only the legacy tag and compiling; the plan's wording invites the opposite.)

**D4 — Two structurally identical wrapper types with divergent `Present`
semantics and exported mutable fields invite misuse and runtime-only failures.**
Both `Optional[T]` and `Nullable[T]` are `{Present bool; Value T}`, but `Present`
means "key was present" for Optional and "value is non-null" for Nullable — the
same field name, opposite meaning, assignable to each other's shape. Exported
`Present`/`Value` also let callers construct invariant-violating states (e.g.
`Present:true` on a value that marshals as `null`, or a present nil pointer —
decision #11 lists pointers as supported), which the contract says must be
*rejected at marshal time*, i.e. a runtime error rather than a compile-time or
constructor-time guard. Consider unexported fields plus constructors, or at least
document the present-nil-pointer / cross-assignment footguns explicitly. Decision
#15's "keep the API small" should not be read as "expose raw fields with
enforced-only-at-marshal invariants."

**D5 — Optional/Nullable over registered interfaces hides a decoder-ownership
problem the issue understates.** Issue decision #11 lists registered interfaces
as simply "supported" for Optional. But the generic wrapper's own
`UnmarshalJSON` cannot perform discriminator dispatch: `json.Unmarshal` into an
interface-typed `Value` fails, and the per-package
`__jsonUnmarshal__<Iface>` decoder generated into the consumer
(`schemas.go.tmpl:148-181`) is invisible to the wrapper in the `jsonschema`
package. Interface support therefore requires generated per-field glue in the
existing `SpecialTypes` custom-unmarshaler path — the highest-risk integration
point — not the wrapper alone. The plan is right to defer interfaces to
checkpoint 3 behind a red decoder-state matrix; the *issue* is wrong to present
them as a flat "supported." Also: absent registered-interface bytes currently
hard-error (`json.Unmarshal(nil, …)`); that error text is observable behavior and
should be golden-tested before the template changes (issue-28 Fable §6.7).

### nit

**N1 — The issue's prerequisite section is stale.** Issue #32 says "Issue #29 is
the remaining prerequisite … must land as one safe repair." `#35` ("Closes #29")
already landed on `main` and repaired exported-field traversal, the
registered-interface lookup at `scan_result.go:539`, and `time.Time` exclusion at
`scan_result.go:544`. The checkpoint plan's "Current generator boundary" should
also acknowledge that traversal is now *active* — which is precisely why the
missing `IndexExpr` case (C1) now matters.

**N2 — Proof-sequence omits the `--validate` dependency.** The plan's evidence
sequence lists `go generate ./examples/optionality`, but the `ValidateJSON` proof
depends on the `--validate` flag (it is opt-in;
`gen-jsonschema/main.go:67`). The flag is already embedded in the example's
`go:generate` directive, so this is cosmetic — but the plan reads as if
`ValidateJSON` exists unconditionally.

---

## Bottom line

The wire semantics (three distinct contracts), the Go `omitzero`/`IsZero`
mechanics, the validation-only-catches-presence-at-the-JSON-layer caveat, and the
insistence on generator→generated-consumer proof are sound and well-sourced. The
plan's checkpoint discipline is good. But the plan should **not** proceed to
implementation planning on its current "Current generator boundary" text (C1):
the two field walkers do not handle generic types today, which changes both the
first-slice scope and the ownership story. Nullable-object strictness (D1) and
Anthropic budget (D2) are missing compatibility constraints that the proof
contract must not assume away, and the checkpoint-2 red-example wording (D3) is a
CI hazard as written. I recommend a smaller first slice than either the issue or
the plan implies: ordinary + `Optional[scalar]` + `Nullable[scalar]` only, with
`IndexExpr` classification wired into both walkers and proven end-to-end, before
any struct/pointer/container/interface expansion.
