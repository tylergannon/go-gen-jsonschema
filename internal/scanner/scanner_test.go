package scanner_test

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tylergannon/go-gen-jsonschema/internal/scanner"
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/packages"
	"path/filepath"
)

const subpkg = "github.com/tylergannon/go-gen-jsonschema/internal/scanner/testfixtures/typescanner/scannersubpkg"

type fileSpecs struct {
	file  *ast.File
	specs []ast.Spec
	pkg   *packages.Package
}

func LoadDecls(path, fileName string, tok token.Token) []fileSpecs {
	var result []fileSpecs
	pkgs, err := scanner.Load(path)
	Expect(err).NotTo(HaveOccurred())
	for _, pkg := range pkgs {
		_, err := scanner.LoadPackage(pkg)
		Expect(err).NotTo(HaveOccurred())
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
				fileSpec.file = file
				fileSpec.pkg = pkg
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
		Expect(specs[0].specs).To(HaveLen(10))
	})

	It("Figures them all out", func() {
		for _, spec := range specs[0].specs {
			_calls := scanner.ParseValueExprForMarkerFunctionCall(spec.(*ast.ValueSpec), specs[0].file, specs[0].pkg)
			calls = append(calls, _calls...)
		}
		Expect(calls).To(HaveLen(10))
	})

	It("Call number 1", func() {
		_call := scanner.ParseValueExprForMarkerFunctionCall(specs[0].specs[0].(*ast.ValueSpec), specs[0].file, specs[0].pkg)[0]
		Expect(_call.Function).To(Equal(scanner.MarkerFuncNewJSONSchemaMethod))
		Expect(_call.Arguments).To(HaveLen(1))
		Expect(_call.TypeArgument).To(BeNil())

		schemaMethod, err := _call.ParseSchemaMethod()

		Expect(err).NotTo(HaveOccurred())
		Expect(schemaMethod.FuncName).To(Equal("Schema"))
		Expect(schemaMethod.Receiver.Indirection).To(Equal(scanner.NormalConcrete))
		Expect(schemaMethod.Receiver.DeclaredLocally).To(BeTrue())
		Expect(schemaMethod.Receiver.TypeName).To(Equal("TypeForSchemaMethod"))
	})
	It("Call number 2", func() {
		_call := scanner.ParseValueExprForMarkerFunctionCall(specs[0].specs[1].(*ast.ValueSpec), specs[0].file, specs[0].pkg)[0]
		Expect(_call.Function).To(Equal(scanner.MarkerFuncNewJSONSchemaMethod))
		Expect(_call.Arguments).To(HaveLen(1))
		Expect(_call.TypeArgument).To(BeNil())
		schemaMethod, err := _call.ParseSchemaMethod()

		Expect(err).NotTo(HaveOccurred())
		Expect(schemaMethod.FuncName).To(Equal("Schema"))
		Expect(schemaMethod.Receiver.Indirection).To(Equal(scanner.Pointer))
		Expect(schemaMethod.Receiver.DeclaredLocally).To(BeTrue())
		Expect(schemaMethod.Receiver.TypeName).To(Equal("PointerTypeForSchemaMethod"))
	})
	It("Call number 3", func() {
		_call := scanner.ParseValueExprForMarkerFunctionCall(specs[0].specs[2].(*ast.ValueSpec), specs[0].file, specs[0].pkg)[0]
		Expect(_call.Function).To(Equal(scanner.MarkerFuncNewJSONSchemaBuilder))
		Expect(_call.Arguments).To(HaveLen(1))
		Expect(_call.TypeArgument).NotTo(BeNil())
		Expect(_call.TypeArgument.TypeName).To(Equal("TypeForSchemaFunction"))
		Expect(_call.TypeArgument.DeclaredLocally).To(BeTrue())
	})
	It("Call number 4", func() {
		_call := scanner.ParseValueExprForMarkerFunctionCall(specs[0].specs[3].(*ast.ValueSpec), specs[0].file, specs[0].pkg)[0]
		Expect(_call.Function).To(Equal(scanner.MarkerFuncNewJSONSchemaBuilder))
		Expect(_call.Arguments).To(HaveLen(1))
		Expect(_call.TypeArgument).NotTo(BeNil())
		Expect(_call.TypeArgument.TypeName).To(Equal("PointerTypeForSchemaFunction"))
		Expect(_call.TypeArgument.Indirection).To(Equal(scanner.Pointer))
		Expect(_call.TypeArgument.DeclaredLocally).To(BeTrue())
	})

	It("Call number 5", func() {
		_call := scanner.ParseValueExprForMarkerFunctionCall(specs[0].specs[4].(*ast.ValueSpec), specs[0].file, specs[0].pkg)[0]
		Expect(_call.Function).To(Equal(scanner.MarkerFuncNewInterfaceImpl))
		Expect(_call.Arguments).To(HaveLen(4))
		Expect(_call.TypeArgument).NotTo(BeNil())
		Expect(_call.TypeArgument.TypeName).To(Equal("MarkerInterface"))
		Expect(_call.TypeArgument.Indirection).To(Equal(scanner.NormalConcrete))
		Expect(_call.TypeArgument.DeclaredLocally).To(BeTrue())

		callArgs, err := _call.ParseTypesFromArgs()
		Expect(err).NotTo(HaveOccurred())
		Expect(callArgs).NotTo(BeEmpty())
		Expect(callArgs[0].TypeName).To(Equal("Type001"))
		Expect(callArgs[1].TypeName).To(Equal("Type002"))
		Expect(callArgs[2].TypeName).To(Equal("Type003"))
		Expect(callArgs[3].TypeName).To(Equal("Type004"))

	})
	It("Call number 6", func() {
		_call := scanner.ParseValueExprForMarkerFunctionCall(specs[0].specs[5].(*ast.ValueSpec), specs[0].file, specs[0].pkg)[0]
		Expect(_call.Function).To(Equal(scanner.MarkerFuncNewEnumType))
		Expect(_call.Arguments).To(HaveLen(0))
		Expect(_call.TypeArgument).NotTo(BeNil())
		Expect(_call.TypeArgument.TypeName).To(Equal("NiceEnumType"))
		Expect(_call.TypeArgument.Indirection).To(Equal(scanner.NormalConcrete))
		Expect(_call.TypeArgument.DeclaredLocally).To(BeTrue())
	})

	It("Call number 7", func() {
		_call := scanner.ParseValueExprForMarkerFunctionCall(specs[0].specs[6].(*ast.ValueSpec), specs[0].file, specs[0].pkg)[0]
		Expect(_call.Function).To(Equal(scanner.MarkerFuncNewJSONSchemaBuilder))
		Expect(_call.Arguments).To(HaveLen(1))
		Expect(_call.TypeArgument).NotTo(BeNil())
		Expect(_call.TypeArgument.TypeName).To(Equal("TypeForSchemaFunction"))
		Expect(_call.TypeArgument.Indirection).To(Equal(scanner.NormalConcrete))
		Expect(_call.TypeArgument.DeclaredLocally).To(BeFalse())
		Expect(_call.TypeArgument.PkgPath).To(Equal(subpkg))
	})

	It("Call number 8", func() {
		_call := scanner.ParseValueExprForMarkerFunctionCall(specs[0].specs[7].(*ast.ValueSpec), specs[0].file, specs[0].pkg)[0]
		Expect(_call.Function).To(Equal(scanner.MarkerFuncNewJSONSchemaBuilder))
		Expect(_call.Arguments).To(HaveLen(1))
		Expect(_call.TypeArgument).NotTo(BeNil())
		Expect(_call.TypeArgument.TypeName).To(Equal("PointerTypeForSchemaFunction"))
		Expect(_call.TypeArgument.Indirection).To(Equal(scanner.Pointer))
		Expect(_call.TypeArgument.DeclaredLocally).To(BeFalse())
		Expect(_call.TypeArgument.PkgPath).To(Equal(subpkg))
	})

	It("Call number 9", func() {
		_call := scanner.ParseValueExprForMarkerFunctionCall(specs[0].specs[8].(*ast.ValueSpec), specs[0].file, specs[0].pkg)[0]
		Expect(_call.Function).To(Equal(scanner.MarkerFuncNewInterfaceImpl))
		Expect(_call.Arguments).To(HaveLen(4))
		Expect(_call.TypeArgument).NotTo(BeNil())
		Expect(_call.TypeArgument.TypeName).To(Equal("MarkerInterface"))
		Expect(_call.TypeArgument.Indirection).To(Equal(scanner.NormalConcrete))
		Expect(_call.TypeArgument.DeclaredLocally).To(BeFalse())
		Expect(_call.TypeArgument.PkgPath).To(Equal(subpkg))

		callArgs, err := _call.ParseTypesFromArgs()
		Expect(err).NotTo(HaveOccurred())
		Expect(callArgs).NotTo(BeEmpty())
		Expect(callArgs[0].TypeName).To(Equal("Type001"))
		Expect(callArgs[0].PkgPath).To(Equal(subpkg))
		Expect(callArgs[1].TypeName).To(Equal("Type002"))
		Expect(callArgs[2].TypeName).To(Equal("Type003"))
		Expect(callArgs[2].Indirection).To(Equal(scanner.Pointer))
		Expect(callArgs[3].TypeName).To(Equal("Type004"))
		Expect(callArgs[3].Indirection).To(Equal(scanner.Pointer))
	})

	It("Call number 10", func() {
		_call := scanner.ParseValueExprForMarkerFunctionCall(specs[0].specs[9].(*ast.ValueSpec), specs[0].file, specs[0].pkg)[0]
		Expect(_call.Function).To(Equal(scanner.MarkerFuncNewEnumType))
		Expect(_call.Arguments).To(HaveLen(0))
		Expect(_call.TypeArgument).NotTo(BeNil())
		Expect(_call.TypeArgument.TypeName).To(Equal("NiceEnumType"))
		Expect(_call.TypeArgument.Indirection).To(Equal(scanner.NormalConcrete))
		Expect(_call.TypeArgument.DeclaredLocally).To(BeFalse())
		Expect(_call.TypeArgument.PkgPath).To(Equal(subpkg))
	})
})

var _ = Describe("Scanner", func() {
	It("Basically does stuff", func() {
		//loadPackage()
	})
})

func printStuff(it any) {
	fmt.Printf("%T %#v\n", it, it)
}
