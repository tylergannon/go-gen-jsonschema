package scanner

import (
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/packages"
)

type PackageScanner interface {
	Scan(*packages.Package) ([]TypeDecl, error)
}

// Comments
// A declaration can have doc comments either "Before" or "After" the
// declaration.
// ## "After" Comments:
// (a) MUST begin on the same line as the ending of the declaration.
// (b) MAY be multi-line as long as it is a block comment or as long as all the line comments have the same indentation.
// ```Example "After" comments:
//
//	type ExampleTypeWithAfterComments struct {
//	    FieldWithOneLineAfterComment int // this is a one-line "After" comment.
//	    FieldWithMultiLineAfterComment int // this is a multi-line "After" comment.
//	                                       // it trails after the declaration, but it
//	                                       // can be distinguished by the way that it
//	                                       // maintains the same indentation on each line.
//	    FieldWithSingleLineBlockAfterComment int /* This is a block-comment that only takes one line and counts as a "After" comment.  */
//	    FieldWithMultiLineBlockAfterComment int /*
//	       This is a block comment in a "After" comment.
//	       	It can do whatever it wants to with indentation or formatting.
//	      Because it continues until the end of the block comment.
//	       /*
//	} // The end of a type declaration may also contain "After" comments.
//
// ## "Before" Comments:
// Follow the same rules for composition as After comments, except that some
// rules are needed to establish the beginning of the Before comment.
// (a) MUST be a single contiguous block of comments that are adjacent to the declaration.
// (b) MAY be multiple lines, as long as they are block comments or contiguous line comments
//
//	having the same level of indentation.
//
// ```Example "Before" comments:
// // This is a before comment to a GenDecl.
// type (
//
//	/* this one isn't a before comment or an after comment */
//	      // This also isn't a before comment, because it has a different level of indentation than the following.
//	  // Here is a valid multi-line comment.
//	  // Each line has got the same indentation.
//	  // The indentation is based upon the position of the first slash "/" character, not the beginning of the text.
//	  type Foo int
//
// )
type Comments struct {
	// Each individual comment gets its own string.  I think this means that a multi-line line comment
	// should have multiple entries, whereas a block comment (slash-star style) will have a single entry,
	// though I might be mistaken.
	After  []string
	Before []string
}

type StructFieldInfo struct {
	Index int
	// Name is left empty in the event that the field is embedded.  Embedded field
	// name is a meaningless proposition.
	Name     string
	Embedded bool
	Tag      string // the full text of the tag.
	Comments Comments
	// The entire
	TypeInfo           ast.Expr
	InlineStructFields []StructFieldInfo
}

// TypeDecl refers to the
type TypeDecl struct {
	Node     *ast.DeclStmt
	Comments Comments
	Pkg      *packages.Package
	File     *ast.File
	Pos      token.Pos
	Decls    []TypeInfo
}

type StructTypeInfo struct {
	typeInfo
	Fields []StructFieldInfo
}

type InterfaceTypeInfo struct {
	typeInfo
}

// AliasTypeInfo
// refers to type alias declarations.
type AliasTypeInfo struct {
	typeInfo
}

// DefinitionTypeInfo
// A Definition type Refers to any type that's declared as an instance of another type,
// including type-instantiated generics.
// e.g.:
// ```
// type MyInt int
// type MyString string
// type MyFooType somepackage.SomeGenericType[MyInt, MyType]
// ```
type DefinitionTypeInfo struct {
	typeInfo
}

var _ TypeInfo = StructTypeInfo{}
var _ TypeInfo = InterfaceTypeInfo{}

type typeInfo struct {
	Identity *ast.Ident
	TypeInfo ast.Expr
	File     *ast.File
	FileSet  *token.FileSet
	Pkg      *packages.Package
	Pos      token.Pos
	Comments []string
	Decl     *TypeDecl
}

type TypeInfo interface {
	GetTypeName() string
	GetPkgPath() string
	GetTypeInfo() ast.Expr
	GetFile() *ast.File
	GetPkg() *packages.Package
	GetTokenPos() token.Pos
	GetComments() []string
	GetDecl() *TypeDecl
}

func (t typeInfo) GetPosition() token.Position {
	return t.FileSet.Position(t.Pos)
}

func (t typeInfo) GetComments() []string {
	return t.Comments
}

func (t typeInfo) GetDecl() *TypeDecl {
	return t.Decl
}

// Getters for typeInfo fields.
func (t typeInfo) GetTypeName() string {
	return t.Identity.Name
}

func (t typeInfo) GetPkgPath() string {
	return t.Pkg.PkgPath
}

func (t typeInfo) GetTypeInfo() ast.Expr {
	return t.TypeInfo
}

func (t typeInfo) GetFile() *ast.File {
	return t.File
}

func (t typeInfo) GetPkg() *packages.Package {
	return t.Pkg
}

func (t typeInfo) GetTokenPos() token.Pos {
	return t.Pos
}
