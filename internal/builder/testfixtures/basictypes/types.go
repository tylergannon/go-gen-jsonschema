package basictypes

//go:generate go run github.com/tylergannon/go-gen-jsonschema/gen-jsonschema/ --pretty

import (
	_ "github.com/dave/dst"
	_ "github.com/tylergannon/structtag"
)

// TypeInItsOwnDecl is an integer type that is the only item in its GenDecl
//
// ```go
// var foo = 1
// ```
//
// this should be squashed into a single line.
// Along with this.
type TypeInItsOwnDecl int

type (
	// TypeInNestedDecl is an integer type that's nested in a GenDecl
	TypeInNestedDecl int
)

type (
	// TypeInSharedDecl is an integer type that shares its GenDecl
	TypeInSharedDecl int

	// StringTypeInSharedDecl is foobarbaz
	StringTypeInSharedDecl string
)
