package schemabuilder

import (
	"encoding/json"
	"fmt"
	"github.com/fatih/structtag"
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
	"go/types"
	"golang.org/x/tools/go/packages"
	"log"
	"strings"
)

type SchemaBuilder struct {
	RootNamedType  *types.Named
	RootTypeStruct *types.Struct
	AllPackages    map[string]*packages.Package
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

// findNamedType looks for a type with the given name in the package scope
func findNamedType(pkg *packages.Package, typeName string) (*types.Named, error) {
	obj := pkg.Types.Scope().Lookup(typeName)
	if obj == nil {
		return nil, fmt.Errorf("type %q not found in package %q", typeName, pkg.PkgPath)
	}

	// Check if it's a type declaration
	typeObj, ok := obj.(*types.TypeName)
	if !ok {
		return nil, fmt.Errorf("%q is not a type", typeName)
	}

	// Get the named type
	named, ok := typeObj.Type().(*types.Named)
	if !ok {
		return nil, fmt.Errorf("%q is not a named type", typeName)
	}

	return named, nil
}

// getUnderlyingStruct gets the underlying struct type, following type aliases
func getUnderlyingStruct(named *types.Named) (*types.Struct, error) {
	// Handle type aliases by getting the final underlying type
	underlying := named.Underlying()

	// If it's a struct, return it
	if structType, ok := underlying.(*types.Struct); ok {
		return structType, nil
	}

	return nil, fmt.Errorf("type %q is not a struct type", named)
}

func (b *SchemaBuilder) scan() error {

	for i := 0; i < b.RootTypeStruct.NumFields(); i++ {
		field := b.RootTypeStruct.Field(i)

		tag := b.RootTypeStruct.Tag(i)
		fmt.Printf("Field of type %s with tag `%s`", field.Type().String(), tag)
		tags, err := structtag.Parse(tag)
		if err != nil {
			fmt.Printf("Error parsing struct tag %q: %v", tag, err)
			return fmt.Errorf("could not parse struct tag %q: %v", tag, err)
		}
		jsonTag, err := tags.Get("json")
		if err != nil {
			log.Fatalf("Required JSON annotations not implemented on struct %s", b.RootNamedType.String())
		}
		if jsonTag.Name == "-" {
			continue
		}
	}
	return nil
}

func New(pkgs map[string]*packages.Package, initialTypeName string, initialPkgPath string) (*SchemaBuilder, error) {
	pkg := pkgs[initialPkgPath]
	if pkg == nil {
		return nil, fmt.Errorf("package %q not found", initialPkgPath)
	}

	// Find the named type and its comments
	info, err := findTypeInfoInPackage(pkg, initialTypeName)
	if err != nil {
		return nil, err
	}

	// Get the underlying struct
	structType, err := getUnderlyingStruct(info.named)
	if err != nil {
		return nil, err
	}

	// Create the schema builder
	sb := &SchemaBuilder{
		RootNamedType:  info.named,
		RootTypeStruct: structType,
		AllPackages:    pkgs,
		Root: jsonschema.JSONSchema{
			Type:                 jsonschema.Object,
			Properties:           make(map[string]json.Marshaler),
			AdditionalProperties: false,
			Description:          info.comments,
		},
		Defs:    make(map[string]jsonschema.JSONSchema),
		DefsMap: make(map[string]string),
	}

	return sb, nil
}

func GenerateSchemas(pkg *packages.Package, typeNames []string) error {
	// Create a package map (we only have one for now)
	pkgs := map[string]*packages.Package{
		pkg.PkgPath: pkg,
	}

	// Process each type
	for _, typeName := range typeNames {
		sb, err := New(pkgs, strings.TrimSpace(typeName), pkg.PkgPath)
		if err != nil {
			return fmt.Errorf("error processing type %q: %v", typeName, err)
		}

		// TODO: Add schema generation logic here
		_ = sb // Using sb to prevent unused variable warning
		sb.scan()
	}

	return nil
}
