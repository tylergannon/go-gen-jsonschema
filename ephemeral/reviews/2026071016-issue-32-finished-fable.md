# Issue 32 finished-product review (Fable, 2026-07-10)

## Exact prompt received

> You are the independent finished-product reviewer for issue 32 in
> go-gen-jsonschema.
>
> Goal: review the complete current worktree implementation against the concrete
> definition of done at `docs/design/issue-32-definition-of-done.md`. The product
> is public `jsonschema.Optional[T]` and `jsonschema.Nullable[T]`, direct-field
> generator support, generated decoder behavior, legacy optional-tag removal,
> and an executable public consumer proof. OpenAI compatibility must follow the
> published strict Structured Outputs rules in the artifact; no live API call or
> credential is part of the proof.
>
> Review the actual current worktree, including tracked and untracked files. Read
> the definition of done, `git diff origin/main`, the runtime types, syntax and
> builder implementation, generated code/template changes, fixtures, public
> docs, `examples/optionality`, its negative fixtures, and its deterministic
> proof transcript. Run relevant commands where useful.
>
> Give a broad, adversarial review focused on exactly these four axes:
>
> 1. correctness;
> 2. absence of over-engineering;
> 3. completeness against the concrete product contract;
> 4. appropriate level of code factoring.
>
> Do not give general approval. Find concrete issues. Pay particular attention to
> presence/null runtime semantics, canonical wrapper recognition, unsupported
> placement diagnostics, Optional requiredness and omitzero, Nullable schema
> encoding, named/scalar/container/object/interface paths, transactional generated
> decoding, legacy behavior removal, generator idempotence, and whether the proof
> actually exercises the public behavior it claims.
>
> Use these labels exactly:
>
> - critical: must fix before proceeding.
> - bug: demonstrable incorrect behavior, broken contract, race, or regression.
> - design: architecture, boundary, scope, maintainability, or proof issue that is
>   materially likely to cause problems.
> - nit: small cleanup that should not block progress.
>
> Write the exact prompt you received followed by your findings to
> `ephemeral/reviews/2026071016-issue-32-finished-fable.md`. Include file and line
> references, evidence, and the smallest appropriate fix. Do not edit product
> code or any other artifact.

## What I ran

- `go test ./...` — all packages pass.
- `go run ./examples/optionality/cmd/proof` — exit 0, transcript matches
  `proof/expected.json` byte-for-byte, all four negative generator fixtures fail
  with the promised diagnostics.
- `go generate ./...` twice — tree byte-identical before/after (idempotent).
- `JSONSCHEMA_NO_CHANGES=1 go generate ./...` — exit 0. `git diff --check` clean.
- `rg -n 'jsonschema:"[^"`]*optional' --glob '*.go' .` — no matches (legacy tag
  fully removed from parsing and fixtures; inert-tag test exists at
  `internal/syntax/node_wrappers_test.go:70-86`).
- Two out-of-tree generator probes against the real CLI/builder (evidence below).

Runtime semantics verified by reading `optionality.go` and its tests: zero-value
absence, `IsZero` contracts, present-zero/present-empty round-trips, null
rejection for Optional, null↔`Present==false` for Nullable, present-null marshal
rejection (nil pointers, NaN/Inf), and receiver preservation on failed decode are
all implemented and unit-tested (`optionality_test.go:9-157`). The regenerated
template decoder is now transactional: `__next` is built and only assigned on
full success (`internal/builder/schemas.go.tmpl:137-157`), and
`test9-v1-interfaces-options/optionality_test.go` proves no-mutation on unknown
discriminators and null Optional interface input. Wrapper recognition is
canonical-path-based including aliases and rejects same-name third-party types
(`internal/syntax/node_wrappers.go:685-726`, tests at
`node_wrappers_test.go:11-69`).

## Findings

### 1. bug — `Nullable[T]` over a legacy `NewEnumType` enum silently generates a self-contradictory schema

- Where: `internal/builder/gen_schema.go:1219-1238` (`nullableSchema`),
  `internal/builder/gen_schema.go:610-613` (`mapType` → `Constants` →
  `mapEnumType`).
- Contract: DoD line 97-99 — "Nullable arrays/slices, consts, enums, registered
  interfaces, explicit refs, providers, and templates **fail generation
  clearly**."
