# Issue 28 optionality decision brief

Status: design investigation; do not implement until the user chooses a
direction.

Primary issue: [GitHub issue #28](https://github.com/tylergannon/go-gen-jsonschema/issues/28)

Related prerequisite bug:
[GitHub issue #29](https://github.com/tylergannon/go-gen-jsonschema/issues/29)

Full implementation plan:
[issue-28-optional-plan.md](issue-28-optional-plan.md)

## Assignment for Fable

Read this entire brief and the linked implementation plan. Independently
inspect the current repository, especially the syntax traversal, schema IR,
schema marshalers, enum and ref handling, and generated interface unmarshaler.
Investigate current primary sources outside the repository, including the
OpenAI Structured Outputs rules and the JSON Schema semantics relevant to
required properties, null, type arrays, `anyOf`, refs, enums, and consts.

Challenge the framing rather than merely choosing from the proposed answers.
Then write a self-contained prescription to:

```text
docs/design/issue-28-fable-review.md
```

The review must distinguish verified facts from inferences, cite external
sources, identify hidden implementation risks, and recommend the public API,
schema encodings, implementation order, and proof gates. Do not implement the
feature or edit the plan.

## The decision

The repository needs an explicit Go representation for a JSON property whose
zero value is meaningful. Issue #28 originally proposed only an omittable
`Optional[T]`. Current OpenAI Structured Outputs guidance makes a second
representation important: a required property whose value may be JSON `null`.

We need to decide whether to ship:

1. only omittable `Optional[T]`;
2. only required-but-nullable `Nullable[T]`; or
3. both, allowing authors to choose the actual wire contract.

The working hypothesis is option 3. Cost is not the primary discriminator.
Incorrect schemas, rejected API calls, retry loops, silently conflated states,
and hand-written decoding machinery are more expensive than emitting a modest
number of explicit `null` values.

## The three semantic cases

These cases are not interchangeable in JSON Schema:

| Go field | Property required | Allowed JSON values | No-value wire form | OpenAI strict potential |
|---|---:|---|---|---|
| `T` | yes | `T` | none | yes |
| `Optional[T]` | no | `T` | property absent | no |
| `Nullable[T]` | yes | `T` or `null` | explicit `null` | yes |

JSON Schema's `required` keyword controls key presence. JSON `null` is a value
and is not equivalent to an absent property. That distinction should remain
visible in both the public Go API and the generator IR.

Potential fourth semantics—absent, explicit null, or concrete value—is a true
three-state field. It is not required by issue #28 and should not be created by
accidentally nesting the proposed wrappers. If needed later, it deserves an
explicit contract and name.

## External facts

### JSON Schema

The official JSON Schema documentation establishes:

- Properties are optional unless named by `required`, and a required property
  with `null` is different from a missing property:
  <https://json-schema.org/understanding-json-schema/reference/object#required-properties>
- `null` is its own JSON type, not absence:
  <https://json-schema.org/understanding-json-schema/reference/null>
- The `type` keyword may be an array of basic JSON type names. The instance is
  valid if it matches any listed type. This syntax is not limited to Go-like
  primitives; JSON Schema's basic types include object and array:
  <https://json-schema.org/understanding-json-schema/reference/type>
- `anyOf` accepts an instance matching one or more subschemas:
  <https://json-schema.org/understanding-json-schema/reference/combining#anyof>

### OpenAI Structured Outputs

The current OpenAI guide establishes:

- Every field or function parameter must appear in `required` for strict
  Structured Outputs.
- OpenAI explicitly recommends a union with `null` to emulate an optional
  value and shows the compact form `"type": ["string", "null"]`.
- Supported schema forms include object, array, enum, and nested `anyOf`.
- The root must be an object and must not itself be `anyOf`.
- Every object must set `additionalProperties: false`.
- Every nested `anyOf` branch must itself comply with the supported subset.
- The guide demonstrates nullable recursion using `anyOf` with a `$ref` branch
  and a `{ "type": "null" }` branch.

Source:
<https://developers.openai.com/api/docs/guides/structured-outputs#supported-schemas>

These facts mean omittable `Optional[T]` cannot participate in an OpenAI-strict
object, while required `Nullable[T]` can, subject to the rest of the schema
also satisfying OpenAI's supported subset.

## Issue 28's original proposal

Issue #28 proposed:

```go
type Optional[T ~int | ~string | ~bool] struct {
	Present bool
	Value   T
}
```

The wrapper would use `IsZero` plus `json:",omitzero"`, render the inner schema
for `T`, and exclude the property from `required`. Its motivation is sound:
plain `T` cannot distinguish missing from present-zero, and a struct tag keeps
optionality outside the type system.

Decisions already made during review:

- Delete the legacy `jsonschema:"optional"` behavior instead of maintaining a
  compatibility branch.
- Do not impose the original primitive constraint on `Optional`; presence is
  useful for structs, slices, arrays, pointers, named types, refs, and other
  values the generator already supports. Use `Optional[T any]` and delegate
  inner-type support to the ordinary renderer.
- Include every integer variant and both float widths wherever the ordinary
  renderer supports them. The current renderer should also close its existing
  `byte`, `rune`, and `uintptr` parity gaps.
- Explicitly reject JSON `null` for `Optional`, because its generated schema
  permits `T`, not null.

## Proposed public runtime contracts

### Omittable value

```go
type Optional[T any] struct {
	Present bool
	Value   T
}
```

Contract:

- zero value means absent;
- `IsZero()` returns `!Present` so `json:",omitzero"` omits it;
- present values marshal exactly as `T`;
- input `null` fails;
- decode into a temporary and mutate only after success;
- a present value that marshals to `null` should fail, preserving the
  invariant that `Present` implies a concrete schema-valid `T` value;
- without `omitzero`, a containing struct cannot represent absence correctly.

### Required nullable value

```go
type Nullable[T any] struct {
	Present bool
	Value   T
}
```

Contract:

- zero value means the value is null;
- `Present == false` marshals as JSON `null`;
- non-null input decodes into a temporary, assigns, and sets `Present`;
- input `null` clears `Value` and `Present` without error;
- a present value that itself marshals to `null` should fail rather than
  collapse the two states;
- `IsZero()` should return false so an accidental `json:",omitzero"` cannot
  remove a schema-required property;
- plain `encoding/json` cannot distinguish a missing key from a null key after
  decoding. Schema validation supplies that guarantee for strict output.

`Nullable` is preferred over `StrictOptional` as a name because the wrapper
does not make its containing object strict. It only permits null while the
property remains required. A generator-level strict-compatibility check may be
useful separately.

## Schema encodings for Nullable

Correctness comes before compactness. The generator should choose encoding
from the rendered inner schema, not from an arbitrary Go primitive constraint.

### Direct typed schemas

For a simple node with one direct `type`, use the compact type array:

```json
{"type":["integer","null"]}
```

JSON Schema permits this for every basic type, including object and array.
Whether the first implementation uses the compact form beyond uncomplicated
scalar nodes is an implementation choice, not a semantic limitation.

### Refs and composed schemas

For `$ref`, an existing union, const, or another schema that cannot be made
nullable by changing one direct `type`, use `anyOf`:

```json
{
  "anyOf": [
    {"$ref":"#/$defs/Policy"},
    {"type":"null"}
  ]
}
```

For an existing `anyOf`, prefer appending or otherwise flattening the null
branch rather than producing unnecessary nested unions, provided semantics and
property order remain stable.

### Enums and consts

`enum` and `const` constrain values independently of `type`. Merely changing
`"type":"string"` to `"type":["string","null"]` does not necessarily make
null valid under a standards-compliant validator when an enum or const remains.
The implementation must either include null in the value constraint or use an
`anyOf` branch. This requires explicit local-validator and OpenAI proof.

## Current codebase anatomy

The fresh checkout currently has these relevant seams:

- `internal/syntax/scan_result.go:resolveTypeExpr` handles identifiers,
  pointers, arrays/slices, structs, and selectors, but not generic
  `*dst.IndexExpr` fields.
- Decorated one-argument generics were previously spiked and observed as a
  canonical-path `*dst.Ident` base plus the unchanged inner expression in
  `IndexExpr.Index`. That makes exact wrapper recognition and lossless
  normalization feasible.
- `internal/builder/model.go:ObjectProp` stores one `Optional bool`. The new
  design should represent property presence separately from value nullability.
- `PropertyNode.Typ`, `ArrayNode.Type()`, and `ObjectNode.Type()` model direct
  types. `RefNode` and `UnionTypeNode` do not.
- `renderStructField` handles refs and field-configured enums and interfaces
  before its ordinary `renderSchema` fallback. Every one of those paths needs
  the same normalized wrapper metadata.
- Field-specific enum detection currently requires the raw field type to be a
  direct `*dst.Ident`.
- Registered-interface discovery also requires a direct `*dst.Ident`.
- Generated interface unmarshaling shadows the interface field with
  `json.RawMessage` and assigns directly to the field. Both wrappers require
  explicit generated assignment and presence/null transitions.
- The public manual schema API has a single-string `DataType`, plus separate
  `UnionSchemaEl`/`JSONUnionType` support. Avoid an unnecessary breaking change
  to that API while implementing generated-schema nullable nodes.

## Previously validated AST and IR findings

A throwaway spike before the fresh re-clone exercised scalar, named struct,
slice, pointer, inline struct, map, registered interface, and ref arguments.
It established:

- each one-argument wrapper was a canonical-path `*dst.IndexExpr`;
- the inner node retained its ordinary AST shape;
- manually rendering the inner node produced the existing scalar, object,
  array, pointer-object, inline-object, registered-union, and ref-target IR;
- maps retained the ordinary `mapType/chanType not allowed` error;
- registered interfaces were the exceptional generated-code path because
  interface discovery rejected the wrapper as a nested unsupported location.

The same spike exposed the exported-field traversal bug now preserved in issue
#29. That bug is a prerequisite because it masks generic traversal entirely.

## Candidate internal architecture

Use one exact syntax classifier for direct fields:

```go
type FieldValueMode uint8

const (
	RequiredValue FieldValueMode = iota
	OmittableValue
	NullableValue
)

type FieldValueSemantics struct {
	Inner dst.Expr
	Mode  FieldValueMode
}
```

The classifier should recognize only canonical library types:

```text
github.com/tylergannon/go-gen-jsonschema.Optional[T]
github.com/tylergannon/go-gen-jsonschema.Nullable[T]
```

Every scanner and builder path consumes the same normalized inner expression
and mode. Do not duplicate generic matching across traversal, rendering, enum,
ref, provider, or interface logic.

Requiredness and nullability then remain orthogonal:

- required ordinary: required property, ordinary inner schema;
- omittable: not required, ordinary inner schema;
- nullable: required, nullable inner schema.

Add a dedicated null schema node and a nullable-schema transformation that:

- emits a compact type array for uncomplicated directly typed nodes;
- emits or flattens `anyOf` for refs, unions, consts, and other composed nodes;
- preserves comments, property order, discriminators, refs, enums, and type
  identities;
- never degrades an unsupported inner type to `{}`.

## Generated interface unmarshaling

For a registered interface field, generated code must distinguish wrapper
mode using the shadowing `json.RawMessage`:

- ordinary interface: retain existing behavior;
- `Optional[Interface]`: nil raw bytes mean absent; non-nil bytes must decode a
  concrete implementation; null fails; assign `.Value` and then `.Present`;
- `Nullable[Interface]`: nil raw bytes mean missing input, which is invalid at
  the schema layer; raw `null` means `Present == false`; other bytes decode,
  assign `.Value`, then set `.Present`.

Failures must not partially mutate the destination. Cover both value and
pointer implementations and both interface registration paths that remain
supported.

## Correctness versus output cost

Nullable strict output can emit more tokens because every nullable property
name and a value—often `null`—must appear. That is real but should not drive
the API design without workload measurements. A modest deterministic token
increase is normally dominated by the cost of malformed output, rejected
strict schemas, retries, manual decoding, or silently wrong default behavior.

The durable design criterion should be semantic correctness:

- choose `Optional` when absence itself is meaningful in the wire protocol;
- choose `Nullable` when the key must always exist and null represents no
  value, especially for strict Structured Outputs;
- choose ordinary `T` when a concrete value is always required.

## Fresh-clone baseline debt

On 2026-07-10, a fresh clone failed `go test ./...` because ignored
`jsonschema_gen.go` files were absent from `examples/structs` and
`examples/providers_rendering`. Targeted generation restored a green test
baseline. Repository-wide `go generate ./...` also failed because a syntax
test fixture invokes the removed `-type` CLI flag.

This is inherited debt, not caused by the design docs, but issue #28 closeout
cannot honestly claim fresh-clone or repository-wide generation proof until it
is fixed.

## Questions Fable must answer

1. Should the project ship both `Optional[T any]` and `Nullable[T any]`, or is
   one semantic a better public default?
2. Are these names and `{Present, Value}` fields the clearest API?
3. Should nullable rendering use compact type arrays for all direct typed
   nodes or only simple scalar nodes?
4. Where is `anyOf` required or preferable for refs, objects, arrays, enums,
   consts, and existing interface unions?
5. Is a presence-mode enum the right IR boundary, or is there a simpler design
   that does not spread wrapper knowledge through the codebase?
6. What runtime null and partial-mutation invariants should each wrapper
   enforce?
7. What compatibility checker or generator option, if any, should assert the
   whole schema satisfies OpenAI's strict subset?
8. What should be implemented first, and which parts should remain separate
   review boundaries?
9. What important failure modes or alternatives has this brief missed?
10. What concrete tests prove the design against the local validator and the
    live or documented OpenAI behavior?

