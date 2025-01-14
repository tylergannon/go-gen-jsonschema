package scanner_test

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tylergannon/go-gen-jsonschema/internal/scanner"
	"go/ast"
	"go/token"
	"path/filepath"
)

type MyType struct {
	InlineStructField struct {
		ArrayOfInlineStruct []struct {
			FooField string
		}
	}
}

func LoadDecls(path, fileName string, tok token.Token) []ast.Spec {
	fmt.Println("Looking for decls ", tok)
	var result []ast.Spec
	pkgs, err := scanner.Load(path)
	Expect(err).NotTo(HaveOccurred())
	for _, pkg := range pkgs {
		fmt.Println(pkg.PkgPath)
		for _, file := range pkg.Syntax {
			var pos = pkg.Fset.Position(file.Pos())
			fmt.Println(pos.Filename)
			if filepath.Base(pos.Filename) != fileName {
				fmt.Println("Nope", filepath.Base(pos.Filename), fileName)
				continue
			} else {
				fmt.Println("Equals", filepath.Base(pos.Filename), fileName)
			}
			fmt.Println("Has Decls ", len(file.Decls))
			for _, decl := range file.Decls {
				genDecl, ok := decl.(*ast.GenDecl)
				if !ok || genDecl.Tok != tok {
					fmt.Println("Nope", genDecl.Tok, pos.Filename)
					continue
				}
				fmt.Println("Found items ", len(genDecl.Specs))
				for _, spec := range genDecl.Specs {
					fmt.Printf("spec %+v\n", spec)
					result = append(result, spec)
				}
			}
		}
	}
	return result
}

var _ = Describe("FuncCallParser", func() {
	var specs []ast.Spec
	BeforeEach(func() {
		specs = LoadDecls("./testfixtures/typescanner", "calls.go", token.VAR)
		Expect(specs).To(HaveLen(12))
	})
	It("Figures them all out", func() {
		for _, spec := range specs {
			fmt.Println(spec)
		}
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
			//	fmt.Printf("Comment %s:%d: %s\n", file.Name, i, comment.Text())
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
