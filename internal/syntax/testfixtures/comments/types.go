package comments

/*

This comment doesn't go with any type because it stands alone.

*/

// This is another comment.
// this one belongs with it but neither one belongs with any given item.

// This comment belongs to the following declaration.
// This one does as well.
type (
	// This comment should be divorced because of its different indentation.
	Foobar struct {
		// Here is a separate comment.

		// This one will be connected to FieldOne.
		// It should have two lines.
		FieldOne int `json:"field_one"` // This is the "After" comment.
		// Separate comment.  Should be disconnected.

		// This one is connected to FieldTwo.
		// This one also has two lines.
		FieldTwo int `json:"field_two"` /* This is the "After" Comment.
		It should keep its formatting.
		*/
		/*This is the "BeforeComment for FieldThree.
		It has multiple lines.
		*/
		FieldThree int `json:"field_three"` // After comment for FieldThree.
	} // This is an "After" comment.

	// This is the "Before" comment for a defined type.
	IntDef int // This is the "After" comment for IntDef.

	// Type Alias Comment
	// Multi line
	AliasInGroupDecl = IntDef // After comment for Alias
) /* Here is an "After" comment that

should be associated with the TypeDecl.

It also has multiple lines.
*/

/*
Before comment for

StructDefInItsOwnGenDecl has multiple lines.
*/
type StructDefInItsOwnGenDecl struct {
} /*
After comment for StructDefInItsOwnGenDecl

Has multiple lines.

*/

// Before comment for TypeDefInOwnGenDecl
// Has two lines.
type TypeDefInOwnGenDecl Foobar // After comment for TypeDefInOwnGenDecl

// BeforeComment for TypeAliasInOwnGenDecl
type TypeAliasInOwnGenDecl = Foobar // AfterComment for TypeAliasInOwnGenDecl

// BeforeComment for IntType
type IntType int // After comment for IntType

// BeforeComment for Const definition
const ConstDefinedInOwnGenDecl IntType = 123 /*

This is a multi-line

Comment block

*/

// BeforeComment for a const GenDecl
const (
	// BeforeComment for const decl
	Const1 IntType = iota // after comment
	// BeforeComment for const2
	Const2 // const2 after comment
)

type StringType string

// BeforeComment for const declaration
const StringType1 StringType = "foobar" // AfterComment for const decl

// BeforeComment for
// gendecl
const (
	// BeforeComment
	// Multi line
	StringType2 StringType = "foobar" // After Comment
	// BeforeComment
	// Multi line
	StringType3 StringType = "foobar" // After comment
	// BeforeComment
	// Multi line
	StringType4 StringType = "foobar" // After comment
)

type StructType struct {
	Foobar Foobar `json:"field_one"`
	// Before Comment
	FieldThree int `json:"field_two"` // After comment

	Foobar2 *Foobar `json:"field_three"`

	InlineStructField struct {
		// 123 Before
		SecondInlineStructField struct {
			// ThirdInlineStructField Before Comment
			ThirdInlineStructField struct {
				// FourthInlineStructField Before Comment
				FourthInlineStructField struct {
					Field1 Foobar `json:"field_1"`
					Field2 Foobar `json:"field_2"`
					Field3 Foobar `json:"field_3"`
					Field4 Foobar `json:"field_4"`
				} `json:"field_one99"` // "After" Comment
			} `json:"field_one99"` // "After" Comment
		} `json:"field_one99"` // "After" Comment
	} `json:"field_one99"` // "After" Comment
}
