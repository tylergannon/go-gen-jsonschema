package scanner

import "golang.org/x/tools/go/packages"

var PackageLoadNeeds = packages.NeedDeps |
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

func Load(path string) ([]*packages.Package, error) {
	return packages.Load(DefaultPackageCfg, path)
}
