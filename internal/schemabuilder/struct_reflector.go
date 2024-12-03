package schemabuilder

import (
	"encoding/json"
	"fmt"
	"github.com/dave/dst"
	"github.com/tylergannon/go-gen-jsonschema/internal/typeregistry"
	"github.com/tylergannon/structtag"
	"go/types"
	"log"
)

type StructReflector struct {
	builder          *SchemaBuilder
	typeSpec         typeregistry.TypeSpec
	StructNode       *dst.StructType
	AddDiscriminator bool
}

func (r *StructReflector) typeName() *types.TypeName {
	return r.typeSpec.Pkg().Types.Scope().Lookup(r.typeSpec.GetTypeSpec().Name.Name).(*types.TypeName)
}

// flattenEmbeddedFields returns a flattened slice of `[]*dst.Field`
// in which any embedded fields have been (a) resolved to their actual
// definition and (b) the unnamed "embedded" field has been replaced by
// the list of valid fields for the embedded type.
// Anything that is not aresolvable struct type will result in an error.
// Also this function will lint for:
//  1. Uniqueness of json property names within the flattened list
//  2. All non-ignored field types in the flattened list must be valid:
//     no channels, complex numbers, interface types, private fields,
//     functions, or anything in the `sync` package.
func (r *StructReflector) flattenEmbeddedFields() ([]*dst.Field, error) {
	var (
		fields []*dst.Field
		//seen       = make(map[string]bool)
		//err        error
		//tempFields []*dst.Field
	)
	for _, field := range r.StructNode.Fields.List {
		if len(field.Names) == 0 {
			inspect("Embedded field.Type", field.Type)
			//if tempFields, err = r.flattenEmbeddedFields(field.Type.(*dst.StructType)); err != nil {
			//	return nil, fmt.Errorf("failed to flatten embedded field %v of type %T: %w", field, field, err)
			//}
			//fields = append(fields, tempFields...)
		}
	}

	return fields, nil
}

func (r *StructReflector) Render() json.Marshaler {
	var properties namedProps
	if r.AddDiscriminator {
		properties.Add(
			discriminatorKey,
			json.RawMessage(fmt.Sprintf(`{"const": "%s"}`, r.builder.getDiscriminator(r.typeSpec))),
		)
	}

	typeName := r.typeName()
	namedType := typeName.Type().(*types.Named)
	inspect("namedType", namedType)
	inspect("typeName", typeName)
	inspect("namedType.Underlying", namedType.Underlying())
	typeStruct := namedType.Underlying().(*types.Struct)

	_, _ = r.flattenEmbeddedFields()

	for i, field := range r.StructNode.Fields.List {
		// TODO: parse this elsewhere, earlier, so that we're not generating an error here.
		var (
			tags, err = structtag.Parse(field.Tag.Value[1 : len(field.Tag.Value)-1])
			jsonProp  string
		)
		if err != nil {
			log.Printf("failed to parse struct tag %s: %v", field.Tag.Value, err)
		}
		if jsonTag, err := tags.Get("json"); err == nil {
			jsonProp = jsonTag.Options[0]
			if jsonProp == "-" {
				continue
			}
		} else {
			jsonProp = field.Names[0].Name
		}

		typedField := typeStruct.Field(i)
		inspect("typedField type", typedField.Type())

		var newProp json.Marshaler

		newProp = r.resolveFieldType(typedField.Type(), field, newProp)

		properties.Add(jsonProp, newProp)

		switch typed := field.Type.(type) {
		case *dst.SelectorExpr:
			log.Printf("SelectorExpr: %v\n", typed)
		case *dst.Ident:
			log.Println(typed.Name)
			log.Printf("Ident: %v (obj: %v)\n", typed, typed.Obj)
		default:
			log.Printf("default: %T %v\n", typed, typed)
		}

	}
	return &StrictSchema{
		Description: buildComments(r.typeSpec.Decorations()),
		Props:       properties,
	}
}

func (r *StructReflector) resolveFieldType(fieldType types.Type, field *dst.Field, newProp json.Marshaler) json.Marshaler {
	switch typed := fieldType.(type) {
	case *types.Basic:
		if propValue, err := renderBasicType(typed, buildComments(field.Decorations())); err != nil {
			panic("need to return an error from this func")
		} else {
			newProp = propValue
		}
	case *types.Pointer:
		return r.resolveFieldType(typed.Elem(), field, newProp)
	}
	return newProp
}
