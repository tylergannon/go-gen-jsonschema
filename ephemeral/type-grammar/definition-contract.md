# Authoritative resolved definition grammar

The Go declarations in `internal/typegrammar/grammar.go` and the admission rules in `internal/typegrammar/validate.go` are the executable authority for the static grammar that subsequent backend work will consume. `docs/spec/v1.md` remains the product contract. This step does not adapt the existing scanner/builder or change generation output.

## Constructors

```text
Definitions = ordered named definitions
Type        = Scalar | Time | Enum | Object
            | Pointer(Type) | Slice(Type) | Array(length, Type) | Ref(Name)
FieldValue  = Required(Type) | Optional(Type) | Nullable(Type)
            | Union(config) | OptionalUnion(config) | UnionSlice(config)
```

Only pointers to the declared Go node structs implement these closed interfaces. Validation rejects nil/unknown implementations and every invalid composition; an embedded marker method cannot extend the accepted language. The graph may share children, but cannot contain a back edge through any child, reference, or union implementation. Definitions and properties retain their supplied order.

Union and wrapper constructors are deliberately outside `Type`. A union cannot occur as a standalone definition, inside an arbitrary collection, behind a pointer, or behind a named alias. Its supported forms are direct I, Optional[I], and direct []I, on a named object owner. A collection of named owners with their own union fields is a different, valid composition. Each union occurrence owns its registration: interface identity alone cannot be used as a cache key for wire mapping.

Required values are non-null; Optional adds absence; Nullable adds required null. Nullable accepts scalar/enum/object/time leaves and a pointer to an object, including resolved references, rather than all ordinary type constructors. Pointers preserve Go indirection without implying null. Integer widths/signedness and fixed array lengths remain in the model, without inventing schema constraints or JavaScript numeric precision guarantees.

EnumValues uses exact underlying string/integer constants and can compose under ordinary pointers and collections. EnumNames uses integer constant names as wire strings and requires a direct field of a named object owner; it is not a blanket transport promise for container adapters. The same Go enum type may have different resolved registrations in different fields. Duplicate underlying values are permitted in value mode, but ambiguous in name mode.

Empty explicit union tags are valid; tag/key whitespace and escaping are preserved. The effective discriminator key is supplied after source defaults. Value and pointer implementations are distinct identities; exact duplicate registrations or duplicate tags are rejected. A shared payload's compatible discriminator property is contextually narrowed to a required exact string, without mutation. A payload constraint that cannot admit the tag is rejected.

## Boundary with source lowering

The forthcoming adapter must validate facts that a normalized graph does not retain: package/source eligibility, Go interface satisfaction, syntactically direct union/wrapper forms, wrapper aliases, field promotion/embedding and JSON selection, Optional's omitzero tag, handwritten codec ownership, and custom mappings. It must resolve names and registration defaults before constructing the graph. It must diagnose providers and unresolved external schema shapes rather than invent an opaque node. The retained provider API is not removed by this grammar.

`Ref` is a named-definition edge, not JSON Schema's output `$ref` policy. AsRef and artifact options belong to the adapter/backend configuration. Ordinary byte-like slices, maps, channels, functions, arbitrary schema fragments and unproved custom hook mappings are not members of this static portable grammar.

## Inductive use

Validation first establishes a finite DAG. A backend can then define a projection for every constructor: scalar/time/enum bases; object properties and field-presence constructors; pointer/array/slice composition; resolved references; and field-specific tagged object variants. Given correct projections of all children, each constructor must preserve its documented structural meaning. The closed constructor set plus a rejecting default prevents silent fallback to any/unknown. Acyclic references allow topological induction instead of a fixed-point argument.

This is a framework for an explicit correctness argument, not a machine-checked proof or a claim that finite tests prove the emitter. There is no emitter in this step. TypeScript literal representability and compiler evidence come next; definitive bidirectional runtime transport remains #71, depending on #57.

## Next step

Implement a checked adapter from current source/registration data into these types, prove it against actual accepted and rejected Go fixtures, then build the TypeScript projection and printer on the validated grammar. Keep the adapter's source-admissibility diagnostics distinct from target-language projection limits. Do not repair or infer codec semantics in that adapter.
