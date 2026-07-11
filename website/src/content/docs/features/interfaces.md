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
func (BankTransfer) isPaymentMethod() {}

type Payment struct {
    Methods []PaymentMethod `json:"methods"`
}
```

Register the field, its implementations, and optionally a custom discriminator:

```go
var _ = jsonschema.NewJSONSchemaMethod(
    Payment.Schema,
    jsonschema.WithInterface(Payment{}.Methods),
    jsonschema.WithInterfaceImpls(Payment{}.Methods, Card{}, BankTransfer{}),
    jsonschema.WithDiscriminator(Payment{}.Methods, "!kind"),
)
```

The default discriminator property is `!type`. Each implementation's JSON must
carry the discriminator expected by the generated unmarshaler. Slice elements
are decoded in order, and an invalid element reports its zero-based index
without partially assigning the destination slice.

The slice must be the direct field type. Fixed arrays, nested slices, named
slice containers, `Optional[[]I]`, `Nullable[[]I]`, and inline interface
declarations are rejected.

See the compiling [`examples/sealed_interface_slices`](https://github.com/tylergannon/go-gen-jsonschema/tree/main/examples/sealed_interface_slices)
package for schema and runtime coverage, including value and pointer
implementations.