- The guard for enums only covers the `WithEnum` v1 path (`specialSource =
  "enums"`, gen_schema.go:1147). A legacy `jsonschema.NewEnumType[Color]()`
  registration renders through `renderSchema` → `mapType` → `mapEnumType`,
  returning `PropertyNode[string]{Enum: ...}`, which `nullableSchema` accepts
  because it switches only on the node's Go type, never inspecting `Enum`/`Const`.
- Evidence (real CLI run against a copied fixture module in /tmp):

  ```go
  var _ = jsonschema.NewEnumType[Color]()
  type Config struct {
      Shade schema.Nullable[Color] `json:"shade"`
  }
  ```

  generated successfully (exit 0) with:

  ```json
  "shade": { "type": ["string", "null"], "enum": ["red", "blue"] }
  ```

  JSON `null` satisfies `type` but fails `enum`, so the exact value the runtime
  marshals for an absent `Nullable` (`null`) can **never** validate against the
  generated schema. Generation should have failed; instead it emits a schema
  that contradicts the runtime.
- Smallest fix: in `nullableSchema`, reject any `PropertyNode` whose
  `len(Enum) > 0` or `Const != nil` with the existing "V1 supports scalar
  values, structs, and pointers to structs" error (or an enum-specific message
  matching the WithEnum path).

### 2. bug — `[]byte` now generates an array-of-integer schema while encoding/json emits a base64 string

- Where: `internal/builder/gen_schema.go:660` — this diff adds `"byte"` (and
  `"rune"`, `"uintptr"`) to the scalar-integer ident list.
- On `origin/main`, a `Data []byte` field failed generation (ident `byte` fell
  through to named-type lookup → "type byte not found"). With this change it
  generates silently:

  ```json
  "data": { "type": "array", "items": { "type": "integer" } }
  ```

  while `json.Marshal(struct{ Data []byte }{...})` produces
  `{"data":"AQID"}` — a base64 **string**. Verified with a real CLI probe plus
  a runtime marshal check. Every `[]byte` schema this generator now emits is
  wrong for the value Go actually produces, and it validates/rejects exactly
  backwards. (The same mismatch pre-existed for the `[]uint8` spelling, but
  `byte` is the spelling users actually write; this diff widens a latent hole
  into the common path while the DoD's intent was only scalar `byte` fields as
  integers.)
- `Optional[[]byte]` inherits the same mismatch. `[]rune` is fine (array of
  numbers at runtime).
- Smallest fix: in the `*dst.ArrayType` branch of `renderSchema`
  (gen_schema.go:723-731), special-case an element ident of `byte`/`uint8`:
  either return a clear generation error or emit
  `{"type":"string"}` with a base64 description. Add a fixture either way.

### 3. design — delivery state does not meet the DoD's committed-proof requirements

- The branch has **zero commits**: everything (including `optionality.go`,
  `examples/optionality/`, the DoD document itself) is unstaged or untracked
  (`git log origin/main..HEAD` is empty).
- Consequences against the contract:
  - DoD step 5 exit command `git diff --exit-code -- examples/optionality` is
    vacuous — the entire example is untracked, so `git diff` cannot see it.
    "Commit the generated schemas, generated Go code, and expected transcript"
    (DoD :260) is not yet true.
  - Step 7's clean-tree sequence and the fresh-checkout proof cannot be
    meaningfully executed until the candidate is committed.
  - Every implementation checklist box in sections 1–6 is unchecked even though
    the work and exit commands demonstrably pass; the document states "Check an
    item only when its listed acceptance behavior and command both pass," so the
    controlling checklist is stale relative to reality.
- Smallest fix: commit the complete candidate, update the checklist to actual
  state, then run the step-7 closeout sequence against the committed tree.

### 4. design — README overclaims V1 Optional coverage (`enum`, `provider`) with zero proof

- Where: `README.md:189-190` — "V1 `Optional` supports ordinary scalar, struct,
  pointer, array/slice, explicit ref, enum, provider, and registered-interface
  field paths."
