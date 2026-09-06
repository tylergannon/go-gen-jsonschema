package syntax

import (
	"fmt"
	"go/token"
	"go/types"
	"slices"
	"strings"
)

// inferSealedUnion infers the wire membership of the named interface type
// typeName declared in pkg from its sealing method(s): the unexported methods
// its own body declares directly. It returns the variants in deterministic
// (scope-sorted) order.
//
// Every rule from issue #87 is enforced here and reported as an error naming
// the offending type, never as a silent omission:
//
//   - an interface with no unexported method of its own is not sealed;
//   - an interface whose only unexported methods arrive through an embedded
//     interface is unsupported;
//   - a variant is a named struct type in the same package that declares every
//     sealing method directly on T or *T and satisfies the complete interface;
//     the sealing method's receiver decides whether it is a value or pointer
//     variant;
//   - a type that satisfies the interface only through an embedded field
//     (inheriting the sealing method) is excluded, with its own diagnostic;
//   - a type that declares the sealing method directly but is not a struct or
//     does not satisfy the complete interface is an invalid candidate;
//   - a sealed interface with zero variants is an error.
func inferSealedUnion(pkg *types.Package, typeName string, position token.Position) ([]TypeID, error) {
	object, ok := pkg.Scope().Lookup(typeName).(*types.TypeName)
	if !ok {
		return nil, fmt.Errorf("interface %s at %s was not resolved by go/types", typeName, position)
	}
	iface, ok := object.Type().Underlying().(*types.Interface)
	if !ok {
		return nil, fmt.Errorf("type %s at %s is not an interface", typeName, position)
	}
	iface.Complete()

	var sealing []string
	for method := range iface.ExplicitMethods() {
		if !method.Exported() {
			sealing = append(sealing, method.Name())
		}
	}
	if len(sealing) == 0 {
		for method := range iface.Methods() {
			if !method.Exported() {
				return nil, fmt.Errorf("interface %s at %s acquires its sealing method %s by embedding another interface, which is unsupported; declare the unexported method directly in %s", typeName, position, method.Name(), typeName)
			}
		}
		return nil, fmt.Errorf("interface %s at %s is not sealed: it declares no unexported method of its own, so its membership cannot be inferred; non-sealed unions are unsupported", typeName, position)
	}

	var variants []TypeID
	for _, name := range pkg.Scope().Names() {
		candidate, ok := pkg.Scope().Lookup(name).(*types.TypeName)
		if !ok || candidate == object || candidate.IsAlias() {
			continue
		}
		named, ok := candidate.Type().(*types.Named)
		if !ok {
			continue
		}
		if _, isInterface := named.Underlying().(*types.Interface); isInterface {
			continue
		}
		// Explicit methods only: promoted methods are never listed on the
		// Named type itself, so a struct that inherits the sealing method
		// through an embedded field is never a direct candidate.
		declared := 0
		pointer := false
		for method := range named.Methods() {
			if !slices.Contains(sealing, method.Name()) {
				continue
			}
			declared++
			if signature, ok := method.Type().(*types.Signature); ok {
				if _, isPointer := signature.Recv().Type().(*types.Pointer); isPointer {
					pointer = true
				}
			}
		}
		if declared == 0 {
			if types.Implements(named, iface) || types.Implements(types.NewPointer(named), iface) {
				return nil, fmt.Errorf("type %s satisfies sealed interface %s (at %s) only through an embedded field and is excluded from the union; a variant must declare %s directly on %s or *%s", name, typeName, position, sealingList(sealing), name, name)
			}
			continue
		}
		if declared != len(sealing) {
			return nil, fmt.Errorf("type %s declares only some of the sealing methods of %s (at %s) directly; a variant must declare %s on %s or *%s", name, typeName, position, sealingList(sealing), name, name)
		}
		if _, isStruct := named.Underlying().(*types.Struct); !isStruct {
			return nil, fmt.Errorf("type %s declares the sealing method of %s (at %s) but is not a struct type; union variants must be named struct types", name, typeName, position)
		}
		var implementation types.Type = named
		if pointer {
			implementation = types.NewPointer(named)
		}
		if !types.Implements(implementation, iface) {
			receiver := name
			if pointer {
				receiver = "*" + name
			}
			return nil, fmt.Errorf("type %s declares the sealing method of %s (at %s) on %s but %s does not implement the complete interface", name, typeName, position, receiver, receiver)
		}
		variant := TypeID{PkgPath: pkg.Path(), TypeName: name}
		if pointer {
			variant.Indirection = Pointer
		}
		variants = append(variants, variant)
	}
	if len(variants) == 0 {
		return nil, fmt.Errorf("sealed interface %s at %s has no supported variants: no named struct type in package %s declares %s", typeName, position, pkg.Path(), sealingList(sealing))
	}
	return variants, nil
}

func sealingList(methods []string) string {
	quoted := make([]string, len(methods))
	for i, method := range methods {
		quoted[i] = method + "()"
	}
	return strings.Join(quoted, ", ")
}
