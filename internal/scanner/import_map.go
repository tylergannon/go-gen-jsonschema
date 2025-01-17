package scanner

import (
	"github.com/dave/dst"
	"path/filepath"
	"strings"
)

const (
	SchemaPackagePath   = "github.com/tylergannon/go-gen-jsonschema"
	schemaPackagePrefix = "jsonschema"
)

type ImportMap []*dst.ImportSpec

func (m ImportMap) GetPackageForPrefix(prefix string) (string, bool) {
	for _, spec := range m {
		path := strings.Trim(spec.Path.Value, "\"")
		if spec.Name == nil {
			if filepath.Base(path) == prefix {
				return path, true
			}
		} else {
			if spec.Name.Name == prefix {
				return path, true
			}
		}
	}
	return "", false
}

// GetGenJSONPrefix returns the prefix used for the go-gen-jsonschema package
// in the given import map, or else returns false for the second result
// if that package is not present in the file.
func (m ImportMap) GetGenJSONPrefix() (string, bool) {
	for _, _import := range m {
		if len(_import.Path.Value) == len(SchemaPackagePath)+2 && _import.Path.Value[1:len(SchemaPackagePath)+1] == SchemaPackagePath {
			if _import.Name == nil {
				return schemaPackagePrefix, true
			}
			return _import.Name.Name, true
		}
	}
	return "", false
}
