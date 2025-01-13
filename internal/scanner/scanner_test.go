package scanner_test

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tylergannon/go-gen-jsonschema/internal/scanner"
)

type MyType struct {
	InlineStructField struct {
		ArrayOfInlineStruct []struct {
			FooField string
		}
	}
}

type commentsMap map[string][]string

func loadPackage() {
	pkgs, err := scanner.Load("./testfixtures/comments/...")
	Expect(err).ToNot(HaveOccurred())
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			for i, comment := range file.Comments {
				fmt.Printf("Comment %s:%d: %s\n", file.Name, i, comment.Text())
			}
			fmt.Printf("## Begin file: %s\n", file.Package)
			for _, decl := range file.Decls {
				fmt.Printf("GenDecl: %T, %v\n", decl, decl)
			}
		}
	}
}

var _ = Describe("Scanner", func() {
	It("Basically does stuff", func() {
		loadPackage()
	})
})
