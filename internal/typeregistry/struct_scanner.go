package typeregistry

import (
	"fmt"
	"github.com/dave/dst"
	"go/types"
	"log"
)

type NamedStructTypeSpec struct {
	typeSpec
	TypeName *types.TypeName
	// First one is the actual struct, following ones are embedded structs,
	// in order of their declaration
	Structs []*types.Struct
	// The resolved struct nodes, meaning that all embedded fields have been
	// tracked from the type name to their declaration.
	StructNodes      []*dst.StructType
	numInlineStructs int
}

var _ TypeSpec = (*NamedStructTypeSpec)(nil)

//func (r *Registry) resolveType(ts GetTypeSpec, typeExpr dst.Expr, t types.Type) (GetTypeSpec, error) {
//	file := ts.File()
//	imports := NewImportMap(ts.Pkg().PkgPath, file.Imports)
//
//}

//func (r *Registry) resolveInlineStructType

// The *typeSpec has already had its package loaded into the registry.
//
// What we need to do here:
// 1. Flatten embedded fields
func (r *Registry) registerNamedStructType(ts *typeSpec) (spec *NamedStructTypeSpec, err error) {
	var (
		typeName      = ts.Pkg().Types.Scope().Lookup(ts.GetTypeSpec().Name.Name).(*types.TypeName)
		namedType     = typeName.Type().(*types.Named)
		structs       = []*types.Struct{namedType.Underlying().(*types.Struct)}
		structNodes   = []*dst.StructType{ts.GetTypeSpec().Type.(*dst.StructType)}
		namedTypeSpec = &NamedStructTypeSpec{
			typeSpec:    *ts,
			TypeName:    typeName,
			Structs:     structs,
			StructNodes: structNodes,
		}
	)
	r.resolveStructFields(namedTypeSpec, namedType.Underlying().(*types.Struct), ts.GetTypeSpec().Type.(*dst.StructType))

	log.Println("Borkface")

	return namedTypeSpec, nil
}

// resolveType determines the type indicated by the two type nodes,
// looks up a package / file if needed, registers any needed types,
// and returns a TypeSpec of the correct subtype (or else error).
// May recurse upon itself or else call out to funcs like registerNamedStructType,
// which in turn will call back here.
// Guard against infinite recursion by building a type ID as soon as possible
// and checking the registry for that.
func (r *Registry) resolveType(ts TypeSpec, typeExpr dst.Expr, t types.Type) (TypeSpec, error) {
	inspect("Type Expr", typeExpr)
	inspect("Type", t)
	switch typedType := t.(type) {
	case *types.Basic:
		return resolveBasicType(typedType)
	case *types.Named:
		//r.GetType()
	}
	return nil, nil
}

// resolveStructFields will ensure that all types on the given struct
// exist within the type Registry and contain valid types.
// Returns error if any types are invalid or can't be resolved.
// Returns slices of type info representing any embedded structs.
func (r *Registry) resolveStructFields(
	parent *NamedStructTypeSpec,
	typeStruct *types.Struct,
	structNode *dst.StructType,
) (structs []*types.Struct, structNodes []*dst.StructType, fields []TypeSpec, err error) {
	structs = []*types.Struct{typeStruct}
	structNodes = []*dst.StructType{structNode}

	for structIdx := 0; structIdx < len(structs); structIdx++ {
		typeStruct = structs[structIdx]
		structNode = structNodes[structIdx]

		inspect("typeStruct", typeStruct)
		for i := 0; i < typeStruct.NumFields(); i++ {
			field := typeStruct.Field(i)
			fieldNode := structNode.Fields.List[i]
			if field.Embedded() {
				inspect("field is embedded", field)
				_, _ = r.resolveType(parent, fieldNode.Type, field.Type())

				//embeddedType := field.Type()
				//fieldNode.Type
			}
			switch typedFieldType := field.Type().(type) {
			case *types.Named:
				inspect("field type is named", typedFieldType)
			case *types.Struct:
				inspect("field type is struct", typedFieldType)
			case *types.Basic:
			default:
				inspect("what will it be?", typedFieldType)
			}
			inspect("field", fieldNode)
			inspect("field type", field.Type())

			inspect("field node", fieldNode)
		}
	}
	return structs, structNodes, fields, nil
}

//func (r *Registry) registerInlineStructType(parent *NamedStructTypeSpec) (spec *InlineStructTypeSpec, err error) {
//
//}

// InlineStructTypeSpec is for fields of a struct that are themselves
// inline struct definitions.  This is the only place we support
// inline struct declarations.
// The Name for the struct is "{parent_name}_inline{idx}"
type InlineStructTypeSpec struct {
	// We embed the parent type since it's guaranteed to be in the same
	// package and file.
	TypeSpec
	parentIdx   int
	structNodes []*dst.StructType
	structs     []*types.Struct
}

func (ts *InlineStructTypeSpec) ID() TypeID {
	parentID := ts.TypeSpec.ID()
	return TypeID(fmt.Sprintf("%s_inline%d", parentID, ts.parentIdx))
}

var _ TypeSpec = (*InlineStructTypeSpec)(nil)

func resolveBasicType(t *types.Basic) (BasicType, error) {
	switch t.Kind() {
	case types.String:
		return BasicTypeString, nil
	case types.Bool:
		return BasicTypeBool, nil
	case types.Int:
		return BasicTypeInt, nil
	case types.Float32, types.Float64:
		return BasicTypeFloat, nil
	case types.Int8, types.Int16, types.Int32, types.Int64, types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64, types.Uintptr:
		return BasicTypeInt, nil
	default:
		return BasicTypeString, fmt.Errorf("unsupported type %v", t.Kind())
	}
}
