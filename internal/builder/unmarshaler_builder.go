package builder

import (
	_ "embed"
	"fmt"
	_ "github.com/santhosh-tekuri/jsonschema" // just to ensure that `go mod tidy` won't delete this from our go.mod
	"github.com/tylergannon/go-gen-jsonschema/internal/loader"
	"github.com/tylergannon/go-gen-jsonschema/internal/typeregistry"
	"go/token"
	"go/types"
	"log"
	"os"
	"slices"
	"text/template"
	"unicode"
)

const JSONSchemaValidationPackagePath = "github.com/santhosh-tekuri/jsonschema"

//go:embed tmpl/unmarshalers.go.tmpl
var unmarshalerTemplateData string

var outputTemplate *template.Template

func initialize() (err error) {
	outputTemplate, err = template.New("unmarshalerTemplate").Funcs(template.FuncMap{
		"lower": func(s string) string {
			r := []rune(s)
			r[0] = unicode.ToLower(r[0])
			return string(r)
		},
	}).Parse(unmarshalerTemplateData)
	return err
}

// UnmarshalerBuilder is the argument to the template
type UnmarshalerBuilder struct {
	// ImportMap holds information about the packages that must be imported in
	// the generated go code.
	*ImportMap
	DiscriminatorPropName string
	// SchemaDir is the relative path to where the json schema files have been
	// written
	SchemaDir string
	// TypeNames includes the names of the types, defined inside the local
	// package, that are being given json schemas.
	TypeNames []string
	// Validate is set true when the generated code should include
	// json schema validation in the json unmarshal logic.
	Validate bool
	// TypesWithAlts includes information about the types that need special
	// unmarshaling logic for discriminating between different possibilities in
	// a union type.
	TypesWithAlts []typeWithAlts

	// path prefix used for jsonschema package from santhosh-tekuri/jsonschema
	JSONSchemaPrefix string

	// Unmarshaler Only
	UnmarshalersOnly bool
}

// TypesWithoutAlts is a convenience function for the template, when --validate
// was specified.  These types will be given a `UnmarshalJSON` function that
// validates the incoming data and then passes the data to the standard
// json.Unmarshal function.
func (b *UnmarshalerBuilder) TypesWithoutAlts() []string {
	var result []string
	for _, t := range b.TypeNames {
		if !slices.ContainsFunc(b.TypesWithAlts, func(withAlts typeWithAlts) bool {
			return withAlts.TypeName == t
		}) {
			result = append(result, t)
		}
	}
	return result
}

func (b *UnmarshalerBuilder) HaveAlts() bool {
	return len(b.TypesWithAlts) > 0
}

type typeAlt struct {
	Discriminator string
	TypeName      string
	FuncName      string
}

type typeWithAlts struct {
	TypeName string
	Alts     []typeAlt
	// HasSchema denotes whether this is a top-level type that is meant to be
	// presented with a full schema.
	// Used when generating validation code to determine whether to validate
	// a given typeWithAlts
	HasSchema bool
}

func RenderGoCode(fileName, schemaDir string, graphs []*typeregistry.SchemaGraph, discriminatorMap *DiscriminatorMap, validate, unmarshalersOnly bool) error {
	// all given graphs will be in the same root package so we
	// use a single importMap
	if len(graphs) == 0 {
		return nil
	}
	var (
		importMap = NewImportMap(graphs[0].RootNode.Pkg())
		altsMap   = map[typeregistry.TypeID]typeWithAlts{}
	)

	builder := &UnmarshalerBuilder{
		ImportMap:             importMap,
		SchemaDir:             schemaDir,
		DiscriminatorPropName: discriminatorPropName,
		Validate:              validate,
		UnmarshalersOnly:      unmarshalersOnly,
	}
	for _, graph := range graphs {
		if t, ok := graph.RootNode.Type().(*types.Named); !ok {
			return fmt.Errorf("graph root %T is not a named type", graph.RootNode.Type())
		} else {
			builder.TypeNames = append(builder.TypeNames, t.Obj().Name())
		}
	}

	for _, graph := range graphs {
		for id, node := range graph.Nodes {
			if nodeWithAlts, ok := node.(typeregistry.NamedTypeWithAltsNode); ok {
				var (
					namedType = nodeWithAlts.NamedType()
					alts      []typeAlt
					typeName  = namedType.Obj().Name()
				)
				if importMap.localPackage.PkgPath != nodeWithAlts.Pkg().PkgPath {
					//	this was defined in a different package.  Its unmarshaler should likewise have
					// been defined already.
					if hasPointerJSONUnmarshaler(namedType) {
						//	okay, implemented.
						continue
					} else {
						return fmt.Errorf("type %s defined in package %s has no unmarshaler.  quitting.", typeName, nodeWithAlts.Pkg().PkgPath)
					}
				}
				for _, _alt := range nodeWithAlts.TypeSpec.Alternatives() {
					importMap.AddPackage(_alt.TypeSpec.Pkg())
					var (
						alt   typeAlt
						found bool
					)

					alt.FuncName = _alt.ConversionFunc
					alt.Discriminator, found = discriminatorMap.GetAlias(typeregistry.TypeID(fmt.Sprintf("%s~", _alt.TypeSpec.ID())))
					if !found {
						return fmt.Errorf("discriminator not found for type alt %s", _alt.TypeSpec.ID())
					}
					alt.TypeName = importMap.PrefixExpr(_alt.TypeSpec.GetTypeSpec().Name.Name, _alt.TypeSpec.Pkg())
					alts = append(alts, alt)
				}

				altsMap[id] = typeWithAlts{
					TypeName:  typeName,
					Alts:      alts,
					HasSchema: slices.Contains(builder.TypeNames, typeName),
				}
			}
		}
	}
	for _, t := range altsMap {
		builder.TypesWithAlts = append(builder.TypesWithAlts, t)
	}
	f, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	if validate {
		if pkg, err := loader.Load(JSONSchemaValidationPackagePath); err != nil {
			return fmt.Errorf("failed to load schema verification package %s: %w", JSONSchemaValidationPackagePath, err)
		} else if len(pkg) != 1 {
			return fmt.Errorf("loaded package '%s' but got %d packages instead of 1", JSONSchemaValidationPackagePath, len(pkg))
		} else {
			importMap.AddPackage(pkg[0])
			builder.JSONSchemaPrefix = importMap.Alias(pkg[0])
		}
	}

	if outputTemplate == nil {
		if err = initialize(); err != nil {
			return err
		}
	}
	if err = outputTemplate.Execute(f, builder); err != nil {
		return fmt.Errorf("template execute: %w", err)
	}
	if exitCode, _, _, err := RunCommand("goimports", ".", "-w", fileName); err != nil || exitCode != 0 {
		log.Fatalf("error running goimports (exit code %d): %v", exitCode, err)
	}
	return nil
}

// hasPointerJSONUnmarshaler reports whether the pointer to the given named type
// implements json.Unmarshaler.
//
// That is, for a type T, we check whether *T implements:
//
//	UnmarshalJSON([]byte) error
func hasPointerJSONUnmarshaler(named *types.Named) bool {
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
