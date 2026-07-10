package syntax

import (
	"go/token"
	"testing"

	"github.com/dave/dst"
	"github.com/stretchr/testify/require"
)

func TestSkipField(t *testing.T) {
	tests := []struct {
		name  string
		names []string
		tag   string
		skip  bool
	}{
		{name: "one exported name", names: []string{"Exported"}},
		{name: "one unexported name", names: []string{"unexported"}, skip: true},
		{name: "group with an exported name", names: []string{"unexported", "Exported"}},
		{name: "group with all unexported names", names: []string{"first", "second"}, skip: true},
		{name: "JSON ignored", names: []string{"Exported"}, tag: `json:"-"`, skip: true},
		{name: "explicit schema ref", names: []string{"Exported"}, tag: `jsonschema:"ref=definitions/Other"`, skip: true},
		{name: "embedded field"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := &dst.Field{Type: dst.NewIdent("string")}
			for _, name := range tt.names {
				field.Names = append(field.Names, dst.NewIdent(name))
			}
			if tt.tag != "" {
				field.Tag = &dst.BasicLit{Kind: token.STRING, Value: "`" + tt.tag + "`"}
			}

			require.Equal(t, tt.skip, skipField(field))
		})
	}
}

func TestRegisteredInterfaceIdentifierResolves(t *testing.T) {
	scan := loadTypeScannerResult(t)
	holder := scan.LocalNamedTypes["ResolutionHolder"]
	fields := holder.Type().Expr().(*dst.StructType).Fields.List

	for _, field := range fields {
		t.Run(field.Names[0].Name, func(t *testing.T) {
			require.NoError(t, scan.resolveTypeExpr(NewExpr(field.Type, holder.Pkg(), holder.File()), nil))
		})
	}
}

func TestUnknownLocalIdentifierFails(t *testing.T) {
	scan := loadTypeScannerResult(t)
	holder := scan.LocalNamedTypes["ResolutionHolder"]
	fieldType := holder.Type().Expr().(*dst.StructType).Fields.List[0].Type.(*dst.Ident)
	originalName := fieldType.Name
	fieldType.Name = "UnknownLocal"
	t.Cleanup(func() { fieldType.Name = originalName })

	err := scan.resolveTypeExpr(NewExpr(fieldType, holder.Pkg(), holder.File()), nil)
	require.ErrorContains(t, err, "undeclared local UnknownLocal type")
	require.ErrorContains(t, err, "local_func_defs.go")
}

func TestAlreadyLoadedRegisteredEnumResolves(t *testing.T) {
	pkgs, err := Load("./testfixtures/typescanner")
	require.NoError(t, err)
	require.NotEmpty(t, pkgs)

	deps := make(map[string]ScanResult)
	root := newScanResult(pkgs[0], deps)
	remote := newScanResult(pkgs[0], deps)
	remote.Constants["RemoteEnum"] = &EnumSet{}
	const remotePath = "example.com/remoteenum"
	deps[remotePath] = remote
	root.remoteTypes.addType(remotePath, "RemoteEnum")

	require.NotPanics(t, func() {
		err = root.resolveTypes()
	})
	require.NoError(t, err)
}

func TestIsTimeType(t *testing.T) {
	require.True(t, IsTimeType("time", "Time"))
	require.False(t, IsTimeType("time", "Duration"))
	require.False(t, IsTimeType("example.com/time", "Time"))
}

func TestTimeTypeIsNotRegisteredAsRemote(t *testing.T) {
	pkgs, err := Load("./testfixtures/typescanner")
	require.NoError(t, err)
	require.NotEmpty(t, pkgs)

	tests := []struct {
		name string
		expr dst.Expr
	}{
		{name: "decorated ident", expr: &dst.Ident{Path: "time", Name: "Time"}},
		{name: "selector fallback", expr: &dst.SelectorExpr{X: dst.NewIdent("time"), Sel: dst.NewIdent("Time")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scan := newScanResult(pkgs[0], make(map[string]ScanResult))
			require.NoError(t, scan.resolveTypeExpr(NewExpr(tt.expr, nil, nil), nil))
			require.Empty(t, scan.remoteTypes)
		})
	}
}

func TestRendererRejectedShapesAreDiscoveryBoundaries(t *testing.T) {
	tests := []struct {
		name string
		expr dst.Expr
	}{
		{name: "map", expr: &dst.MapType{Key: dst.NewIdent("string"), Value: dst.NewIdent("string")}},
		{name: "interface", expr: &dst.InterfaceType{Methods: &dst.FieldList{}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scan := newScanResult(nil, make(map[string]ScanResult))
			require.NoError(t, scan.resolveTypeExpr(NewExpr(tt.expr, nil, nil), nil))
			require.Empty(t, scan.remoteTypes)
		})
	}
}

func loadTypeScannerResult(t *testing.T) ScanResult {
	t.Helper()
	pkgs, err := Load("./testfixtures/typescanner")
	require.NoError(t, err)
	require.NotEmpty(t, pkgs)
	scan, err := LoadPackage(pkgs[0])
	require.NoError(t, err)
	return scan
}
