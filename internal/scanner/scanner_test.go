package scanner_test

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tylergannon/go-gen-jsonschema/internal/importmap"
	"github.com/tylergannon/go-gen-jsonschema/internal/scanner"
	"go/ast"
	"go/token"
	"path/filepath"
)

type fileSpecs struct {
	importMap importmap.ImportMap
	specs     []ast.Spec
}

func LoadDecls(path, fileName string, tok token.Token) []fileSpecs {
	var result []fileSpecs
	pkgs, err := scanner.Load(path)
	Expect(err).NotTo(HaveOccurred())
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			var pos = pkg.Fset.Position(file.Pos())
			if filepath.Base(pos.Filename) != fileName {
				continue
			}
			var fileSpec fileSpecs
			for _, decl := range file.Decls {
				genDecl, ok := decl.(*ast.GenDecl)
				if !ok || genDecl.Tok != tok {
					continue
				}
				for _, spec := range genDecl.Specs {
					fileSpec.specs = append(fileSpec.specs, spec)
				}
			}
			if len(fileSpec.specs) > 0 {
				fileSpec.importMap = importmap.New(file.Imports)
				result = append(result, fileSpec)
			}
		}
	}
	return result
}

var _ = Describe("FuncCallParser", Ordered, func() {
	var (
		specs []fileSpecs
		calls []scanner.MarkerFunctionCall
	)
	BeforeAll(func() {
		specs = LoadDecls("./testfixtures/typescanner", "calls.go", token.VAR)
		Expect(specs).To(HaveLen(1))
		Expect(specs[0].specs).To(HaveLen(6))
	})

	It("Figures them all out", func() {
		for _, spec := range specs[0].specs {
			_calls := scanner.ParseValueExprForMarkerFunctionCall(spec.(*ast.ValueSpec), specs[0].importMap)
			calls = append(calls, _calls...)
		}
		Expect(calls).To(HaveLen(6))
	})
	It("Call number 1", func() {
		_call := scanner.ParseValueExprForMarkerFunctionCall(specs[0].specs[0].(*ast.ValueSpec), specs[0].importMap)[0]
		Expect(_call.Function).To(Equal(scanner.MarkerFuncNewJSONSchemaMethod))
		Expect(_call.Arguments).To(HaveLen(1))
		Expect(_call.TypeArgument).To(BeNil())
	})
	It("Call number 2", func() {
		_call := scanner.ParseValueExprForMarkerFunctionCall(specs[0].specs[1].(*ast.ValueSpec), specs[0].importMap)[0]
		Expect(_call.Function).To(Equal(scanner.MarkerFuncNewJSONSchemaMethod))
		Expect(_call.Arguments).To(HaveLen(1))
		Expect(_call.TypeArgument).To(BeNil())
	})
	It("Call number 3", func() {
		_call := scanner.ParseValueExprForMarkerFunctionCall(specs[0].specs[2].(*ast.ValueSpec), specs[0].importMap)[0]
		Expect(_call.Function).To(Equal(scanner.MarkerFuncNewJSONSchemaBuilder))
		Expect(_call.Arguments).To(HaveLen(1))
		Expect(_call.TypeArgument).NotTo(BeNil())
		Expect(_call.TypeArgument.TypeName).To(Equal("TypeForSchemaFunction"))
		Expect(_call.TypeArgument.DeclaredLocally).To(BeTrue())
	})
	It("Call number 4", func() {
		_call := scanner.ParseValueExprForMarkerFunctionCall(specs[0].specs[3].(*ast.ValueSpec), specs[0].importMap)[0]
		Expect(_call.Function).To(Equal(scanner.MarkerFuncNewJSONSchemaBuilder))
		Expect(_call.Arguments).To(HaveLen(1))
		Expect(_call.TypeArgument).NotTo(BeNil())
		Expect(_call.TypeArgument.TypeName).To(Equal("PointerTypeForSchemaFunction"))
		Expect(_call.TypeArgument.Indirection).To(Equal(scanner.Pointer))
		Expect(_call.TypeArgument.DeclaredLocally).To(BeTrue())
	})
	It("Call number 5", func() {
		_call := scanner.ParseValueExprForMarkerFunctionCall(specs[0].specs[4].(*ast.ValueSpec), specs[0].importMap)[0]
		Expect(_call.Function).To(Equal(scanner.MarkerFuncNewInterfaceImpl))
		Expect(_call.Arguments).To(HaveLen(4))
		Expect(_call.TypeArgument).NotTo(BeNil())
		Expect(_call.TypeArgument.TypeName).To(Equal("MarkerInterface"))
		Expect(_call.TypeArgument.Indirection).To(Equal(scanner.NormalConcrete))
		Expect(_call.TypeArgument.DeclaredLocally).To(BeTrue())
	})
	It("Call number 6", func() {
		_call := scanner.ParseValueExprForMarkerFunctionCall(specs[0].specs[5].(*ast.ValueSpec), specs[0].importMap)[0]
		Expect(_call.Function).To(Equal(scanner.MarkerFuncNewEnumType))
		Expect(_call.Arguments).To(HaveLen(0))
		Expect(_call.TypeArgument).NotTo(BeNil())
		Expect(_call.TypeArgument.TypeName).To(Equal("NiceEnumType"))
		Expect(_call.TypeArgument.Indirection).To(Equal(scanner.NormalConcrete))
		Expect(_call.TypeArgument.DeclaredLocally).To(BeTrue())
	})
})

type commentsMap map[string][]string

func loadPackage() {
	pkgs, err := scanner.Load("./testfixtures/comments/...")
	Expect(err).ToNot(HaveOccurred())
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			pos := pkg.Fset.Position(file.Pos())
			fmt.Printf("## Begin file: %s\n", pos.Filename)
			//for i, comment := range file.Comments {
			//	fmt.Printf("Comment %s:%d: %s\n", file.TypeName, i, comment.Text())
			//}
			for _, decl := range file.Decls {
				if genDecl, ok := decl.(*ast.GenDecl); ok {
					switch genDecl.Tok {
					case token.TYPE:
						fmt.Println("## TYPE Declaration")
						for _, spec := range genDecl.Specs {
							fmt.Println(spec.(*ast.TypeSpec))
						}
					case token.VAR:
						fmt.Println("## VAR Declaration")

						for _, spec := range genDecl.Specs {
							vs := spec.(*ast.ValueSpec)
							if len(vs.Values) != 1 {
								continue
							}
							val := vs.Values[0]
							if callExpr, ok := val.(*ast.CallExpr); ok {
								scanner.DecodeFuncCall(callExpr)
							}

						}
					default:
						fmt.Printf("GenDecl: %v, %v\n", genDecl.Tok, genDecl.Specs)
					}
				}
			}
		}
	}
}

var _ = Describe("Scanner", func() {
	It("Basically does stuff", func() {
		//loadPackage()
	})
})
