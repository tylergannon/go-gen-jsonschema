# Issue 32 final pre-delivery review (Sonnet, 2026-07-10)

## Exact prompt received

> You are the final follow-up reviewer for issue 32 in go-gen-jsonschema.
>
> Review the complete current worktree against
> `docs/design/issue-32-definition-of-done.md` on correctness, absence of
> over-engineering, completeness, and appropriate factoring.
>
> Prior reviews:
>
> - `ephemeral/reviews/2026071016-issue-32-finished-fable.md`
> - `ephemeral/reviews/2026071017-issue-32-followup-sonnet.md`
>
> The prior Sonnet review found that `Nullable[[]byte]` bypassed the V1 slice
> rejection after ordinary byte slices were correctly rendered as base64 strings.
> The revised code now rejects Nullable array/slice Go shapes before applying the
> schema transform. The executable proof also added real-generator cases for
> Nullable byte slices, explicit refs, and providers.
>
> Independently verify that fix and review the full product for remaining
> critical, bug, or material design issues. Inspect tracked and untracked files,
> the full diff, runtime/syntax/builder/template code, generated outputs,
> fixtures, docs, and `examples/optionality`. Run relevant tests/proof and
> generator probes. Do not edit product code. If you create probe artifacts,
> keep them outside the worktree or remove and verify them before finishing.
>
> The branch is intentionally uncommitted until this pre-delivery review
> converges. OpenAI compatibility follows the published strict Structured Outputs
> rules; no live API call or credential is part of proof.
>
> Use these labels exactly:
>
> - critical: must fix before proceeding.
> - bug: demonstrable incorrect behavior, broken contract, race, or regression.
> - design: architecture, boundary, scope, maintainability, or proof issue that is
>   materially likely to cause problems.
> - nit: small cleanup that should not block progress.
>
> Write the exact prompt received followed by findings to
> `ephemeral/reviews/2026071018-issue-32-final-sonnet.md`, with evidence and
> file/line references. Do not edit any other artifact.

## What I ran

- `go build ./...` — clean.
- `go test ./...` — all packages pass.
- `go generate ./examples/optionality` then `go run ./examples/optionality/cmd/proof`
  — exit 0; transcript unchanged (`git diff --exit-code -- examples/optionality`
  exits 0), all 13 negative fixtures (`optional_without_omitzero`,
  `nullable_slice`, `nullable_byte_slice`, `nested_wrapper`,
  `wrapper_in_container`, `wrapper_alias`, `defined_wrapper`,
  `embedded_wrapper`, `wrapper_root`, `nullable_enum`, `nullable_interface`,
  `nullable_ref`, `nullable_provider`) fail through the real CLI with the
  expected diagnostics.
- `go generate ./...` repo-wide — `git status --short` unchanged before/after
  (idempotent; no drift beyond the pre-existing modified/untracked set).
- Direct reads of the full `git diff origin/main` for every changed file:
  `optionality.go`, `optionality_test.go`, `internal/syntax/node_wrappers.go`,
  `internal/syntax/node_wrappers_test.go`, `internal/syntax/scan_result.go`,
  `internal/builder/gen_schema.go`, `internal/builder/model.go`,
  `internal/builder/schemas.go.tmpl`, `internal/common/tags.go`, `README.md`,
  `llms.txt`, `skills/go-gen-jsonschema/SKILL.md`,
  `docs/internal-dev-notes.md`, `examples/test_options/types.go`,
  `internal/builder/basic_test.go`, and the `test_run`/`testfixtures` golden
  diffs.
- `internal/builder/test_run/test9-v1-interfaces-options/optionality_test.go`
  read directly to confirm transactional/null-rejection behavior for Optional
  registered interfaces is actually exercised (not just claimed).
- `internal/builder/test_run/test12-optionality/jsonschema/Config.json` read
  directly to confirm required/optional/nullable schema shapes for
  struct/pointer/slice/scalar fields match the DoD table exactly.
- A real-CLI probe reproducing the README's own worked example verbatim
  (`Person` struct with `Optional[string]`/`Nullable[string]` fields), built
  in `/tmp/readme-probe` with a `replace` directive pointing at this worktree
  and a binary built via `go build -o /tmp/gen-jsonschema-bin ./gen-jsonschema`.
  Both the probe directory and the binary were deleted after inspection
  (`rm -rf /tmp/readme-probe /tmp/gen-jsonschema-bin`, verified empty
  afterward). No files inside the worktree were created, edited, or left
  behind by this probe.

## Verification of the targeted fix

**Confirmed fixed.** `internal/builder/gen_schema.go:1206-1208` now checks
`renderType` — the wrapper's *pre-collapse* Go type expression captured at
`gen_schema.go:1067-1070` — for `*dst.ArrayType`, before any byte/uint8
short-circuit in `renderSchema` (`gen_schema.go:723-728`) has a chance to turn
`[]byte`/`[]uint8` into a `PropertyNode[string]`. Because the check inspects
the original AST shape rather than the resulting schema shape, it fires
identically for `Nullable[[]int]` and `Nullable[[]byte]`:

```go
if wrapper == syntax.WrapperNullable {
    if _, isArrayOrSlice := renderType.(*dst.ArrayType); isArrayOrSlice {
        return nil, fmt.Errorf("%s does not support arrays/slices at %s", wrapper, f.Position())
    }
    ...
```

