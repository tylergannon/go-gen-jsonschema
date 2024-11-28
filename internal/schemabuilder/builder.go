package schemabuilder

import (
	"fmt"
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
	"github.com/tylergannon/structtag"
	"golang.org/x/tools/go/packages"
	"log"
)

type SchemaBuilder struct {
	RootType  *StructTypeData
	ImportMap *PackageMap
	// The schema of the actual type we're making a schema for.
	Root jsonschema.JSONSchema
	// Any nested types can be either:
	// (a) A derivative type based on a basic type. Examples: `type Foo int`, `type Bar string`
	// (b) A derivative (named) type that is a named slice of something.
	//     Examples: `type SliceType []int`, `type SliceOfStruct []MyStruct`
	//     NOT ALLOWED: `type SliceOfInlineStruct []struct { Foo int }`
	// (c) Some named type e.g. `type MyType struct { Foo int }`, `type MyType MyOtherType`
	//     NOT ALLOWED (yet): `type MyType interface { ... }`
	// (d) A type alias e.g. `type MyType = MyOtherType`
	//
	// Inline type definitions are currently not allowed.
	//
	// Nested named types will NOT be defined on the attributes of an object but will ALWAYS be
	// added to the Defs, after checking the DefsMap to see if that type has already been defined.
	Defs map[string]jsonschema.JSONSchema
	// A map of type ids to ref string.
	// All named struct types MUST be added to Defs.  The DefsMap is keyed from the `Obj().ID()`
	// on the `*types.Named` object.
	DefsMap map[string]string
}

func (b *SchemaBuilder) scan() error {
	fields := b.RootType.StructType.Fields
	for i := 0; i < fields.NumFields(); i++ {
		field := fields.List[i]

		tag := field.Tag
		fmt.Printf("Field of type %v with tag `%s`", field.Type, tag.Value)
		tags, err := structtag.Parse(tag.Value)
		if err != nil {
			fmt.Printf("Error parsing struct tag %q: %v", tag, err)
			return fmt.Errorf("could not parse struct tag %q: %v", tag, err)
		}
		jsonTag, err := tags.Get("json")
		if err != nil {
			log.Fatalf("Required JSON annotations not implemented on struct %s", b.RootType.TypeSpec.Name.Name)
		}
		if jsonTag.Options[0] == "-" {
			continue
		}
	}
	return nil
}

func GenerateSchemas(pkg *packages.Package, typeNames []string) error {
	// Create a package map (we only have one for now)

	return nil
}
