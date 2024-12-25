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
	SchemaDir string
	TypeNames []string
}

type Alt struct {
}

type HasAlt struct {
	Alts []Alt
}

func RenderGoCode(fileName, schemaDir string, graphs []*typeregistry.SchemaGraph) error {
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
		altsMap   = map[typeregistry.TypeID]*typeregistry.NamedTypeWithAltsNode{}
	)

	builder := &UnmarshalerBuilder{
		ImportMap: importMap,
		SchemaDir: schemaDir,
	}

	for _, graph := range graphs {
		if t, ok := graph.RootNode.Type().(*types.Named); !ok {
			return fmt.Errorf("graph root %T is not a named type", graph.RootNode.Type())
		} else {
			builder.TypeNames = append(builder.TypeNames, t.Obj().Name())
		}

		for id, node := range graph.Nodes {
			if alt, ok := node.(*typeregistry.NamedTypeWithAltsNode); ok {
				for _, _alt := range alt.TypeSpec.Alternatives() {
					importMap.AddPackage(_alt.TypeSpec.Pkg())

				}
				altsMap[id] = alt
			}
		}
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
