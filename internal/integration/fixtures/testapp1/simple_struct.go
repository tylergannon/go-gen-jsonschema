package testapp1

import (
	_ "github.com/dave/dst/decorator"
	_ "github.com/santhosh-tekuri/jsonschema"
	_ "github.com/tylergannon/go-gen-jsonschema"
	_ "github.com/tylergannon/structtag"
	_ "golang.org/x/tools/go/packages"
)

//go:generate go run github.com/tylergannon/go-gen-jsonschema/gen-jsonschema/ -type SimpleStruct -pretty

// Build this struct in order to really get a lot of meaning out of life.
// It's really essential that you get all of this down.
type SimpleStruct struct {
	// There can be comments here
	Foo int    `json:"foo"`
	Bar string `json:"bar"` // There can also be comments to the right
	// There can be
	// multiline comments
	// on a field
	Baz string `json:"baz"`
	// Fields marked as "-" will be ignored.
	quux string `json:"-"`
}
