package syntax_test

import (
	"fmt"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tylergannon/go-gen-jsonschema/internal/syntax"
	"go/token"
	"path/filepath"
)

const (
	pkgPath = "github.com/tylergannon/go-gen-jsonschema/internal/syntax/testfixtures/typescanner"
	subpkg  = "github.com/tylergannon/go-gen-jsonschema/internal/syntax/testfixtures/typescanner/scannersubpkg"
)

type fileSpecs struct {
	file    *dst.File
	specs   []dst.Spec
	genDecl *dst.GenDecl
	pkg     *decorator.Package
}

func LoadDecls(path, fileName string, tok token.Token) []fileSpecs {
	var result []fileSpecs
	pkgs, err := syntax.Load(path)
	Expect(err).NotTo(HaveOccurred())
	for _, pkg := range pkgs {
		_, err = syntax.LoadPackage(pkg)
		Expect(err).NotTo(HaveOccurred())
		for _, file := range pkg.Syntax {
			var pos = nodePosition(pkg, file)
			if filepath.Base(pos.Filename) != fileName {
				continue
			}
			var fileSpec fileSpecs
			for _, decl := range file.Decls {
				genDecl, ok := decl.(*dst.GenDecl)
				if !ok || genDecl.Tok != tok {
					continue
				}
				fileSpec.genDecl = genDecl
				fileSpec.specs = append(fileSpec.specs, genDecl.Specs...)
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
		specs     []fileSpecs
		calls     []syntax.MarkerFunctionCall
		valueSpec = func(idx int) syntax.ValueSpec {
			return syntax.NewValueSpec(specs[0].genDecl, specs[0].specs[idx].(*dst.ValueSpec), specs[0].pkg, specs[0].file)
		}
	)

	BeforeAll(func() {
		specs = LoadDecls("./testfixtures/typescanner", "calls.go", token.VAR)
		Expect(specs).To(HaveLen(1))
		Expect(specs[0].specs).To(HaveLen(10))
	})

	It("Figures them all out", func() {
		for _, spec := range specs[0].specs {
			_calls := syntax.ParseValueExprForMarkerFunctionCall(syntax.NewValueSpec(specs[0].genDecl, spec.(*dst.ValueSpec), specs[0].pkg, specs[0].file))
			calls = append(calls, _calls...)
		}
		Expect(calls).To(HaveLen(10))
	})

	It("Call number 1", func() {
		_call := syntax.ParseValueExprForMarkerFunctionCall(valueSpec(0))[0]
		Expect(_call.CallExpr.MustIdentifyFunc().TypeName).To(Equal(syntax.MarkerFuncNewJSONSchemaMethod))
		Expect(_call.CallExpr.Args()).To(HaveLen(1))
		Expect(_call.TypeArgument()).To(BeNil())

		schemaMethod, err := _call.ParseSchemaMethod()

		Expect(err).NotTo(HaveOccurred())
		Expect(schemaMethod.FuncName).To(Equal("Schema"))
		Expect(schemaMethod.Receiver.Indirection).To(Equal(syntax.NormalConcrete))
		Expect(schemaMethod.Receiver.PkgPath).To(Equal(pkgPath))
		Expect(schemaMethod.Receiver.TypeName).To(Equal("TypeForSchemaMethod"))
	})
	It("Call number 2", func() {
		_call := syntax.ParseValueExprForMarkerFunctionCall(valueSpec(1))[0]
		Expect(_call.CallExpr.MustIdentifyFunc().TypeName).To(Equal(syntax.MarkerFuncNewJSONSchemaMethod))
		Expect(_call.CallExpr.Args()).To(HaveLen(1))
		Expect(_call.TypeArgument()).To(BeNil())
		schemaMethod, err := _call.ParseSchemaMethod()

		Expect(err).NotTo(HaveOccurred())
		Expect(schemaMethod.FuncName).To(Equal("Schema"))
		Expect(schemaMethod.Receiver.Indirection).To(Equal(syntax.Pointer))
		Expect(schemaMethod.Receiver.PkgPath).To(Equal(pkgPath))
		Expect(schemaMethod.Receiver.TypeName).To(Equal("PointerTypeForSchemaMethod"))
	})
	It("Call number 3", func() {
		_call := syntax.ParseValueExprForMarkerFunctionCall(valueSpec(2))[0]
		Expect(_call.CallExpr.MustIdentifyFunc().TypeName).To(Equal(syntax.MarkerFuncNewJSONSchemaBuilder))
		Expect(_call.CallExpr.Args()).To(HaveLen(1))
		Expect(_call.MustTypeArgument()).NotTo(BeNil())
		Expect(_call.MustTypeArgument().TypeName).To(Equal("TypeForSchemaFunction"))
		Expect(_call.MustTypeArgument()).ToNot(BeNil())
		Expect(_call.MustTypeArgument().PkgPath).To(Equal(pkgPath))
	})
	It("Call number 4", func() {
		_call := syntax.ParseValueExprForMarkerFunctionCall(valueSpec(3))[0]
		Expect(_call.CallExpr.MustIdentifyFunc().TypeName).To(Equal(syntax.MarkerFuncNewJSONSchemaBuilder))
		Expect(_call.CallExpr.Args()).To(HaveLen(1))
		Expect(_call.MustTypeArgument()).NotTo(BeNil())
		Expect(_call.MustTypeArgument().TypeName).To(Equal("PointerTypeForSchemaFunction"))
		Expect(_call.MustTypeArgument().Indirection).To(Equal(syntax.Pointer))
		Expect(_call.MustTypeArgument()).ToNot(BeNil())
		Expect(_call.MustTypeArgument().PkgPath).To(Equal(pkgPath))
	})

	It("Call number 5", func() {
		_call := syntax.ParseValueExprForMarkerFunctionCall(valueSpec(4))[0]
		Expect(_call.CallExpr.MustIdentifyFunc().TypeName).To(Equal(syntax.MarkerFuncNewInterfaceImpl))
		Expect(_call.CallExpr.Args()).To(HaveLen(4))
		Expect(_call.MustTypeArgument()).NotTo(BeNil())
		Expect(_call.MustTypeArgument().TypeName).To(Equal("MarkerInterface"))
		Expect(_call.MustTypeArgument().Indirection).To(Equal(syntax.NormalConcrete))
		Expect(_call.MustTypeArgument()).ToNot(BeNil())
		Expect(_call.MustTypeArgument().PkgPath).To(Equal(pkgPath))

		callArgs, err := _call.ParseTypesFromArgs()
		Expect(err).NotTo(HaveOccurred())
		Expect(callArgs).NotTo(BeEmpty())
		Expect(callArgs[0].TypeName).To(Equal("Type001"))
		Expect(callArgs[1].TypeName).To(Equal("Type002"))
		Expect(callArgs[2].TypeName).To(Equal("Type003"))
		Expect(callArgs[3].TypeName).To(Equal("Type004"))

	})
	It("Call number 6", func() {
		_call := syntax.ParseValueExprForMarkerFunctionCall(valueSpec(5))[0]
		Expect(_call.CallExpr.MustIdentifyFunc().TypeName).To(Equal(syntax.MarkerFuncNewEnumType))
		Expect(_call.CallExpr.Args()).To(HaveLen(0))
		Expect(_call.MustTypeArgument()).NotTo(BeNil())
		Expect(_call.MustTypeArgument().TypeName).To(Equal("NiceEnumType"))
		Expect(_call.MustTypeArgument().Indirection).To(Equal(syntax.NormalConcrete))
		Expect(_call.MustTypeArgument()).ToNot(BeNil())
		Expect(_call.MustTypeArgument().PkgPath).To(Equal(pkgPath))
	})

	It("Call number 7", func() {
		_call := syntax.ParseValueExprForMarkerFunctionCall(valueSpec(6))[0]
		Expect(_call.CallExpr.MustIdentifyFunc().TypeName).To(Equal(syntax.MarkerFuncNewJSONSchemaBuilder))
		Expect(_call.CallExpr.Args()).To(HaveLen(1))
		Expect(_call.MustTypeArgument()).NotTo(BeNil())
		Expect(_call.MustTypeArgument().TypeName).To(Equal("TypeForSchemaFunction"))
		Expect(_call.MustTypeArgument().Indirection).To(Equal(syntax.NormalConcrete))
		Expect(_call.MustTypeArgument()).NotTo(BeNil())
		Expect(_call.MustTypeArgument().PkgPath).To(Equal(subpkg))
	})

	It("Call number 8", func() {
		_call := syntax.ParseValueExprForMarkerFunctionCall(valueSpec(7))[0]
		Expect(_call.CallExpr.MustIdentifyFunc().TypeName).To(Equal(syntax.MarkerFuncNewJSONSchemaBuilder))
		Expect(_call.CallExpr.Args()).To(HaveLen(1))
		Expect(_call.MustTypeArgument()).NotTo(BeNil())
		Expect(_call.MustTypeArgument().TypeName).To(Equal("PointerTypeForSchemaFunction"))
		Expect(_call.MustTypeArgument().Indirection).To(Equal(syntax.Pointer))
		Expect(_call.MustTypeArgument()).NotTo(BeNil())
		Expect(_call.MustTypeArgument().PkgPath).To(Equal(subpkg))
	})

	It("Call number 9", func() {
		_call := syntax.ParseValueExprForMarkerFunctionCall(valueSpec(8))[0]
		Expect(_call.CallExpr.MustIdentifyFunc().TypeName).To(Equal(syntax.MarkerFuncNewInterfaceImpl))
		Expect(_call.CallExpr.Args()).To(HaveLen(4))
		Expect(_call.MustTypeArgument()).NotTo(BeNil())
		Expect(_call.MustTypeArgument().TypeName).To(Equal("MarkerInterface"))
		Expect(_call.MustTypeArgument().Indirection).To(Equal(syntax.NormalConcrete))
		Expect(_call.MustTypeArgument()).NotTo(BeNil())
		Expect(_call.MustTypeArgument().PkgPath).To(Equal(subpkg))

		callArgs, err := _call.ParseTypesFromArgs()
		Expect(err).NotTo(HaveOccurred())
		Expect(callArgs).NotTo(BeEmpty())
		Expect(callArgs[0].TypeName).To(Equal("Type001"))
		Expect(callArgs[0].PkgPath).To(Equal(subpkg))
		Expect(callArgs[1].TypeName).To(Equal("Type002"))
		Expect(callArgs[2].TypeName).To(Equal("Type003"))
		Expect(callArgs[2].Indirection).To(Equal(syntax.Pointer))
		Expect(callArgs[3].TypeName).To(Equal("Type004"))
		Expect(callArgs[3].Indirection).To(Equal(syntax.Pointer))
	})

	It("Call number 10", func() {
		_call := syntax.ParseValueExprForMarkerFunctionCall(valueSpec(9))[0]
		Expect(_call.CallExpr.MustIdentifyFunc().TypeName).To(Equal(syntax.MarkerFuncNewEnumType))
		Expect(_call.CallExpr.Args()).To(HaveLen(0))
		Expect(_call.MustTypeArgument()).NotTo(BeNil())
		Expect(_call.MustTypeArgument().TypeName).To(Equal("NiceEnumType"))
		Expect(_call.MustTypeArgument().Indirection).To(Equal(syntax.NormalConcrete))
		Expect(_call.MustTypeArgument()).NotTo(BeNil())
		Expect(_call.MustTypeArgument().PkgPath).To(Equal(subpkg))
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

var _ = printStuff

func nodePosition(pkg *decorator.Package, node dst.Node) token.Position {
	return pkg.Fset.Position(nodePos(pkg, node))
}
func nodePos(pkg *decorator.Package, node dst.Node) token.Pos {
	return pkg.Decorator.Map.Ast.Nodes[node].Pos()
}
