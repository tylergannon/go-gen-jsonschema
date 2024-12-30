package builder

import (
	_ "embed"
	"errors"
	"fmt"
	_ "github.com/santhosh-tekuri/jsonschema" // just to ensure that `go mod tidy` won't delete this from our go.mod
	"github.com/tylergannon/go-gen-jsonschema/internal/loader"
	"github.com/tylergannon/go-gen-jsonschema/internal/typeregistry"
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

func initializeOutputTemplate() (outputTemplate *template.Template, err error) {
	return template.New("unmarshalerTemplate").Funcs(template.FuncMap{
		"lower": func(s string) string {
			r := []rune(s)
			r[0] = unicode.ToLower(r[0])
			return string(r)
		},
	}).Parse(unmarshalerTemplateData)
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
	*typeregistry.AlternativeTypeSpec
}

type InterfaceImpl struct {
	Discriminator string
	TypeName      string
	PrefixedName  string
}

type Interface struct {
	TypeName     string
	PrefixedName string
	FuncName     string
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
	// Need to get a hold of a list of the fields that are interface types,
	// and their functions.

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

	// Perform validation if required
	if validate {
		var pkg, err = loader.Load(JSONSchemaValidationPackagePath)
		if err != nil {
			return fmt.Errorf("failed to load schema verification package %s: %w", JSONSchemaValidationPackagePath, err)
		} else if len(pkg) != 1 {
			return fmt.Errorf("loaded package '%s' but got %d packages instead of 1", JSONSchemaValidationPackagePath, len(pkg))
		}
		builder.ImportMap.AddPackage(pkg[0])
		builder.JSONSchemaPrefix = builder.ImportMap.Alias(pkg[0])
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
					if typeregistry.HasPointerJSONUnmarshaler(namedType) {
						//	okay, implemented.
						continue
					} else {
						return fmt.Errorf("type %s defined in package %s has no unmarshaler.  quitting.", typeName, nodeWithAlts.Pkg().PkgPath)
					}
				}
				for _, _alt := range nodeWithAlts.TypeSpec.Alternatives() {
					importMap.AddPackage(_alt.TypeSpec.Pkg())
					var (
						alt   = typeAlt{AlternativeTypeSpec: _alt}
						found bool
					)

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
	if err := writeTemplate(fileName, builder); err != nil {
		log.Fatalf("failed to generate code: %v", err)
	}
	if exitCode, _, _, err := RunCommand("goimports", ".", "-w", fileName); err != nil || exitCode != 0 {
		log.Fatalf("error running goimports (exit code %d): %v", exitCode, err)
	}
	return nil
}

// writeTemplate accepts the configured UnmarshalerBuilder and writes the
// unmarshaler file template to disk, taking care to flush the results prior
// to exiting.
func writeTemplate(fileName string, builder *UnmarshalerBuilder) error {
	var (
		tmpFileName    = fileName + ".tmp"
		f, err         = os.OpenFile(tmpFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		outputTemplate *template.Template
	)
	if err != nil {
		return fmt.Errorf("failed to create temp file %s: %w", tmpFileName, err)
	}
	defer func() { // Ensure cleanup. Close file and delete temp file in case it's still present.
		if cerr := f.Close(); cerr != nil && !errors.Is(cerr, os.ErrClosed) {
			err = errors.Join(err, cerr)
		}
		if deleteErr := os.Remove(tmpFileName); deleteErr != nil && !errors.Is(deleteErr, os.ErrNotExist) {
			err = errors.Join(err, deleteErr)
		}
	}()
	// Initialize the output template if necessary
	if outputTemplate, err = initializeOutputTemplate(); err != nil {
		return fmt.Errorf("failed to initialize template: %w", err)
	} else if err = outputTemplate.Execute(f, builder); err != nil {
		return fmt.Errorf("template execute: %w", err)
	} else if err = f.Sync(); err != nil {
		return fmt.Errorf("failed to sync temp file %s: %w", tmpFileName, err)
	} else if err = os.Rename(tmpFileName, fileName); err != nil {
		return fmt.Errorf("failed to rename temp file %s to %s: %w", tmpFileName, fileName, err)
	}
	return err
}
