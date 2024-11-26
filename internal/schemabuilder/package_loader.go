package schemabuilder

import "golang.org/x/tools/go/packages"

var PackageLoadNeeds = packages.NeedDeps |
	packages.NeedImports |
	packages.NeedModule |
	packages.NeedName |
	packages.NeedSyntax |
	packages.NeedTypes |
	packages.NeedTypesInfo

var DefaultPackageCfg = &packages.Config{
	Mode:  PackageLoadNeeds,
	Tests: false,
}
