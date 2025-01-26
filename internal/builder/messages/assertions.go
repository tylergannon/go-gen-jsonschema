package messages

//go:generate go run github.com/tylergannon/go-gen-jsonschema/gen-jsonschema/

// Call this function when the flattened struct info you receive has got a named type
// instead of pure-native Go types.  Named types are present in the flattened struct type
// when the type is a union type.  Look in the test data for the discriminator value
// used to select the union type, and present that along with the type name.
// Use the FULL package path of the type as found in the imports block of the file.
// List all the union types found in the struct.
// If the result of a function call also contains a union type, you can call this
// function again with new union types as you find them.
//
// NOTE: Do not bother with union types in the schema that are not present in the test data.
// Those will be handled by other test data.
type ToolFuncGetTypeInfo struct {
	// Include a separate entry for each union type instance found in the test data.
	UnionTypesFound []struct {
		// The value given in the test data that selects the union type.
		Discriminator string `json:"discriminator"`
		// The name of the interface type at the corresponding position of the struct, to the union type being instantiated.
		TypeName string `json:"typeName"`
		// The full package path of the type.  Leave blank if the type has no package alias (and is therefore defined in the current package).
		PkgPath string `json:"pkgPath"`
	} `json:"unionTypesFound"`
}

// A single assertion. There should be one assertion per value given in the test data.
// Whenever there is a union type, there should be a type assertion to validate the selected type,
// followed by value assertions to validate the values of the fields of the selected type.
type Assertion struct {
	// The path to the field in the struct.  Use dot notation to refer to nested fields.
	// Do not include the struct name in the path.  Use `[x]` to refer to the xth element of an array.
	//
	// Example paths:
	//
	// ```
	// ".Field1.NestedField1.Field2"
	// ".Field1[0].Field2"
	// ```
	Path string `json:"path"`
	// The value to assert.
	Value AssertionValue `json:"value"`
}

// The response to the generated test data request.
type GeneratedTestResponse struct {
	// The assertions to make on the generated test data.
	// There should be one assertion per value given in the test data.
	Assertions []Assertion `json:"assertions"`
}

type AssertionValue interface {
	assertionValueMarkerFunc()
}

var _ AssertionValue = AssertNumericValue{}
var _ AssertionValue = AssertStringValue{}
var _ AssertionValue = AssertBoolValue{}
var _ AssertionValue = AssertType{}

type AssertNumericValue struct {
	// The numeric value, encoded as a string. Use ordinary decimal notation. Omit the decimal point if the value is an integer.
	Value string `json:"value"`
}

func (AssertNumericValue) assertionValueMarkerFunc() {}

type AssertStringValue struct {
	// The string value.
	Value string `json:"value"`
}

func (AssertStringValue) assertionValueMarkerFunc() {}

type AssertBoolValue struct {
	// The boolean value.
	Value bool `json:"value"`
}

func (AssertBoolValue) assertionValueMarkerFunc() {}

// Use this whenever there is a union type.  Provide a single type assertion to validate the selected type.
type AssertType struct {
	// The full package path of the type.  Leave blank if the type has no package alias (and is therefore defined in the current package).
	PkgPath string `json:"pkgPath"`
	// The name of the type.
	TypeName string `json:"typeName"`
}

func (AssertType) assertionValueMarkerFunc() {}
