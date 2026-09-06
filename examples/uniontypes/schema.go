//go:build jsonschema

package uniontypes

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
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
	_ = polytype.Declare(Circle.Schema)
	_ = polytype.Declare(Rectangle.Schema)
	_ = polytype.Declare(Triangle.Schema)
	_ = polytype.Declare(CreditCard.Schema)
	_ = polytype.Declare(BankTransfer.Schema)
	_ = polytype.Declare((*DigitalWallet).Schema) // Note pointer receiver

	// Drawing's Shapes field and Payment's Method field are unions of the
	// sealed Shape and PaymentMethod interfaces. Membership is inferred from
	// the unexported sealing method each interface declares: Circle,
	// Rectangle, and Triangle declare isShape; CreditCard, BankTransfer, and
	// DigitalWallet (on a pointer receiver, so it decodes as *DigitalWallet)
	// declare isPaymentMethod. The discriminator property is "type" and each
	// value is the concrete type name.
	_ = polytype.Declare(Drawing.Schema)
	_ = polytype.Declare(Payment.Schema)
)
