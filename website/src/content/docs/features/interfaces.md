---
title: Interfaces and discriminators
description: Generate discriminated unions for interface-typed Go fields.
---

An interface field becomes an `anyOf` over its registered implementations. The
generator also writes `UnmarshalJSON` dispatch code for the containing struct.

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
    Method PaymentMethod `json:"method"`
}
```

Register the field, its implementations, and optionally a custom discriminator:

```go
var _ = jsonschema.NewJSONSchemaMethod(
    Payment.Schema,
    jsonschema.WithInterface(Payment{}.Method),
    jsonschema.WithInterfaceImpls(Payment{}.Method, Card{}, BankTransfer{}),
    jsonschema.WithDiscriminator(Payment{}.Method, "!kind"),
)
```

The default discriminator property is `!type`. Each implementation's JSON must
carry the discriminator expected by the generated unmarshaler.

Only a single interface field is supported. Arrays or slices of interfaces and
inline interface declarations are rejected.

See the compiling [`examples/interfaces_options`](https://github.com/tylergannon/go-gen-jsonschema/tree/main/examples/interfaces_options)
package for the smallest complete registration.
