package loader

import (
	"github.com/dave/dst/decorator"
	"golang.org/x/tools/go/packages"
)

var PackageLoadNeeds = packages.NeedDeps |
	packages.NeedImports |
	packages.NeedModule |
	packages.NeedName |
	packages.NeedSyntax |
	packages.NeedTypes |
	packages.NeedTypesInfo |
	packages.NeedCompiledGoFiles |
	packages.NeedFiles

var DefaultPackageCfg = &packages.Config{
	Mode:  PackageLoadNeeds,
	Tests: false,
}

func Load(path string) ([]*decorator.Package, error) {
	return decorator.Load(DefaultPackageCfg, path)
}
