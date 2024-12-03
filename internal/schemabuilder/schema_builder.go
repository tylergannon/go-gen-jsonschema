package schemabuilder

import (
	"encoding/json"
	"fmt"
	"github.com/tylergannon/go-gen-jsonschema/internal/typeregistry"
	"log"
	"slices"
)

const (
	// DefinitionsKey is the property name in the root object, where
	// definitions are stored.
	definitionsKey   = "$defs"
	discriminatorKey = "__type"
)

type SchemaBuilder struct {
	registry             *typeregistry.Registry
	localPkgPath         string
	target               string
	RootObject           SchemaReflector
	Definitions          map[string]SchemaReflector
	typeDiscriminatorMap map[string][]typeregistry.TypeSpec
}

func (b *SchemaBuilder) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.RootObject.Render())
}

var _ json.Marshaler = (*SchemaBuilder)(nil)

func (b *SchemaBuilder) getDiscriminator(ts typeregistry.TypeSpec) string {
	var (
		name              = ts.GetTypeSpec().Name.Name
		discriminator, ok = b.typeDiscriminatorMap[name]
	)
	//inspect()
	if !ok {
		b.typeDiscriminatorMap[name] = []typeregistry.TypeSpec{ts}
		return ts.GetTypeSpec().Name.Name
	}
	idx := slices.IndexFunc(discriminator, func(it typeregistry.TypeSpec) bool {
		return it.ID() == ts.ID()
	})
	if idx == 0 {
		return ts.GetTypeSpec().Name.Name
	}
	if idx == -1 {
		b.typeDiscriminatorMap[name] = append(discriminator, ts)
		idx = len(b.typeDiscriminatorMap[name]) - 1
	}
	return fmt.Sprintf("%s_%d", name, idx)
}

type SchemaReflector interface {
	Render() json.Marshaler
}

func New(typeName string, pkgPath string, registry *typeregistry.Registry) (*SchemaBuilder, error) {
	if err := registry.LoadAndScan(pkgPath); err != nil {
		return nil, fmt.Errorf("failed to load and scan package: %w", err)
	}
	builder := &SchemaBuilder{
		registry:     registry,
		localPkgPath: pkgPath,
		target:       typeName,
	}
	//ts, _, found := registry.getType(typeName, pkgPath)
	//if !found {
	//	return nil, fmt.Errorf("type %s not found in package %s", typeName, pkgPath)
	//}
	//builder.RootObject = &StructReflector{
	//	builder:    builder,
	//	typeSpec:   ts,
	//	StructNode: ts.GetTypeSpec().Type.(*dst.StructType),
	//}

	return builder, nil
}
func inspect(str string, item any) {
	log.Printf("inspect %s: %T %v\n", str, item, item)
}
