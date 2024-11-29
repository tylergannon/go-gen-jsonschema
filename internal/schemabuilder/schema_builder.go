package schemabuilder

import (
	"fmt"
	"github.com/dave/dst/decorator"
	"github.com/tylergannon/go-gen-jsonschema/internal/typeregistry"
)

type SchemaBuilder struct {
	registry     *typeregistry.Registry
	localPkgPath string
	targets      map[typeregistry.TypeID]typeregistry.TypeSpec
}

func New(typeNames []string, pkgPath string, pkgs []*decorator.Package) (*SchemaBuilder, error) {
	registry, err := typeregistry.NewRegistry(pkgs)
	if err != nil {
		return nil, fmt.Errorf("could not create typeregister registry: %w", err)
	}
	builder := &SchemaBuilder{
		registry: registry,
	}

	// For each of the given types, search the types maps to ensure that all
	// dependency packages have been loaded.

	return builder, nil
}

func (b *SchemaBuilder) addType(name string, pkgPath string) error {
	if err := b.registry.LoadAndScan(pkgPath); err != nil {
		return fmt.Errorf("could not scan package %s: %w", pkgPath, err)
	}
	b.registry.GetType(name, pkgPath)

	return nil
}
