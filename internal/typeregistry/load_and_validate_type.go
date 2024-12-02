package typeregistry

import (
	"errors"
	"fmt"
	"github.com/dave/dst"
	"go/types"
)

func (r *Registry) LoadAndValidateNamedType(namedType *types.Named) (TypeSpec, error) {
	typeName := namedType.Obj()

	// Okay we should be able to find a type spec for the given type.
	// Of not then we scan the package.
	if err := r.LoadAndScan(typeName.Pkg().Path()); err != nil {
		return nil, fmt.Errorf("failed to load and scan package %q: %w", typeName.Pkg().Path(), err)
	}
	ts, _, ok := r.GetType(typeName.Name(), typeName.Pkg().Path())
	if !ok {
		return nil, fmt.Errorf("failed to find type %q in package %q", typeName.Name(), typeName.Pkg().Path())
	}
	switch underlying := namedType.Underlying().(type) {
	case *types.Struct:
		return r.LoadAndValidateStruct(underlying, ts.GetTypeSpec().Type.(*dst.StructType))
	}

	return r.LoadAndValidateType(namedType.Underlying())
}

var (
	ErrUnsupportedType = errors.New("unsupported type")
)

func (r *Registry) LoadAndValidateType(typ types.Type) (TypeSpec, error) {
	inspect("LoadAndValidateType", typ)
	switch t := typ.(type) {
	case *types.Basic:
		return resolveBasicType(t)
	case *types.Named:
		return r.LoadAndValidateNamedType(t)
	case *types.Pointer:
		return r.LoadAndValidateType(t.Elem())
	case *types.Chan:
		return nil, fmt.Errorf("chan types are not supported: %w", ErrUnsupportedType)
	case *types.Signature:
		return nil, fmt.Errorf("signature types are not supported: %w", ErrUnsupportedType)
	case *types.Interface:
		return nil, fmt.Errorf("interface types are not supported: %w", ErrUnsupportedType)
	case *types.Alias:
		return r.LoadAndValidateType(t.Underlying())
	}
	inspect("Type", typ)
	return nil, errors.New("unsupported type")
}

func (r *Registry) LoadAndValidateStruct(typ *types.Struct, str *dst.StructType) (TypeSpec, error) {
	inspect("Struct", typ)
	for i := 0; i < typ.NumFields(); i++ {
		field := typ.Field(i)
		if !field.Exported() {
			return nil, fmt.Errorf("exported fields are not supported: %w", ErrUnsupportedType)
		}
		_, err := r.LoadAndValidateType(field.Type())
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}
