package typeregistry

import (
	"github.com/dave/dst"
	"path"
	"strings"
)

type ImportMap map[string]string

func NewImportMap(localPackage string, imports []*dst.ImportSpec) ImportMap {
	importMap := make(ImportMap)
	for _, importSpec := range imports {
		var (
			pkgPath = deQuote(importSpec.Path.Value)
			name    = path.Base(pkgPath)
		)
		if importSpec.Name != nil {
			if importSpec.Name.Name == "." {
				panic("dot imports not supported")
			}
			name = importSpec.Name.Name
		}
		importMap[name] = pkgPath
	}
	importMap[""] = localPackage
	return importMap
}

func deQuote(s string) string {
	s = strings.TrimSuffix(s, "\"")
	s = strings.TrimPrefix(s, "\"")
	return s
}
