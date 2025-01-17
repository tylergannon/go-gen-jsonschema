package syntax

import (
	"github.com/dave/dst/decorator"
	"golang.org/x/tools/go/packages"
)

const BuildTag = "jsonschema"

const PackageLoadNeeds = packages.NeedDeps |
	packages.NeedModule |
	packages.NeedName |
	packages.NeedSyntax |
	packages.NeedTypes |
	packages.NeedTypesInfo |
	packages.NeedCompiledGoFiles |
	packages.NeedFiles

var DefaultPackageCfg = &packages.Config{
	Mode:       PackageLoadNeeds,
	Tests:      false,
	BuildFlags: []string{"-tags=" + BuildTag},
}

func Load(path string) ([]*decorator.Package, error) {
	return decorator.Load(DefaultPackageCfg, path)
}