Verified with the real generator via the committed
`examples/optionality/negative/nullable_byte_slice` fixture, wired into
`examples/optionality/cmd/proof/main.go:116` with expected substring "does not
support arrays/slices", and confirmed in the actual proof run above (exit 0,
`"nullable byte slice"` / `"does not support arrays/slices"` present in the
`rejected_shapes` transcript). The two other gaps the prior round's Finding 2
listed (Nullable + explicit ref, Nullable + provider) are also now covered by
real fixtures (`nullable_ref`, `nullable_provider`) and exercised through the
real CLI, closing that review's nit as well.

I also independently re-derived why this fix is correct rather than
coincidental: `renderType` is set once, before any special-source branch runs
(`gen_schema.go:1067-1070`), and the array/slice check reads it directly — it
does not depend on which downstream code path (byte/uint8 special case vs.
ordinary `ArrayNode` construction) would have handled the element type. This
closes the class of bug, not just the reported instance.

## New findings

### 1. design — README's own worked example prints a JSON schema that does not match what the generator actually produces

- Where: `README.md:96-152`. The `Person` struct example
  (`README.md:107-119`) declares `Phone jsonschema.Nullable[string] \`json:"phone"\`\`,
  and the "produces `jsonschema/Person.json`" block right below it
  (`README.md:143-152`) shows:

  ```json
  {
    "type": "object",
    "description": "Person is a single contact extracted from the document.",
    "properties": {
      "name": {"type": "string", "description": "Full legal name, e.g. \"Ada Lovelace\"."},
      "age": {"type": "integer", "description": "Age in whole years at the time of writing."},
      "email": {"type": "string", "description": "Email address. Omit when not stated in the source text."}
    },
    "required": ["name", "age", "phone"],
    "additionalProperties": false
  }
  ```

  `"phone"` appears in `required` but has **no entry in `properties`** — the
  block was edited to add `phone` to `required` (correctly reflecting that
  Nullable stays required) without adding the corresponding `phone` property
  entry that shows the actual point of the example: the `["string","null"]`
  type union.

- Evidence: I reproduced this exact struct/tag combination through the real
  `gen-jsonschema` CLI built from this worktree (see "What I ran" above,
  probe deleted after inspection). The real output is:

  ```json
  {"type":"object","description":"Person is a single contact extracted from the document.","properties":{"name":{"type":"string","description":"Full legal name, e.g. \"Ada Lovelace\"."},"age":{"type":"integer","description":"Age in whole years at the time of writing."},"email":{"type":"string","description":"Email address. Omit when not stated in the source text."},"phone":{"type":["string","null"],"description":"Phone number. Emit null when the source explicitly has no phone number."}},"required":["name","age","phone"],"additionalProperties":false}
  ```

  This confirms the product code is correct (the generator does emit the
  `phone` property with the expected `["string","null"]` union and
  description) — the bug is purely in the README's illustration, which is now
  self-contradictory: a `required` key with no matching `properties` entry is
  not a shape the real generator (or JSON Schema in general, meaningfully)
  ever produces.
- Impact: this is the single worked example in the README that demonstrates
  both new wrapper types together, and it is the primary onboarding artifact
  for this feature. A reader copying the "expected output" to sanity-check
  their own setup would see a schema shape the tool never actually emits,
  specifically omitting the one property (`phone`) whose rendering is the
  entire point of introducing `Nullable[T]`. This is exactly the DoD's
  hygiene requirement that "README, `llms.txt`, examples, and the checked-in
  skill agree on the three contracts... and usable examples" (DoD hygiene
  section) — this example is not usable as shown.
- Smallest fix: add the missing `"phone"` entry to the illustrated JSON block:

  ```json
  "phone": {"type": ["string", "null"], "description": "Phone number. Emit null when the source explicitly has no phone number."}
  ```

  (placed after `"email"`, matching field declaration order, per the tool's
  own deterministic-ordering guarantee documented two paragraphs above this
  example).

## Axis summary (findings only, no approval)

- **Correctness**: the previously reported `Nullable[[]byte]` bug is
  genuinely fixed, verified with a real generator run through the committed
  fixture and independently reasoned about from the code structure (not just
  re-reading the diff). No new product-code correctness bugs found in
  runtime, syntax classification, schema generation, or template/decoder
  code. Wrapper recognition, requiredness, nullable scalar/object encoding,
  omitzero enforcement, interface decode transactionality/null-rejection, and
  integer/float width coverage were all independently re-derived from code
  and cross-checked against fixture output (`test12-optionality/Config.json`)
  and tests (`test9-v1-interfaces-options/optionality_test.go`), not merely
  assumed from prior review claims.
- **Over-engineering**: none found. The fix for the targeted bug is a
  two-line AST-shape check reusing an existing captured value
  (`renderType`); no new abstraction was introduced to close it.
- **Completeness vs. contract**: functionally complete against the DoD's
  generator-acceptance and behavioral-proof sections, confirmed by exercising
  scalars, structs, pointers, slices, interfaces, refs, and providers under
  both wrappers via committed fixtures and the real CLI. Finding 1 is a
  completeness gap in the *documentation* deliverable specifically (DoD
  hygiene: "usable examples"), not in the generator or runtime.
- **Factoring**: unchanged from the prior round's assessment — boundaries
  remain appropriate (syntax recognition, requiredness via `Wrapper()`,
  nullability as an isolated schema transform, decode glue confined to the
  template). No new factoring concerns found.

## Sequencing note (per prompt instruction, not a finding)

The branch remains uncommitted, as stated in the prompt as the known,
intentional pre-delivery state. All product files listed in the DoD are
present and functioning as described; this is not reported as a new finding.
