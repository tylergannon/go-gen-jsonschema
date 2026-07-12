---
title: Interfaces and discriminators
description: Generate discriminated unions for interface-typed Go fields.
---

An interface field becomes an `anyOf` over its registered implementations. A
direct one-dimensional slice of that interface becomes an array with the union
under `items.anyOf`. The generator writes `UnmarshalJSON` dispatch code for the
containing struct.

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
var _ = jsonschema.NewJSONSchemaMethod(
    Payment.Schema,
    jsonschema.WithInterface(
        Payment{}.Methods,
        jsonschema.Discriminator("!kind"),
        jsonschema.Impl("card", Card{}),
        jsonschema.Impl("bank_transfer", (*BankTransfer)(nil)),
    ),
)
```

`Impl` keeps each implementation next to its stable wire discriminator. The
default discriminator property is `!type`. The compatible split form using
`WithInterfaceImpls` and `WithDiscriminator` remains supported; without
explicit `Impl` values, wire discriminators derive from Go type names. Each
implementation's JSON must carry the discriminator expected by the generated
unmarshaler. Slice elements are decoded in order, and an invalid element reports
its zero-based index without partially assigning the destination slice.

Value and pointer implementations remain value and pointer values after
decoding. A successful decode assigns the containing struct only after all
registered interface fields and slice elements decode successfully.

## Marshaling interface values

The generator writes `UnmarshalJSON` dispatch code. It does not add a
discriminator when concrete implementations are marshaled. If decoded values
must round-trip into schema-valid JSON, make each implementation emit the same
discriminator required by the schema:

```go
func (c Card) MarshalJSON() ([]byte, error) {
    type plain Card
    return json.Marshal(struct {
        Kind string `json:"!kind"`
        plain
    }{Kind: "card", plain: plain(c)})
}
```

Without this method, ordinary `json.Marshal` emits the concrete fields but not
`!kind`, so the result cannot be decoded by the generated union unmarshaler and
does not satisfy the generated union schema.

The slice must be the direct field type. Fixed arrays, nested slices, named
slice containers, `Optional[[]I]`, `Nullable[[]I]`, and inline interface
declarations are rejected.

An `Optional[I]` scalar interface field is supported. `Nullable[I]` is not.

See the compiling [`examples/sealed_interface_slices`](https://github.com/tylergannon/go-gen-jsonschema/tree/main/examples/sealed_interface_slices)
package for schema and runtime coverage, including value and pointer
implementations.
