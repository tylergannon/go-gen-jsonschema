# Source review for the accepted type-definition grammar

Baseline `604fabb`. `go test ./...` passed before exploration in this worktree. This report is read-only product research: no product files or GitHub state were changed. Root owns grammar implementation and final verification.

## Authority and scope

`docs/spec/v1.md:3-7` is now the **accepted** release contract, while implementation/conformance remain prerequisites. It supersedes the old draft. In particular, providers are no longer an unresolved decision: the current provider API is retained for 1.0 (`docs/spec/v1.md:111-115`), with schema-only limits (`:42`). The grammar should describe statically resolved accepted type definitions; it must not imply that all retained provider schemas have statically derivable Go/TS wire types.

Keep this step to the Go algebra, validation, and its tests. It does not need a scanner adapter, emitter, runtime validator, or the #57/#58 codecs to be complete.

## Precise supported forms

| Form | Accepted intent and source evidence |
| --- | --- |
| Scalars | Named/basic strings, booleans, signed/unsigned Go integers, and floats (`v1.md:30`). Preserve exact Go kind/width; current `renderSchema` collapses all integer widths to integer and both float widths to number (`internal/builder/gen_schema.go:812-819`). `time.Time` is a distinct known leaf with string wire shape, not arbitrary external struct inference (`:821-833`). |
| Ordinary pointers | Required/non-null valid values (`v1.md:22,31`). Pointer-to-scalar and pointer-to-slice are not excluded. Existing fixtures explicitly include `*int`, `*NamedInt`, `*RemoteSlice`, and `[]*RemoteSlice` (`internal/builder/testfixtures/indirecttypes/types.go:13-16,48-52`). Pointers are ordinary type composition, not a synonym for optional or nullable. |
| Ordinary arrays/slices | Supported non-null collections and nested ordinary composition; preserve fixed array length separately from slice identity. Current schema mapping discards the array length (`internal/builder/gen_schema.go:878-889`). Named chains of ordinary slices and cross-package references are demonstrated (`internal/builder/testfixtures/indirecttypes/types.go:18-52`). Byte-like slices whose Go JSON representation is base64 are excluded from portable support (`v1.md:46`). |
| Objects | Ordered resolved JSON properties, including supported embedding/field selection (`v1.md:31,74`). Anonymous ordinary structs can exist; a registered union belongs to a supported named owner, not arbitrary inline ownership (`:74`). |
| Named definitions/aliases | Ordinary named chains are followed by discovery and rendering (`internal/syntax/scan_result.go:539-558`; `internal/builder/gen_schema.go:836-875`). A normalized Ref to a definition is appropriate; distinguish this from a raw JSON Schema URI. Alias expansion must not erase source restrictions on wrappers or registered-interface positions. |
| Optional | Direct named **field** wrapper only, requires `json:",omitzero"`, inner supported scalar/struct/pointer/ordinary array/slice/ref (`v1.md:37,49`; `internal/builder/gen_schema.go:1262-1272`). Named here excludes an embedded anonymous field; it does not mean every ordinary containing struct must itself be named. Missing versus present-zero is semantic. |
| Nullable | Required field, allowed for scalars, enums, objects, pointers to objects, and AsRef objects (`v1.md:38`). No Optional+Nullable combined state. Arrays/slices, registered interfaces, providers, and arbitrary explicit refs are rejected by current Nullable handling (`internal/builder/gen_schema.go:1398-1444`). |
| Union direct `I` | Field-local configured tagged object union; each implementation resolves to a named object type, with pointer/value registration retained (`v1.md:57,61,69-74`; `internal/builder/gen_schema.go:1499-1526,1661-1678`). |
| Union `Optional[I]` | Same case mapping with property absence allowed. Present nil remains invalid (`v1.md:69,72`). |
| Union direct `[]I` | Required non-null one-dimensional slice; tagged object members; allocated empty slice allowed (`v1.md:69,72`). No Optional slice-of-union form. |
| Enums | String enum, numeric integer mode, or field-specific constant-name string mode (`v1.md:32-33,91-97`). Exact underlying values must be preserved; `String()` output is not the wire identity. Field-local adaptation cannot be owned by a single global Go type definition. |
| Refs | Nonrecursive named AsRef definitions and supplied explicit schema targets exist (`v1.md:41`). For this closed resolved grammar, Ref should resolve to a known Definition. Unresolved external schema URIs and provider holes are outside the grammar until their shape is supplied; do not encode them as `any`. |

