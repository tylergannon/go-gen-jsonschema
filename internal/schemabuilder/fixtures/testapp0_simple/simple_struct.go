package testapp0_simple

//go:generate go run github.com/tylergannon/go-gen-jsonschema/cmd/ -type SimpleStruct

// Build this struct in order to really get a lot of meaning out of life.
// It's really essential that you get all of this down.
type SimpleStruct struct {
	// There can be comments here
	Foo int    `json:"foo"`
	Bar string `json:"bar"` // There can also be comments to the right
	// There can be
	// multiline comments
	// on a field
	Baz string `json:"baz"` // But in that case, this will be ignored.
	// Fields marked as "-" will be ignored.
	quux string `json:"-"`
}

// Build this struct in order to really get a lot of meaning out of life.
// It's really essential that you get all of this down.
type SimpleStructWithPointer struct {
	// There can be comments here
	Foo int    `json:"foo"`
	Bar string `json:"bar"` // There can also be comments to the right
	// There can be
	// multiline comments
	// on a field
	Baz *string `json:"baz"` // But in that case, this will be ignored.
	// Fields marked as "-" will be ignored.
	quux *string `json:"-"`
}
