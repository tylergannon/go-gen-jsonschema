package typeregistry

import (
	"go/token"
	"go/types"
)

// HasPointerJSONUnmarshaler reports whether the pointer to the given named type
// implements json.Unmarshaler.
//
// That is, for a type T, we check whether *T implements:
//
//	UnmarshalJSON([]byte) error
func HasPointerJSONUnmarshaler(named *types.Named) bool {
	// Construct the UnmarshalJSON([]byte) error method signature.
	//
	//   func([]byte) error
	//
	// We need a parameter of type []byte and a result of type error.
	byteSlice := types.NewSlice(types.Typ[types.Byte])
	errorType := types.Universe.Lookup("error").Type()

	// Create the method signature:  func([]byte) error
	sig := types.NewSignatureType(
		nil,
		nil,
		nil,
		types.NewTuple(types.NewVar(token.NoPos, nil, "", byteSlice)), // parameters
		types.NewTuple(types.NewVar(token.NoPos, nil, "", errorType)), // results
		false,
	)

	// Create a Func that has the name "UnmarshalJSON" and that signature.
	unmarshalerFunc := types.NewFunc(token.NoPos, nil, "UnmarshalJSON", sig)

	// Create an interface that has just this one method.
	unmarshalerIface := types.NewInterfaceType(
		[]*types.Func{unmarshalerFunc},
		nil,
	)
	unmarshalerIface.Complete() // finalize the interface

	// Construct the pointer type to named (i.e., *named).
	ptrToNamed := types.NewPointer(named)

	// Finally, check if that pointer implements our interface.
	return types.Implements(ptrToNamed, unmarshalerIface)
}