The accepted matrix explicitly is **not** a Cartesian-product guarantee (`v1.md:53`). In particular, registering an integer enum's string mode does not prove conversion inside every pointer/collection combination. Current field enum handling only sees a direct identifier after unwrapping Optional/Nullable (`internal/builder/gen_schema.go:1285-1358`). String-mode adapter contexts beyond those must not become portable capabilities merely because an Enum node can be placed beneath a Slice node.

This must **not** become a blanket restriction on enum composition. The existing registered enum fixture includes `[]EnumType`, `[]RemoteEnumType`, and `[]*RemoteEnumType` (`internal/builder/testfixtures/enums/types.go:24-31`), registers all three generation roots, and registers the ordinary underlying-string enum (`internal/builder/testfixtures/enums/schema.go:27-33`). Preserve those enum constraints; lowering them to Scalar would weaken existing behavior. A mode-aware rule can allow `EnumValues` in ordinary compositions while restricting the additional `EnumNames` conversion to established direct field/wrapper contexts.

## Composition restrictions that should be structural

Put the registered union in the **field-value sum**, not the general Type sum. A viable shape is:

```text
Definition = identity(package path, Go name) + Type + source
Type       = Scalar | Enum | Object | Pointer(Type) | Slice(Type)
             | Array(length, Type) | Ref(Definition identity) | Time
FieldValue = Required(Type) | Optional(Type) | Nullable(Type)
             | Union(config) | OptionalUnion(config) | UnionSlice(config)
```

This makes naked union roots, `Pointer(Union)`, `Slice(Slice(Union))`, and a named alias whose body is Union unrepresentable. It still admits `Slice(Ref(Owner))` where Owner has its own valid union field: nesting a supported named owner is not the same as nesting an unowned interface.

Wrapper forms likewise belong on fields rather than Type. That makes `Optional[Nullable[T]]`, aliases/definitions of wrappers, wrappers as roots, and wrappers inside ordinary containers unrepresentable. An Optional object can still contain its own independent Optional properties; that is not nested wrappers. The scanner enforces direct wrapper positioning and rejects generic wrappers encountered elsewhere (`internal/syntax/scan_result.go:492-537`).

The exact closed union field forms are `I`, `Optional[I]`, and direct `[]I`; the contract excludes `[N]I`, `[][]I`, named interface containers, `*I`, `[]*I`, parenthesized interface forms, `Optional[[]I]`, and every Nullable interface form (`v1.md:48,69-74`). Current resolver distinguishes a syntactically direct identifier/slice identifier (`internal/builder/gen_schema.go:1459-1471,1545-1642`); do not normalize a rejected source form into an accepted form before validating it. Ordinary alias names can normalize; an alias-to-interface is not demonstrated as a supported registration path and current v1 resolution rejects a non-interface AST declaration at `:1586-1591`.

## Recommended validator invariants

1. Definition identities are nonempty/unique; Go identifiers are valid; every Ref resolves. Detect cycles across definitions and through actual pointer-linked in-memory nodes, including pointer and collection edges. Repeated acyclic DAG reuse is allowed; use visiting/visited states, not a global “seen once” prohibition.
2. Every field has exactly one field-value constructor, unique resolved JSON property name, valid source/name metadata, and a non-nil child. Union fields have a named owner. Object field order is stable and embedding has already been resolved by the future adapter.
3. Required/Optional child types are otherwise accepted ordinary types. Nullable resolves Ref chains for classification and accepts only its declared scalar/enum/object/Time or pointer-to-object subset; it must not accidentally admit pointer-to-scalar/array/slice just because the existing schema renderer erases pointers.
4. Scalar kinds are enumerated, numeric kind/width combinations are valid, fixed array lengths are nonnegative, and a Slice whose resolved element has byte-like base64 encoding is outside this closed portable domain. A plain `uint8` scalar remains valid; pointer elements are not equivalent to a byte slice. Exact treatment of custom byte hooks remains an unresolved/custom mapping, not guesswork.
5. Enums are nonempty, use a supported homogeneous underlying string/integer kind, preserve values with `go/constant` or an equivalent exact representation, and validate value kind/range. If field-local wire mapping is represented now, keep numeric identity separate from string-mode names; reject ambiguous duplicate underlying values for that string-mode occurrence per `v1.md:95`. Do not make one global Enum definition choose a mode for every field.
6. Union cases are nonempty; discriminator keys/values are strings with exact contents retained; tags and implementation registrations are unique. Do not reject an empty string tag merely by assuming it is invalid: accepted v1 requires string tags but does not specify nonempty tags. Default-key/name derivation belongs before the resolved grammar. Each implementation identity resolves to an object definition; pointer/value form is explicit and part of registration identity. A single-case union still keeps its tag.
7. Resolve any payload property sharing the discriminator key deliberately. Treat the tag as a contextual refinement without mutating the reused payload definition: an unconstrained string property or string enum containing the tag is compatible, not just a singleton enum. The union's effective property is required/non-null exact tag; impossible constraints such as a numeric property or a string enum without the tag are static conflicts. Optional/Nullable string fields can likewise be narrowed by the union's required tag. Time/custom codecs require care beyond their superficial string shape. A runtime custom object's unknown output cannot be certified by static grammar validation; the accepted custom-hook rules remain codec/conformance obligations (`v1.md:71,80-89`).
8. Providers, raw external refs, arbitrary custom wire mappings, maps, chans, funcs, inline interfaces, recursive types, and unsupported source wrapper/interface forms cannot masquerade as accepted static nodes. This is a boundary of this grammar, not removal of the retained provider product API.

