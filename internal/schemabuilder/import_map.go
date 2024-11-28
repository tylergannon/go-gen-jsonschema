package schemabuilder

import "github.com/dave/dst"

type ImportMap map[string]string

func NewImportMap(localPackage string, imports []*dst.ImportSpec) ImportMap {
	importMap := make(ImportMap)
	for _, importSpec := range imports {
		importMap[importSpec.Name.Name] = importSpec.Path.Value
	}
	importMap[""] = localPackage
	return importMap
}
