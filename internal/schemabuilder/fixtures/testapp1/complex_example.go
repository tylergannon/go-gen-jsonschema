package testapp1

import "github.com/tylergannon/go-gen-jsonschema/internal/schemabuilder/fixtures/testapp1/subpkg"

// Build this struct in order to really get a lot of meaning out of life.
// It's really essential that you get all of this down.
type ComplexExample struct {
	// There can be comments here
	Foo int    `json:"foo"`
	Bar string `json:"bar"` // There can also be comments to the right
	// There can be
	// multiline comments
	// on a field
	Baz string `json:"baz"` // But in that case, this will be ignored.
	// Fields marked as "-" will be ignored.
	quux string `json:"-"`

	DefinedElsewhere subpkg.SubType `json:"definedElsewhere"`
}

type (
	// These are the comments that will be used, not the ones on the other type.
	ComplexDefinition ComplexExample
	// These comments will be used, not the ones on the aliased type.
	ComplexAlias = ComplexDefinition

	// These are the comments that will be used, not the ones on the other type.
	RemoteDefinition subpkg.SubType
	// These comments will be used, not the ones on the aliased type.
	RemoteAlias = subpkg.SubType
)