- The DoD's V1 Optional list (DoD :92-95) is "scalar and named scalar values,
  structs, pointers, arrays/slices, explicit supported refs, and registered
  interfaces" — providers are not in it, and no test, fixture, or proof case
  anywhere in the tree exercises `Optional[T]` with a provider
  (`TemplateHoleNode` path, gen_schema.go:1179-1193) or with `WithEnum`
  (gen_schema.go:1082 uses `renderType`, so it plausibly works, but nothing
  proves it). Rendered/template types also skip `ValidateJSON`, so
  Optional-with-provider requiredness has no validation backstop.
- Smallest fix: either trim the README sentence to the proven set, or add a
  small fixture covering `Optional` + `WithEnum` and `Optional` + provider
  before claiming them.

### 5. design — legacy-interface path lacks the Nullable rejection that the v1 path has

- Where: `internal/builder/gen_schema.go:1363-1377`. The `IfaceV1` branch
  rejects `Nullable` wrappers explicitly (:1333-1335), but the legacy
  `findInterfaceImpl` branch appends
  `InterfaceProp{..., Optional: wrapper == syntax.WrapperOptional}` for a
  Nullable-wrapped field with no error.
- Today this is masked: schema rendering later fails in `nullableSchema`
  (UnionTypeNode → default branch), so generation still aborts — but with the
  generic "inner schema shape builder.UnionTypeNode is unsupported" message
  instead of the promised "does not support registered interfaces", and only
  because `mapNamedType` happens to render schemas after resolving interface
  props. If that ordering ever changes, the template would emit
  `__next.Field, err = __jsonUnmarshal__...` assigning an interface value into a
  `Nullable[...]` struct — non-compiling generated code.
- Smallest fix: replicate the two-line Nullable guard at the top of the legacy
  branch (before gen_schema.go:1372).

### 6. nit — negative-coverage gaps and a misnamed fixture

- `examples/optionality/negative/nested_wrapper` actually tests a wrapper
  inside a slice (`[]jsonschema.Optional[int]` — container placement), not a
  nested wrapper (`Optional[Optional[int]]`). Both funnel through
  `scan_result.go:490`, but the transcript's "nested wrapper" label documents
  something the fixture doesn't show. Rename it `wrapper_in_container` or make
  it a real nested wrapper.
- No test or fixture anywhere exercises the embedded-wrapper diagnostic
  (`scan_result.go:519`) or a defined wrapper type
  (`type MyOpt jsonschema.Optional[int]`, as opposed to the alias fixture).
  Cheap to add alongside the existing negatives.

### 7. nit — no migration note for the removed `jsonschema:"optional"` tag

- The breaking change is implemented exactly as agreed (tag removed from
  `internal/common/tags.go`, inert-tag behavior tested), and README/llms.txt/
  SKILL.md all describe the new contract. But nothing anywhere tells an
  existing user that regenerating will silently flip previously-optional
  fields to required. One sentence in the README ("the `jsonschema:"optional"`
  tag is no longer honored; use `jsonschema.Optional[T]`") would prevent the
  only silent-behavior-change hazard this migration has.

### 8. nit — small factoring frictions (none blocking)

- `specialSource` (gen_schema.go:1054, 1071, 1147, 1173, 1189) is a
  stringly-typed sentinel that doubles as error text; at this size it's fine,
  but a tiny enum would resist typo drift if a sixth special path appears.
- `StructField.Required()` (node_wrappers.go:649-652) silently discards
  `Wrapper()`'s arity error. Unreachable for code that compiles, but a one-line
  comment saying so would save the next reader the analysis.

## Axis summary (findings only, no approval)

- **Correctness**: two demonstrated bugs (findings 1, 2), one masked guard gap
  (finding 5). The core runtime/decoder/requiredness semantics, canonical
  recognition, transactional decoding, idempotence, and the proof's assertions
  all held up under direct execution.
- **Over-engineering**: none found; the implementation is lean (one new file,
  ~90 lines of syntax classification, a value-copy schema wrapper, template
  glue). Finding 8 is the only factoring note and cuts the other way.
- **Completeness vs contract**: findings 1 (Nullable enum must fail), 3
  (committed-proof requirements unmet), 4 (doc claims exceed proven coverage),
  6 (missing negative coverage for two promised diagnostics).
- **Factoring**: boundaries are appropriate (recognition in syntax,
  requiredness via `Required()`, nullability as a schema transform, decode glue
  in the template); findings 5 and 8 are the residual rough edges.
