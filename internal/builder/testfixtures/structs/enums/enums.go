package enums

// HotdogTopping is a string for hotdog toppings
type HotdogTopping string

const (
	// Mustard Bold and tangy zing.
	Mustard HotdogTopping = "mustard"
	// Ketchup Sweet and bright.
	Ketchup HotdogTopping = "ketchup"
	// Mayonnaise Rich and creamy.
	Mayonnaise HotdogTopping = "mayonnaise"
	// Relish Crisp, pickled crunch.
	Relish HotdogTopping = "relish"
	// Sauerkraut Sharp, fermented bite.
	Sauerkraut HotdogTopping = "sauerkraut"
)

type HotdogType string

const (
	Beef       HotdogType = "beef"
	Chicken    HotdogType = "chicken"
	Turkey     HotdogType = "turkey"
	Vegetarian HotdogType = "veggie"
)
