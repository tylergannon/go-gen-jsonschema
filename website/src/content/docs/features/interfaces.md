---
title: Interfaces and discriminators
description: Generate discriminated unions for interface-typed Go fields.
---

An interface field becomes an `anyOf` over its registered implementations. A
direct one-dimensional slice of that interface becomes an array with the union
under `items.anyOf`. The generator writes `MarshalJSON` and `UnmarshalJSON`
methods on the containing struct.

```go
type PaymentMethod interface{ isPaymentMethod() }

type Card struct {
    Number string `json:"number"`
}
func (Card) isPaymentMethod() {}

type BankTransfer struct {
    Account string `json:"account"`
}
func (*BankTransfer) isPaymentMethod() {}

type Payment struct {
    Methods []PaymentMethod `json:"methods"`
}
```

Register the field, its implementations, and optionally a custom discriminator:

```go
var _ = jsonschema.Declare(Payment.Schema).
    Interface(
        Payment{}.Methods,
        jsonschema.Discriminator("!kind"),
        jsonschema.Impl("card", Card{}),
        jsonschema.Impl("bank_transfer", (*BankTransfer)(nil)),
    )
```

`Impl` keeps each implementation next to its stable wire discriminator. The
default discriminator property is `type`. Migration: `NewJSONSchemaMethod(
Payment.Schema, WithInterface(Payment{}.Methods, Impl(...), ...))` is now
`Declare(Payment.Schema).Interface(Payment{}.Methods, Impl(...), ...)`. The
legacy split form using `WithInterfaceImpls` and `WithDiscriminator` remains
supported and source-compatible; without explicit `Impl` values, wire
discriminators derive from Go type names. Each
field's generated encoder supplies the discriminator expected by the decoder.
Slice elements are decoded in order, and an invalid element reports
its zero-based index without partially assigning the destination slice.

Value and pointer implementations remain value and pointer values after
decoding. A successful decode assigns the containing struct only after all
registered interface fields and slice elements decode successfully.

## Marshaling interface values

Marshal the containing value or its pointer. Its generated codec supplies the
field-specific discriminator:

```go
payment := Payment{Methods: []PaymentMethod{
    Card{Number: "example"},
    &BankTransfer{Account: "example"},
}}
data, err := json.Marshal(payment)
// {"methods":[{"!kind":"card","number":"example"},
//             {"!kind":"bank_transfer","account":"example"}]}
```

Marshaling `Card` by itself uses normal Go encoding; the same implementation
may have a different discriminator in another owner field. No global
implementation marshaler is required. A handwritten concrete hook may return
an object with a missing or matching discriminator; conflicting/non-string
discriminators, null, and non-object payloads are errors.

Nil required unions, typed-nil implementations, nil required union slices, and
unregistered dynamic types are rejected during encoding. An allocated empty
slice emits `[]`; an absent `Optional[I]` is omitted. Production owner JSON
method collisions are rejected before generation writes output. Use named
fields when nesting owners that each need generated codecs; embedding such an
owner would promote competing JSON methods and is rejected.

The slice must be the direct field type. Fixed arrays, nested slices, named
slice containers, `Optional[[]I]`, `Nullable[[]I]`, and inline interface
declarations are rejected.

An `Optional[I]` scalar interface field is supported. `Nullable[I]` is not.

See the compiling [`examples/sealed_interface_slices`](https://github.com/tylergannon/go-gen-jsonschema/tree/main/examples/sealed_interface_slices)
package for schema and runtime coverage, including value and pointer
implementations.
