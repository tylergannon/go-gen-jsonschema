package builder

import (
	_ "embed"
	"fmt"
	"github.com/tylergannon/go-gen-jsonschema/internal/typeregistry"
	"go/types"
	"log"
	"os"
	"text/template"
)

//go:embed tmpl/unmarshalers.go.tmpl
var unmarshalerTemplateData string

var outputTemplate *template.Template

func initialize() (err error) {
	outputTemplate, err = template.New("unmarshalerTemplate").Parse(unmarshalerTemplateData)
	return err
}

type UnmarshalerBuilder struct {
	*ImportMap
	DiscriminatorPropName string
	SchemaDir             string
	TypeNames             []string
	TypesWithAlts         []typeWithAlts
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

func RenderGoCode(
	fileName, schemaDir string,
	graphs []*typeregistry.SchemaGraph,
	discriminatorMap *DiscriminatorMap,
) error {
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
