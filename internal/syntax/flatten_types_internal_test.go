package syntax

import (
	"bytes"
	"fmt"
	"go/token"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/dave/dst/decorator/resolver/gopackages"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/tools/imports"
)

var _ = Describe("FlattenTypes", Ordered, func() {
	var (
		scanResult          ScanResult
		experimentRun       TypeSpec
		experimentRunStruct StructType
	)
	BeforeAll(func() {
		pkgs, err := Load("./testfixtures/structtype")
		Expect(err).NotTo(HaveOccurred())
		Expect(pkgs).To(HaveLen(1))
		Expect(pkgs[0].Name).To(Equal("structtype"))
		scanResult, err = loadPackageForTest(pkgs[0], "ExperimentRun", "ArrayOfSuperStruct")
		Expect(err).NotTo(HaveOccurred())
	})
	It("should flatten types", func() {
		Expect(scanResult.LocalNamedTypes).To(HaveLen(2))
	})
	It("Should have a struct type for ExperimentRun", func() {
		experimentRun = scanResult.LocalNamedTypes["ExperimentRun"]
		Expect(experimentRun.Name()).To(Equal("ExperimentRun"))
		st, ok := experimentRun.Type().Expr().(*dst.StructType)
		Expect(ok).To(BeTrue())
		Expect(st.Fields.List).To(HaveLen(16))

		experimentRunStruct = NewStructType(st, experimentRun)
	})
	It("Should flatten the struct", func() {
		flattened, err := experimentRunStruct.Flatten(scanResult.Pkg.PkgPath, scanResult.resolveType, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(flattened).To(BeAssignableToTypeOf(StructType{}))
		file := &dst.File{Name: dst.NewIdent(scanResult.Pkg.Name)}
		ts := flattened.TypeSpec.Concrete
		ts.Type = flattened.Expr

		file.Decls = append(file.Decls, &dst.GenDecl{Tok: token.TYPE, Specs: []dst.Spec{
			ts,
		}})
		buf := bytes.Buffer{}
		printer := decorator.NewRestorerWithImports(
			"github.com/tylergannon/go-gen-jsonschema/internal/syntax/testfixtures/structtype",
			gopackages.New("./testfixtures/structtype"),
		)
		err = printer.Fprint(&buf, file)
		Expect(err).NotTo(HaveOccurred())
		formatted, err := FormatCodeWithGoimports(buf.Bytes())
		Expect(err).NotTo(HaveOccurred())
		fmt.Println(string(formatted))
		// err = os.WriteFile("test.go", formatted, 0644)
		// Expect(err).NotTo(HaveOccurred())
	})
	It("Should stringify the struct", func() {

	})
})

func FormatCodeWithGoimports(source []byte) ([]byte, error) {
	options := &imports.Options{
		Comments:  true, // Preserve comments
		TabIndent: true, // Use tabs for indentation
	}

	formatted, err := imports.Process("", source, options)
	if err != nil {
		return nil, fmt.Errorf("failed to format code: %w", err)
	}

	return formatted, nil
}
