package syntax

import (
	"go/token"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
		file := &dst.File{Name: dst.NewIdent("borksauce")}
		ts := flattened.TypeSpec.Concrete
		ts.Type = flattened.Expr
		file.Decls = append(file.Decls, &dst.GenDecl{Tok: token.TYPE, Specs: []dst.Spec{
			ts,
		}})
		decorator.Print(file)
	})
	It("Should stringify the struct", func() {

	})
})
