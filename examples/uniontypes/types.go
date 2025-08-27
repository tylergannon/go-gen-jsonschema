package uniontypes

//go:generate gen-jsonschema

import "time"

// Shape is an interface that represents any geometric shape.
// By using this interface with multiple implementations and registering them,
// we can create a union type in the JSON Schema.
type Shape interface {
	// Area calculates the area of the shape.
	// This is just a method to make this a proper interface
	// that all implementations must provide.
	Area() float64
}

// Circle implements the Shape interface.
type Circle struct {
	// Radius is the distance from the center to the edge.
	Radius float64 `json:"radius"`

	// Color is an optional display color.
	Color string `json:"color,omitempty"`
}

// Area calculates the area of the circle.
func (c Circle) Area() float64 {
	return 3.14159 * c.Radius * c.Radius
}

// Rectangle implements the Shape interface.
type Rectangle struct {
	// Width is the first dimension.
	Width float64 `json:"width"`

	// Height is the second dimension.
	Height float64 `json:"height"`

	// Color is an optional display color.
	Color string `json:"color,omitempty"`
}

// Area calculates the area of the rectangle.
func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}

// Triangle implements the Shape interface.
type Triangle struct {
	// Base is the length of the base.
	Base float64 `json:"base"`

	// Height is the height from the base to the opposite vertex.
	Height float64 `json:"height"`

	// Color is an optional display color.
	Color string `json:"color,omitempty"`
}

// Area calculates the area of the triangle.
func (t Triangle) Area() float64 {
	return 0.5 * t.Base * t.Height
}

// Drawing demonstrates how to use a union type (Shape) in a struct.
// The Shapes field will accept any type that implements the Shape interface
// that has been registered in the schema.go file.
type Drawing struct {
	// ID is a unique identifier.
	ID string `json:"id"`

	// Title is the name of the drawing.
	Title string `json:"title"`

	// CreatedAt is when the drawing was created.
	CreatedAt time.Time `json:"createdAt"`

	// NOTE: Arrays of interface types ([]Shape) are not yet supported by the generator.
	// The tool currently only supports interface types as direct field types, not nested
	// within arrays, maps, or other composite types.
	//
	// To work around this limitation, you can either:
	// 1. Use a single interface field instead of an array
	// 2. Create a wrapper type that contains the interface
	//
	// For demonstration, we use a single Shape field here:
	MainShape Shape `json:"mainShape,omitempty"`

	// Original field that would be ideal but isn't supported yet:
	// Shapes []Shape `json:"shapes"`
}

// PaymentMethod is another example of an interface that creates a union type.
// This demonstrates how to use interfaces for domain modeling.
type PaymentMethod interface {
	// Process handles payment processing.
	// This is just a method to make this a proper interface.
	Process() error
}

// CreditCard implements PaymentMethod for credit card payments.
type CreditCard struct {
	// CardNumber is the credit card number.
	CardNumber string `json:"cardNumber"`

	// ExpiryMonth is the month the card expires.
	ExpiryMonth int `json:"expiryMonth"`

	// ExpiryYear is the year the card expires.
	ExpiryYear int `json:"expiryYear"`

	// CVV is the card verification value.
	CVV string `json:"cvv"`
}

// Process implements the PaymentMethod interface.
func (c CreditCard) Process() error {
	return nil
}

// BankTransfer implements PaymentMethod for bank transfers.
type BankTransfer struct {
	// AccountNumber is the bank account number.
	AccountNumber string `json:"accountNumber"`

	// RoutingNumber is the bank routing number.
	RoutingNumber string `json:"routingNumber"`

	// BankName is the name of the bank.
	BankName string `json:"bankName"`
}

// Process implements the PaymentMethod interface.
func (b BankTransfer) Process() error {
	return nil
}

// DigitalWallet implements PaymentMethod for digital wallet payments.
// This demonstrates using pointer receivers with interfaces.
type DigitalWallet struct {
	// Provider is the wallet provider name (e.g., "PayPal", "Apple Pay").
	Provider string `json:"provider"`

	// Email is the account email.
	Email string `json:"email"`

	// PhoneNumber is the associated phone number.
	PhoneNumber string `json:"phoneNumber,omitempty"`
}

// Process implements the PaymentMethod interface using a pointer receiver.
// This demonstrates how pointer receivers work with the interface registration.
func (d *DigitalWallet) Process() error {
	return nil
}

// Payment demonstrates using the PaymentMethod union type.
type Payment struct {
	// ID is a unique identifier.
	ID string `json:"id"`

	// Amount is the payment amount.
	Amount float64 `json:"amount"`

	// Currency is the payment currency code.
	Currency string `json:"currency"`

	// Date is when the payment occurred.
	Date time.Time `json:"date"`

	// Method is the payment method used.
	// This can be any registered implementation of PaymentMethod.
	Method PaymentMethod `json:"method"`
}
