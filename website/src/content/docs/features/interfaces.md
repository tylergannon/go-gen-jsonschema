---
title: Sealed interfaces and discriminators
description: Generate discriminated unions for sealed interface fields.
---

A field whose type is a **sealed interface** becomes an `anyOf` union of the
interface's variants, discriminated by a `"type"` property whose value is the
concrete type name. An interface is sealed when its own body declares an
unexported method; its variants are inferred: every named struct type in the
same package that declares that method directly. The receiver of the sealing
method decides the variant kind: a value receiver is a value variant, a
pointer receiver is a pointer variant, and decoding constructs the variant
accordingly. Nothing is declared at the field. A direct one-dimensional slice
of the interface becomes an array with the union under `items.anyOf`.

The generator writes `MarshalJSON` and `UnmarshalJSON` methods on the
containing struct.

```go
// PaymentMethod is sealed by its unexported method.
type PaymentMethod interface{ isPaymentMethod() }

type Card struct {
    Number string `json:"number"`
}
func (Card) isPaymentMethod() {}          // value variant, wire value "Card"

type BankTransfer struct {
    Account string `json:"account"`
}
func (*BankTransfer) isPaymentMethod() {} // pointer variant, wire value "BankTransfer"

type Payment struct {
    Methods []PaymentMethod `json:"methods"`
}

// schema.go (//go:build jsonschema)
var _ = polytype.Declare(Payment.Schema)
```

## What is rejected

Each of these fails generation with a diagnostic naming the type or field:

- a reachable interface field whose interface declares no unexported method
  of its own. Non-sealed unions are unsupported; a wire contract needs a
  closed membership visible in one place, and there is no explicit fallback;
- an interface whose unexported method arrives only by embedding another
  interface;
- a type that satisfies the interface only through an embedded field: it
  inherits the sealing method and is excluded, with a diagnostic distinct
  from an invalid direct candidate;
- a type that declares the sealing method directly but does not implement the
  complete interface, or is not a struct;
- a sealed interface with zero variants;
- an embedded interface payload;
- a variant whose payload property collides with the discriminator property.

Variants behind other build tags are not discovered: the scanner loads the
package with the `jsonschema` build tag.

## Custom discriminator

The discriminator property is a property of the union, never of a field. A
union has one codec, so it has one discriminator. The default `"type"` needs
no declaration; to use another property, declare it once, in the build-tagged
file of the package that declares the interface:

```go
//go:build jsonschema

var _ = polytype.SealedUnion[PaymentMethod]("kind")
```

Every use of `PaymentMethod` in every generated schema, codec, and TypeScript
output now discriminates on `"kind"`; nothing changes at any field, and the
values are still the concrete type names. The argument must be a string
literal naming a nonempty property. A declaration in another package, a
duplicate declaration, a declaration for a non-sealed interface or a
non-interface type, a non-literal argument, an invalid property name, or a
variant payload property colliding with the declared name is a generation
error naming the interface.

## Wire-contract hazard

The discriminator value is the concrete type name, so renaming a variant type
changes its wire value, and adding a qualifying implementation changes the
union's membership. Both show up as diffs in the generated schema and
TypeScript; review them as contract changes.

## Marshaling interface values

Marshal the containing value or its pointer. Its generated codec supplies the
discriminator:

```go
payment := Payment{Methods: []PaymentMethod{
    Card{Number: "example"},
    &BankTransfer{Account: "example"},
}}
data, err := json.Marshal(payment)
// {"methods":[{"type":"Card","number":"example"},
//             {"type":"BankTransfer","account":"example"}]}
```

Marshaling `Card` by itself uses normal Go encoding. A handwritten concrete
hook may return an object with a missing or matching discriminator;
conflicting/non-string discriminators, null, and non-object payloads are
errors.

Slice elements are decoded in order, and an invalid element reports its
zero-based index without partially assigning the destination slice. Value and
pointer variants remain value and pointer values after decoding. A successful
decode assigns the containing struct only after all interface fields and
slice elements decode successfully.

Nil required unions, typed-nil implementations, and nil required union slices
are rejected during encoding. An allocated empty slice emits `[]`; an absent
`Optional[I]` is omitted. Production owner JSON method collisions are rejected
before generation writes output. Use named fields when nesting owners that
each need generated codecs; embedding such an owner would promote competing
JSON methods and is rejected.

The slice must be the direct field type. Fixed arrays, nested slices, named
slice containers, `Optional[[]I]`, `Nullable[[]I]`, and inline interface
declarations are rejected. An `Optional[I]` scalar interface field is
supported. `Nullable[I]` is not.

## Migration

`Declare(T.Schema).Interface(field, ...)`, `WithInterface`,
`WithInterfaceImpls`, `WithDiscriminator`, `Impl`, `Discriminator`, and the
package-level `NewInterfaceImpl[I](...)` are removed. Seal the interface with
an unexported method, declare it directly on every variant, and delete the
field-level declaration. Wire values are the concrete type names; a custom
discriminator property moves to one `SealedUnion[I](name)` declaration in the
interface's package.

See the compiling [`examples/sealed_interface_slices`](https://github.com/tylergannon/polytype/tree/main/examples/sealed_interface_slices)
package for schema and runtime coverage, including value and pointer
variants.
