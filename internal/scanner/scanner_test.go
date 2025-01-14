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

var _ = Describe("FuncCallParser", func() {
	var specs []fileSpecs
	BeforeEach(func() {
		specs = LoadDecls("./testfixtures/typescanner", "calls.go", token.VAR)
		Expect(specs).To(HaveLen(1))
		Expect(specs[0].specs).To(HaveLen(6))
	})

	It("Figures them all out", func() {
		for _, fileSpec := range specs {
			_, ok := fileSpec.importMap.GetGenJSONPrefix()
			if !ok {
				fmt.Println("No gen json prefix on file")
				continue
			}
			for _, spec := range fileSpec.specs {
				scanner.ParseValueExprForMarkerFunctionCall(spec.(*ast.ValueSpec), fileSpec.importMap)
			}
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
