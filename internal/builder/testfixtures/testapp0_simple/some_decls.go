package testapp0_simple

import (
	"github.com/tylergannon/go-gen-jsonschema/internal/typeregistry/testfixtures/testapp0_simple/subpkg"
)

type (
	DeclaredTypeDirect struct {
		Foo int
		Bar string
	}
	DeclaredTypePointer *int
	// DeclaredTypeDefinition documentation
	DeclaredTypeDefinition      DeclaredTypeDirect
	DeclaredTypeAsPointer       *DeclaredTypeDefinition
	DeclaredAsRemoteType        subpkg.Baz
	DeclaredAsSliceOfRemoteType []subpkg.Baz
	DeclaredAsArrayOfRemoteType [10]subpkg.Baz
	StructWithVariousTypes      struct {
		DeclaredTypeDirect
		Field2 DeclaredTypeDirect
	}
	DeclaredTypeAlias        = DeclaredTypeDefinition
	DeclaredTypeAliasedAlias = DeclaredTypeAlias
)

func (def *DeclaredTypeDefinition) Func() int {
	return def.Foo
}
