package syntax

import (
	"fmt"
	"go/token"
	"go/types"
)

// EnumMarkerMethod is the name of the marker method that declares a named
// type as an enum: `func (T) enum() {}`. The generator infers enum
// membership for the type wherever it is used.
const EnumMarkerMethod = "enum"

// hasEnumMarker reports whether the named type declares the enum marker
// method with the required shape: a value receiver, no parameters, and no
// results. A method named enum with any other shape is an error naming the
// type rather than a silent miss.
func hasEnumMarker(pkg *types.Package, typeName string, position token.Position) (bool, error) {
	if pkg == nil {
		return false, nil
	}
	object, ok := pkg.Scope().Lookup(typeName).(*types.TypeName)
	if !ok {
		return false, nil
	}
	named, ok := types.Unalias(object.Type()).(*types.Named)
	if !ok {
		return false, nil
	}
	for method := range named.Methods() {
		if method.Name() != EnumMarkerMethod {
			continue
		}
		signature, ok := method.Type().(*types.Signature)
		if !ok {
			continue
		}
		if _, pointer := signature.Recv().Type().(*types.Pointer); pointer {
			return false, fmt.Errorf("enum marker on %s at %s must use a value receiver: func (%s) enum()", typeName, position, typeName)
		}
		if signature.Params().Len() != 0 || signature.Results().Len() != 0 {
			return false, fmt.Errorf("enum marker on %s at %s must take no parameters and return nothing: func (%s) enum()", typeName, position, typeName)
		}
		return true, nil
	}
	return false, nil
}
