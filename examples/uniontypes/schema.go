//go:build jsonschema

package uniontypes

import (
	"encoding/json"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

// Schema method for Circle.
func (Circle) Schema() json.RawMessage {
	panic("not implemented")
}

// Schema method for Rectangle.
func (Rectangle) Schema() json.RawMessage {
	panic("not implemented")
}

// Schema method for Triangle.
func (Triangle) Schema() json.RawMessage {
	panic("not implemented")
}

// Schema method for Drawing.
func (Drawing) Schema() json.RawMessage {
	panic("not implemented")
}

// Schema method for CreditCard.
func (CreditCard) Schema() json.RawMessage {
	panic("not implemented")
}

// Schema method for BankTransfer.
func (BankTransfer) Schema() json.RawMessage {
	panic("not implemented")
}

// Schema method for DigitalWallet.
// Note that this matches the receiver type of the Process method.
func (*DigitalWallet) Schema() json.RawMessage {
	panic("not implemented")
}

// Schema method for Payment.
func (Payment) Schema() json.RawMessage {
	panic("not implemented")
}

// These marker variables register the types and interfaces.
var (
	// Register schema methods for the concrete implementations - each is a
	// plain type with no interface field of its own.
	_ = jsonschema.Declare(Circle.Schema)
	_ = jsonschema.Declare(Rectangle.Schema)
	_ = jsonschema.Declare(Triangle.Schema)
	_ = jsonschema.Declare(CreditCard.Schema)
	_ = jsonschema.Declare(BankTransfer.Schema)
	_ = jsonschema.Declare((*DigitalWallet).Schema) // Note pointer receiver

	// Register Drawing along with its Shape union field. This is what
	// creates the union type - it tells the generator that Drawing.Shapes
	// can contain a Circle, Rectangle, or Triangle. Impl's first argument
	// is the exact discriminator value ("Circle", "Rectangle", "Triangle" -
	// the derived Go type names, matching what the split
	// WithInterface/WithInterfaceImpls form derived automatically).
	_ = jsonschema.Declare(Drawing.Schema).
		Interface(Drawing{}.Shapes,
			jsonschema.Impl("Circle", Circle{}),
			jsonschema.Impl("Rectangle", Rectangle{}),
			jsonschema.Impl("Triangle", Triangle{}),
		)

	// Register Payment along with its PaymentMethod union field. This
	// demonstrates including a pointer receiver implementation - for
	// pointer receivers, use (*Type)(nil) syntax.
	_ = jsonschema.Declare(Payment.Schema).
		Interface(Payment{}.Method,
			jsonschema.Impl("CreditCard", CreditCard{}),
			jsonschema.Impl("BankTransfer", BankTransfer{}),
			jsonschema.Impl("DigitalWallet", (*DigitalWallet)(nil)),
		)
)
