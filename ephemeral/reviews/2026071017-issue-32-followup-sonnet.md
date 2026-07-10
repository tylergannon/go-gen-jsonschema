# Issue 32 follow-up review (Sonnet, 2026-07-10)

## Exact prompt received

> You are the follow-up independent reviewer for issue 32 in go-gen-jsonschema.
>
> Review the revised finished product against
> `docs/design/issue-32-definition-of-done.md` and the user's four required axes:
> correctness, absence of over-engineering, completeness, and appropriate code
> factoring.
>
> The first finished-product review is recorded at
> `ephemeral/reviews/2026071016-issue-32-finished-fable.md`. Its demonstrated bugs
> and material design findings were addressed. Verify those fixes from the actual
> current worktree, but do not limit the review to them:
>
> - Nullable legacy enum/const schemas now fail generation.
> - slice byte/uint8 schemas now match encoding/json base64 strings.
> - legacy registered-interface Nullable handling has an explicit guard.
> - public support claims were narrowed to the accepted V1 scope.
> - negative proof now separates nested, container, alias, defined, embedded,
>   and root placements and covers Nullable enum/interface failures.
> - the legacy optional-tag migration warning was added.
>
> Inspect tracked and untracked files, `git diff origin/main`, runtime wrappers,
> syntax classification, builder/model/template changes, generated outputs,
> fixtures, docs, `examples/optionality`, and the deterministic proof. Run
> relevant commands and real generator probes where useful. The branch is not yet
> committed because this review precedes the clean-tree/exact-head delivery stage;
> do not report that known sequencing state as a new finding unless product files
> are actually missing.
>
> OpenAI compatibility follows the published strict Structured Outputs rules in
> the controlling artifact. No live API call or credential belongs in this proof.
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
> `ephemeral/reviews/2026071017-issue-32-followup-sonnet.md`. Include file/line
> references, evidence, and the smallest appropriate fix. Do not edit product
> code or other artifacts.

## What I ran

- `go build ./...` and `go test ./...` — all packages pass.
- `go run ./examples/optionality/cmd/proof` — exit 0, transcript byte-identical
  to `examples/optionality/proof/expected.json`.
- `go generate ./...` twice in place — `git status --short` identical before and
  after the second run (idempotent); `JSONSCHEMA_NO_CHANGES=1 go generate ./...`
  exits 0.
- Real CLI probes against throwaway fixture directories created under the repo
  root and deleted immediately after inspection (not committed, no diff left
  behind): `jsonschema.Nullable[[]byte]`, `jsonschema.Nullable[string]` with an
  explicit `ref=` tag.
- `rg`/`grep` over `git diff origin/main` for every changed file listed in the
  prompt, plus direct reads of `optionality.go`, `internal/syntax/node_wrappers.go`,
  `internal/syntax/scan_result.go`, `internal/builder/gen_schema.go`,
  `internal/builder/model.go`, `internal/builder/schemas.go.tmpl`, README/llms.txt/
  SKILL.md/internal-dev-notes diffs, `examples/optionality/**`, and
  `ephemeral/worklog/202607101403-issue-32-fresh-research.md`.

## Verification of the six prior fixes

1. **Nullable legacy enum/const now fails generation — confirmed.**
   `nullableSchema`/`nullableProperty` (`internal/builder/gen_schema.go:1224-1244`)
   reject any `PropertyNode` with `len(Enum) > 0 || Const != nil` regardless of
   which path produced it (v1 `WithEnum`/`WithStringerEnum`, which sets
   `specialSource = "enums"` at `gen_schema.go:1152` and is caught by the
   `specialSource != ""` check at `gen_schema.go:1207-1209`, *or* a legacy
   `NewEnumType` registration, which falls through to the generic fallback
   render and is caught by the `Enum`-length check itself). Both paths are
   exercised: `examples/optionality/negative/nullable_enum` and the
   `WithEnum`-registered numeric config. Real probe not needed; the negative
   fixture in `cmd/proof/main.go:122` already asserts the message
   `"enums and consts are unsupported"`.

2. **Slice `byte`/`uint8` now match `encoding/json` base64 — confirmed, but see
   new Finding 1 below.** `gen_schema.go:723-728` special-cases `[]byte`/`[]uint8`
   (`node.Len == nil` guards against fixed-size arrays, which `encoding/json`
   does *not* base64-encode — correct) to `PropertyNode[string]`.
   `examples/optionality/types.go:50` (`Blob jsonschema.Optional[[]byte]`)
   round-trips through the real proof at `cmd/proof/main.go:79` (`"blob":"AQID"`)
   and the generated schema at `examples/optionality/jsonschema/Config.json`
   shows `"blob": {"type": "string"}`. Confirmed correct for the `Optional` and
   ordinary-field cases. The fix does, however, introduce a new gap for
   `Nullable` — see Finding 1.

