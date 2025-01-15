package builder

import (
	"github.com/dave/dst/decorator"
	"github.com/tylergannon/go-gen-jsonschema/internal/scanner"
)

func New(pkg *decorator.Package) (SchemaBuilder, error) {
	var builder = SchemaBuilder{
		LocalPkg: pkg,
	}
	data, err := scanner.LoadPackage(pkg)
	if err != nil {
		return builder, err
	}
	for _, _ = range data.SchemaMethods {
	}

	return builder, nil
}

type SchemaBuilder struct {
	LocalPkg *decorator.Package
	Packages map[string]scanner.ScanResult
}

// loadScanResult gets the scan result associated with the given scanner.TypeID
func (s SchemaBuilder) loadScanResult(t scanner.TypeID) (scanner.ScanResult, error) {
	var pkgPath = t.PkgPath
	if pkgPath == "" {
		pkgPath = s.LocalPkg.PkgPath
	}
	if _, ok := s.Packages[pkgPath]; !ok {
		if pkgs, err := scanner.Load(pkgPath); err != nil {
			return scanner.ScanResult{}, err
		} else if s.Packages[pkgPath], err = scanner.LoadPackage(pkgs[0]); err != nil {
			return scanner.ScanResult{}, err
		}
	}
	return s.Packages[pkgPath], nil
}

// mapType
func (s SchemaBuilder) mapType(t scanner.TypeID, seen map[scanner.TypeID]bool) error {
	scanResult, err := s.loadScanResult(t)
	if err != nil {
		return err
	}
	_ = scanResult

	if _, ok := seen[t]; ok {

	}

	return nil
}
