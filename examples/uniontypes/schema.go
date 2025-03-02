//go:build jsonschema
// +build jsonschema

package uniontypes

import (
	"encoding/json"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

// Schema method for Circle.
func (Circle) Schema() (json.RawMessage, error) {
	panic("not implemented")
}

// Schema method for Rectangle.
func (Rectangle) Schema() (json.RawMessage, error) {
	panic("not implemented")
}

// Schema method for Triangle.
func (Triangle) Schema() (json.RawMessage, error) {
	panic("not implemented")
}

// Schema method for Drawing.
func (Drawing) Schema() (json.RawMessage, error) {
	panic("not implemented")
}

// Schema method for CreditCard.
func (CreditCard) Schema() (json.RawMessage, error) {
	panic("not implemented")
}

// Schema method for BankTransfer.
func (BankTransfer) Schema() (json.RawMessage, error) {
	panic("not implemented")
}

// Schema method for DigitalWallet.
// Note that this matches the receiver type of the Process method.
func (*DigitalWallet) Schema() (json.RawMessage, error) {
	panic("not implemented")
}

// Schema method for Payment.
func (Payment) Schema() (json.RawMessage, error) {
	panic("not implemented")
}

// These marker variables register the types and interfaces.
var (
	// Register schema methods for concrete types
	_ = jsonschema.NewJSONSchemaMethod(Circle.Schema)
	_ = jsonschema.NewJSONSchemaMethod(Rectangle.Schema)
	_ = jsonschema.NewJSONSchemaMethod(Triangle.Schema)
	_ = jsonschema.NewJSONSchemaMethod(Drawing.Schema)
	_ = jsonschema.NewJSONSchemaMethod(CreditCard.Schema)
	_ = jsonschema.NewJSONSchemaMethod(BankTransfer.Schema)
	_ = jsonschema.NewJSONSchemaMethod((*DigitalWallet).Schema) // Note pointer receiver
	_ = jsonschema.NewJSONSchemaMethod(Payment.Schema)

	// Register Shape interface implementations
	// This is what creates the union type - it tells the generator that
	// any field of type Shape can contain a Circle, Rectangle, or Triangle.
	// The generated schema will include all three as possible types using anyOf.
	_ = jsonschema.NewInterfaceImpl[Shape](
		Circle{},    // Value receiver implementation
		Rectangle{}, // Value receiver implementation
		Triangle{},  // Value receiver implementation
	)

	// Register PaymentMethod interface implementations
	// This demonstrates including a pointer receiver implementation.
	// For pointer receivers, use (*Type)(nil) syntax.
	_ = jsonschema.NewInterfaceImpl[PaymentMethod](
		CreditCard{},          // Value receiver implementation
		BankTransfer{},        // Value receiver implementation
		(*DigitalWallet)(nil), // Pointer receiver implementation
	)
)