3. **Legacy registered-interface Nullable guard — confirmed.** Both branches of
   `resolveLocalInterfaceProps` now reject `Nullable`: the v1-config branch at
   `gen_schema.go:1342-1344` and the legacy `findInterfaceImpl` branch at
   `gen_schema.go:1378-1380`, both returning
   `"%s does not support registered interfaces at %s"` before any `InterfaceProp`
   is constructed. `examples/optionality/negative/nullable_interface` exercises
   this through the real CLI (`cmd/proof/main.go:123`).

4. **Public support claims narrowed — confirmed.** `README.md:189-192` now reads
   "V1 `Optional` supports scalar and named scalar values, structs, pointers,
   arrays/slices, explicit supported refs, and registered interfaces. V1
   `Nullable` supports scalars, structs, and pointers to structs" — this matches
   DoD lines 92-99 verbatim in substance (no more `enum`/`provider` overclaim for
   Optional). `llms.txt:73-80` and `skills/go-gen-jsonschema/SKILL.md:161-168`
   were updated to the same scope.

5. **Negative proof restructured — confirmed.** `examples/optionality/negative/`
   now has distinct `nested_wrapper`, `wrapper_in_container`, `wrapper_alias`,
   `defined_wrapper`, `embedded_wrapper`, and `wrapper_root` directories, each
   wired into `cmd/proof/main.go:111-124` with a distinct expected-error
   substring, plus `nullable_enum`, `nullable_interface`, and `nullable_slice`.
   All ten negative cases are invoked through the real `gen-jsonschema` CLI
   (`cmd/proof/main.go:127`) and must match their expected failure text or the
   proof itself fails — this is real generator exercise, not a mocked check.

6. **Legacy-tag migration warning — confirmed.** `README.md:196-198`: "Migration
   note: `jsonschema:\"optional\"` is no longer honored. Replace it with
   `jsonschema.Optional[T]` and add `json:\",omitzero\"`; otherwise the field is
   required when schemas are regenerated." This directly closes the silent
   behavior-change hazard the first review flagged.

All six are real, verified in the current worktree, not just claimed.

## New findings

### 1. bug — `Nullable[[]byte]` (and `Nullable[[]uint8]`) silently generates a schema instead of failing generation

- Where: `internal/builder/gen_schema.go:723-728` (the byte/uint8 → string
  special case, added by prior-round Fix 2) interacts with
  `internal/builder/gen_schema.go:1224-1244` (`nullableSchema`).
- Contract: DoD lines 97-99 — "Nullable arrays/slices, consts, enums,
  registered interfaces, explicit refs, providers, and templates **fail
  generation clearly**." `README.md:191-192` and `llms.txt:79-80` state the same
  narrowed scope this round explicitly added: "V1 `Nullable` supports scalars,
  structs, and pointers to structs" — slices are excluded by name.
- Root cause: for an ordinary `[]int` or `[]string` field, `Nullable[[]T]`'s
  inner schema renders as an `ArrayNode`, which `nullableSchema`'s `switch`
  correctly falls through to the `default` branch and rejects (proved by
  `examples/optionality/negative/nullable_slice`, using `[]int`). But `[]byte`/
  `[]uint8` no longer reaches `ArrayNode` construction at all — the new
  short-circuit at `gen_schema.go:723-728` returns `PropertyNode[string]`
  directly from the `*dst.ArrayType` case in `renderSchema`. `nullableSchema`
  then matches the `PropertyNode[string]` case
  (`gen_schema.go:1228`/`nullableProperty`), sees `Enum` and `Const` both empty,
  and accepts it as an ordinary nullable scalar string — the slice-rejection
  path is never reached for this specific element type.
- Evidence (real CLI run against a throwaway fixture inside this repo, deleted
  immediately after, no diff left behind):

  ```go
  type Config struct {
      Blob jsonschema.Nullable[[]byte] `json:"blob"`
  }
  ```

  generated successfully (exit 0) with:

  ```json
  {"type":"object","properties":{"blob":{"type":["string","null"]}},"required":["blob"],"additionalProperties":false}
  ```

  Generation should have failed with the same "does not support
  arrays/slices"-style diagnostic that `[]int` gets; instead it silently emits a
  schema for a shape the DoD and the freshly-narrowed docs both say V1 does not
  support. (The runtime side is not broken — `optionality.go`'s
  `marshalPresent` still rejects `Present:true` with a nil `[]byte` because that
  marshals to JSON `null` — but the generator contract violation stands on its
  own and is exactly the kind of "silently generates a self-contradictory
  schema" class of bug the previous round's Finding 1 already established for
  enums.)
