# TypeScript structural projection

This backend projects an admitted `typegrammar.Definitions` DAG to TypeScript
declarations. It describes the JSON value structure accepted by the source
grammar. It does not claim JSON Schema validation equivalence, runtime codec
equivalence, formal verification, or definitive bidirectional transport.

## Constructor projection

The projection `P` is defined for every admitted ordinary type constructor:

- `Scalar(Bool)` becomes `boolean`; `Scalar(String)` becomes `string`; every
  admitted integer and floating kind becomes `number`.
- `Time` becomes `string`.
- `EnumValues` becomes a union of its exact string or integer wire literals.
  An integer is emitted only when Go's constant package proves that its value is
  exactly representable as a JavaScript `number`; otherwise generation fails.
  `EnumNames` becomes a union of the exact Go constant names as string literals.
- A nonempty `Object` becomes a property-bearing object type in grammar order.
  An empty `Object` becomes `object`, because TypeScript `{}` also admits
  primitives. `object` excludes primitives and `null` without `any` or
  `unknown`; it remains a structural approximation rather than an exact JSON
  object validator.
- `Pointer(T)` becomes `P(T)` and does not introduce `null`.
- `Slice(T)` and `Array(length, T)` become `Array<P(T)>`. The grammar retains
  fixed length, but the accepted structural projection does not add a tuple
  constraint.
- `Ref(name)` becomes the allocated exported declaration name for `name`.

Direct field constructors project as follows:

- `Required(T)` is a required property of type `P(T)`.
- `Optional(T)` is an optional property of type `P(T)`.
- `Nullable(T)` is a required property of type `P(T) | null`.
- Each `Union` variant becomes
  `Omit<Implementation, discriminator> & { discriminator: exactTag }` and the
  field type is the union of those branches. This adds a required singleton tag
  when the payload lacks that property and narrows a compatible existing
  required, optional, or nullable property without changing the shared payload
  declaration. A one-variant union remains tagged rather than collapsing to the
  payload.
- `OptionalUnion` uses that same union as an optional property.
- `UnionSlice` uses `Array<that union>` as a required property.

All JSON property names and string literals are JSON-quoted, which is valid
TypeScript syntax. Documentation comments normalize line endings, replace
invalid UTF-8, escape control characters and line separators, and interrupt
`*/`. Exported type names are ASCII TypeScript identifiers. A source name is
encoded when needed; TypeScript keywords and emitted helpers are reserved; and
colliding base names receive an injective hex encoding of package path plus Go
name. Barrel exports sort the final allocated names.

## Constructor-complete induction

`Definitions.Validate` is the admission boundary. It proves that every node is
one of the constructors above, every field form is admitted in its documented
context, every reference and union implementation resolves, and all child,
reference, field, and union edges form a finite DAG.

Induct on a topological ordering of that DAG. The base cases `Scalar`, `Time`,
and `Enum` have direct projections above; exact-literal failure is an explicit
backend result rather than a fallback. For the inductive step, assume every
outgoing child edge of a node has a projection. `Pointer`, `Slice`, `Array`, and
`Ref` compose already-projected children or declarations. `Object` composes the
projection of each admitted field form, and each union field composes resolved
object declarations with an occurrence-local discriminator object. These cases
exhaust the sealed ordinary and field constructor sets. Therefore every
validated definition either has a finite TypeScript structural projection or
fails specifically because an enum integer cannot be represented exactly as a
TypeScript `number`; no case needs `any`, `unknown`, or an inferred codec rule.
