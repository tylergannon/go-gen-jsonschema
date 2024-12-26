package builder

import (
	_ "embed"
	"fmt"
	_ "github.com/santhosh-tekuri/jsonschema" // just to ensure that `go mod tidy` won't delete this from our go.mod
	"github.com/tylergannon/go-gen-jsonschema/internal/loader"
	"github.com/tylergannon/go-gen-jsonschema/internal/typeregistry"
	"go/types"
	"log"
	"os"
	"slices"
	"text/template"
)

const schemaVerificationPackage = "github.com/santhosh-tekuri/jsonschema"

//go:embed tmpl/unmarshalers.go.tmpl
var unmarshalerTemplateData string

var outputTemplate *template.Template

func initialize() (err error) {
	outputTemplate, err = template.New("unmarshalerTemplate").Parse(unmarshalerTemplateData)
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
}

func RenderGoCode(fileName, schemaDir string, graphs []*typeregistry.SchemaGraph, discriminatorMap *DiscriminatorMap, validate bool) error {
	if outputTemplate == nil {
		if err := initialize(); err != nil {
			return err
		}
	}
	// all given graphs will be in the same root package so we
	// use a single importMap
	if len(graphs) == 0 {
		return nil
	}
	var (
		importMap = NewImportMap(graphs[0].RootNode.Pkg())
		altsMap   = map[typeregistry.TypeID]typeWithAlts{}
	)

	if validate {
		if pkg, err := loader.Load(schemaVerificationPackage); err != nil {
			return fmt.Errorf("failed to load schema verification package %s: %w", schemaVerificationPackage, err)
		} else if len(pkg) != 1 {
			return fmt.Errorf("loaded package '%s' but got %d packages instead of 1", schemaVerificationPackage, len(pkg))
		} else {
			importMap.AddPackage(pkg[0])
		}
	}

	builder := &UnmarshalerBuilder{
		ImportMap:             importMap,
		SchemaDir:             schemaDir,
		DiscriminatorPropName: discriminatorPropName,
	}

	for _, graph := range graphs {
		if t, ok := graph.RootNode.Type().(*types.Named); !ok {
			return fmt.Errorf("graph root %T is not a named type", graph.RootNode.Type())
		} else {
			builder.TypeNames = append(builder.TypeNames, t.Obj().Name())
		}

		for id, node := range graph.Nodes {
			if nodeWithAlts, ok := node.(typeregistry.NamedTypeWithAltsNode); ok {
				var alts []typeAlt
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
					TypeName: nodeWithAlts.NamedType().Obj().Name(),
					Alts:     alts,
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
	if err = outputTemplate.Execute(f, builder); err != nil {
		return fmt.Errorf("template execute: %w", err)
	}
	if exitCode, _, _, err := RunCommand("goimports", ".", "-w", fileName); err != nil || exitCode != 0 {
		log.Fatalf("error running goimports (exit code %d): %v", exitCode, err)
	}
	return nil
}
