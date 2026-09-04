package syntax

import (
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// JSONMethod identifies a handwritten production JSON codec method. Generated
// output and generation-only declaration stubs are deliberately excluded.
type JSONMethod struct {
	Receiver string
	Name     string
	Position token.Position
}

// FindProductionJSONMethods finds active production MarshalJSON and
// UnmarshalJSON declarations for the requested local receiver types.
func FindProductionJSONMethods(dir string, receivers []string) ([]JSONMethod, error) {
	wanted := make(map[string]bool, len(receivers))
	for _, receiver := range receivers {
		wanted[receiver] = true
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	fset := token.NewFileSet()
	var methods []JSONMethod
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") || name == "jsonschema_gen.go" {
			continue
		}
		matches, err := IsProductionGoFile(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		if !matches {
			continue
		}
		file, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, 0)
		if err != nil {
			return nil, err
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || len(fn.Recv.List) != 1 || !slices.Contains([]string{"MarshalJSON", "UnmarshalJSON"}, fn.Name.Name) {
				continue
			}
			receiver, ok := receiverTypeName(fn.Recv.List[0].Type)
			if !ok || (len(wanted) > 0 && !wanted[receiver]) {
				continue
			}
			methods = append(methods, JSONMethod{
				Receiver: receiver,
				Name:     fn.Name.Name,
				Position: fset.Position(fn.Pos()),
			})
		}
	}
	return methods, nil
}

// IsProductionGoFile reports whether filename is active for the current
// production build. GOFLAGS tags are included so collision discovery follows
// the same custom-tag environment as the Go command that invoked generation.
func IsProductionGoFile(filename string) (bool, error) {
	context := build.Default
	context.BuildTags = productionBuildTags()
	return context.MatchFile(filepath.Dir(filename), filepath.Base(filename))
}

func goFlagBuildTags() []string {
	fields := strings.Fields(os.Getenv("GOFLAGS"))
	var tags []string
	for index := 0; index < len(fields); index++ {
		value := ""
		switch {
		case strings.HasPrefix(fields[index], "-tags="):
			value = strings.TrimPrefix(fields[index], "-tags=")
		case fields[index] == "-tags" && index+1 < len(fields):
			index++
			value = fields[index]
		}
		value = strings.Trim(value, `"'`)
		tags = append(tags, strings.FieldsFunc(value, func(r rune) bool { return r == ',' || r == ' ' })...)
	}
	slices.Sort(tags)
	return slices.Compact(tags)
}

func productionBuildTags() []string {
	tags := goFlagBuildTags()
	return slices.DeleteFunc(tags, func(tag string) bool { return tag == BuildTag })
}

func generationBuildTags() []string {
	tags := productionBuildTags()
	tags = append(tags, BuildTag)
	slices.Sort(tags)
	return slices.Compact(tags)
}

func receiverTypeName(expr ast.Expr) (string, bool) {
	switch value := expr.(type) {
	case *ast.Ident:
		return value.Name, true
	case *ast.StarExpr:
		return receiverTypeName(value.X)
	case *ast.ParenExpr:
		return receiverTypeName(value.X)
	case *ast.IndexExpr:
		return receiverTypeName(value.X)
	case *ast.IndexListExpr:
		return receiverTypeName(value.X)
	default:
		return "", false
	}
}