## Intended contract versus current implementation

- The current generator still lacks the union encoder and explicitly tests loss of the discriminator on re-marshaling (`examples/sealed_interface_slices/schema_test.go:66-71`). #57 owns that difference. Grammar validation need not wait for or implement the codec.
- The accepted contract requires generated-owner replacement on successful decode (`v1.md:63`), but the existing direct slice path preserves an old slice when missing (`internal/builder/schemas.go.tmpl:184-185`). This is an implementation gap, not a reason to add a fourth union shape or a patch/merge mode to the type grammar.
- Duplicate legacy type names must be rejected (`v1.md:70`); current codec preparation silently suffixes duplicates (`internal/builder/gen_schema.go:1092-1100`) while schema objects retain bare names (`:1674-1678`). A resolved Union must have one definitive mapping.
- Nullable's stated pointer support is pointer-to-object (`v1.md:38`), but current general pointer erasure followed by scalar Nullable handling can admit more (`internal/builder/gen_schema.go:874-875,1419-1434`). Treat the accepted subset as authoritative; do not adopt accidental broader behavior.
- Integer widths, array lengths, pointer shape, field-local codec ownership, and raw named identities are absent or incomplete in the existing schema model (`internal/builder/model.go:31-105`). The new grammar should preserve them rather than alias the old schema algebra.
- Enum discovery/rendering contains literal trimming and index-based numeric assumptions (`internal/builder/gen_schema.go:668-686,705-724,1323-1355`). These are not authoritative semantics for the new grammar; exact typed constant values are the correct boundary. Fixing the scanner is outside this step.
- The renderer has an unresolved-external fallback to an empty ObjectNode (`internal/builder/gen_schema.go:843-851`), despite the accepted contract's diagnostic requirement (`v1.md:51`). That fallback must not define an accepted “unknown” grammar node. Its comment says permissive schema, but ObjectNode actually emits a closed object (`internal/builder/model.go:283`), another reason to avoid inheriting it.
- The helper `syntax.TypeID.Concrete()` currently mutates `t.Indirection` but returns the unchanged copy (`internal/syntax/type_id.go:35-38`). New identities should not reuse that behavior as a normalization primitive. This report does not change it.

These discrepancies justify an explicit accepted grammar with tests, not a concurrent rewrite of rendering, scanning, or codecs.

## Follow-up clarifications sent during API design

- Explicit empty `Impl("")` wire values are accepted by `strconv.Unquote` and registration validation (`internal/syntax/scan_expr.go:397-417`; `internal/builder/gen_schema.go:146-172`). The accepted contract does not impose a nonempty-tag condition. Preserve exact whitespace and empty tags unless the contract is deliberately amended. Valid UTF-8 follows the lossless-domain restriction.
- An empty configured discriminator **key** currently means default `type`, not a literal empty JSON property (`internal/builder/model.go:478-480`). A resolved grammar should receive the effective key after defaulting; document that normalization boundary rather than claim raw empty keys were forbidden registration syntax.
- Duplicate exact implementation registrations with different tags are rejected (`internal/builder/gen_schema.go:162-164`). `T` and `*T` have distinct implementation identities; both may be explicitly registered with distinct tags if they satisfy the interface. Under legacy derived tags their shared bare type name is a collision and must be rejected under `v1.md:70`.
- `prependDiscriminator` currently **prepends and copies** a colliding payload property (`internal/builder/model.go:482-493`); it does not replace it. Do not derive the new grammar's collision contract from that duplicate-property behavior.