- Smallest fix: check the Go type shape, not the resulting schema shape, before
  collapsing to a scalar. In `renderStructField` (or in `nullableSchema` by
  passing the pre-collapse type through), reject when `wrapper ==
  syntax.WrapperNullable` and `renderType` is a `*dst.ArrayType` with `Len ==
  nil`, e.g. immediately alongside the existing `specialSource` check at
  `gen_schema.go:1206-1213`:

  ```go
  if wrapper == syntax.WrapperNullable {
      if _, isSlice := renderType.(*dst.ArrayType); isSlice {
          return nil, fmt.Errorf("%s does not support arrays/slices at %s", wrapper, f.Position())
      }
      ...
  }
  ```

  Add a `nullable_byte_slice` (or extend `nullable_slice`) negative fixture
  covering `Nullable[[]byte]` specifically, since the existing `[]int` fixture
  does not exercise this code path.

### 2. nit — negative-proof coverage still omits Nullable + explicit-ref and Nullable + provider

- Where: `examples/optionality/negative/` (ten directories, listed in
  `cmd/proof/main.go:111-124`).
- The DoD's Nullable-rejection list (DoD :97-99) names six shapes: "arrays/
  slices, consts, enums, registered interfaces, explicit refs, providers, and
  templates" (seven, counting templates separately). The current negative
  fixtures cover enums, interfaces, and slices, but not explicit refs or
  providers/templates, even though the generic `specialSource != ""` guard at
  `gen_schema.go:1207-1209` already covers all of them mechanically (I verified
  this functions correctly with a real probe: `jsonschema.Nullable[string]`
  tagged `jsonschema:"ref=definitions/Foo"` fails with `"jsonschema.Nullable
  does not support explicit refs"`, exit 1). This is a coverage gap in the
  proof, not a functional bug — but the prompt's "whether the proof actually
  exercises the public behavior it claims" axis calls it out: the proof claims
  (README/DoD) that explicit refs and providers fail under Nullable, but nothing
  in the committed, executable proof demonstrates it.
- Smallest fix: add `nullable_ref` and `nullable_provider` fixtures parallel to
  the existing `nullable_enum`/`nullable_interface`/`nullable_slice` pattern and
  wire them into `cmd/proof/main.go`'s negative list.

### 3. nit — duplicated literal error string for the Nullable-registered-interface guard

- Where: `internal/builder/gen_schema.go:1343` and `gen_schema.go:1379` both
  hard-code the identical format string `"%s does not support registered
  interfaces at %s"` in two separate branches of `resolveLocalInterfaceProps`
  (the v1-config branch and the legacy `findInterfaceImpl` branch). This is
  defensible duplication (the two branches operate on genuinely different data
  and both need the guard independently — not over-engineering), but a shared
  one-line helper (e.g. `nullableInterfaceErr(wrapper, prop)`) would remove the
  risk of the two messages drifting if one is edited later without the other.
  Not blocking.

## Axis summary (findings only, no approval)

- **Correctness**: one new demonstrable bug (Finding 1) — `Nullable[[]byte]`
  passes generation when the DoD and this round's own narrowed documentation
  both require it to fail. All six previously-reported bugs/design gaps are
  genuinely fixed and re-verified against the current worktree, including with
  fresh real-generator probes, not just re-reading the diff.
- **Over-engineering**: none found. The new code (nullable schema transform,
  wrapper classification, template glue, byte/uint8 special case) stays
  proportional to the feature; Finding 3 is a minor duplication note that cuts
  toward under- rather than over-engineering.
- **Completeness vs. contract**: Finding 1 is a real completeness gap (a
  documented-unsupported shape that actually succeeds). Finding 2 is a proof-
  completeness gap only (functionality is correct; the executable proof doesn't
  demonstrate two of the seven Nullable-rejected shapes named in the DoD).
- **Factoring**: boundaries remain appropriate (recognition in `internal/syntax`,
  requiredness via `Required()`/`HasJSONOption`, nullability as a schema
  transform in `nullableSchema`, decode glue in the template). Finding 3 is the
  only factoring note and is not blocking.

## Sequencing note (per prompt instruction, not a finding)

The branch remains uncommitted (`git log origin/main..HEAD` is empty; all
product changes are staged/untracked modifications in the worktree). This is
the same state the first review documented and the prompt says not to re-report
as a new finding — I confirmed no product files listed in the DoD are actually
missing from the worktree (runtime, syntax, builder, template, examples,
fixtures, and docs changes are all present and functioning as described above).
