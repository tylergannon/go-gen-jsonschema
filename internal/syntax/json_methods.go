package syntax

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
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

// BuildContext is the effective Go build environment used by one generator or
// inspection operation. It keeps generation and production file selection in
// sync without consulting the Go environment once per source file.
type BuildContext struct {
	production     build.Context
	generationTags []string
}

type goEnvironment struct {
	GOFLAGS    string
	GOOS       string
	GOARCH     string
	CGOEnabled string `json:"CGO_ENABLED"`
}

// ResolveBuildContext reads the same process and saved GOENV settings used by
// the Go command. The command is read-only and does not modify global config.
func ResolveBuildContext() (BuildContext, error) {
	return resolveBuildContext(os.Environ())
}

func resolveBuildContext(environment []string) (BuildContext, error) {
	command := exec.Command("go", "env", "-json", "GOFLAGS", "GOOS", "GOARCH", "CGO_ENABLED")
	command.Env = environment
	output, err := command.Output()
	if err != nil {
		return BuildContext{}, fmt.Errorf("read effective Go build environment: %w", err)
	}
	var effective goEnvironment
	if err := json.Unmarshal(output, &effective); err != nil {
		return BuildContext{}, fmt.Errorf("decode effective Go build environment: %w", err)
	}
	tags, err := parseGOFLAGSTags(effective.GOFLAGS)
	if err != nil {
		return BuildContext{}, err
	}
	production := build.Default
	production.GOOS = effective.GOOS
	production.GOARCH = effective.GOARCH
	production.CgoEnabled = effective.CGOEnabled == "1"
	production.BuildTags = slices.DeleteFunc(slices.Clone(tags), func(tag string) bool { return tag == BuildTag })
	generationTags := append(slices.Clone(production.BuildTags), BuildTag)
	slices.Sort(generationTags)
	return BuildContext{
		production:     production,
		generationTags: slices.Compact(generationTags),
	}, nil
}

// FindProductionJSONMethods finds active production MarshalJSON and
// UnmarshalJSON declarations for the requested local receiver types.
func FindProductionJSONMethods(dir string, receivers []string) ([]JSONMethod, error) {
	context, err := ResolveBuildContext()
	if err != nil {
		return nil, err
	}
	return context.FindProductionJSONMethods(dir, receivers)
}

// FindProductionJSONMethods scans using this operation's resolved build
// context, excluding the reserved generation tag.
func (c BuildContext) FindProductionJSONMethods(dir string, receivers []string) ([]JSONMethod, error) {
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
		matches, err := c.IsProductionGoFile(filepath.Join(dir, name))
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

// IsProductionGoFile reports whether filename is active for the effective
// production build, including custom tags saved in GOENV or set in GOFLAGS.
func IsProductionGoFile(filename string) (bool, error) {
	context, err := ResolveBuildContext()
	if err != nil {
		return false, err
	}
	return context.IsProductionGoFile(filename)
}

// IsProductionGoFile matches a file using this operation's effective Go
// platform, cgo setting, and custom tags without the reserved generation tag.
func (c BuildContext) IsProductionGoFile(filename string) (bool, error) {
	return c.production.MatchFile(filepath.Dir(filename), filepath.Base(filename))
}

func (c BuildContext) generationBuildFlags() []string {
	if len(c.generationTags) == 0 {
		return nil
	}
	return []string{"-tags=" + strings.Join(c.generationTags, ",")}
}

func parseGOFLAGSTags(value string) ([]string, error) {
	fields, err := splitQuotedFields(value)
	if err != nil {
		return nil, fmt.Errorf("parse effective GOFLAGS: %w", err)
	}
	var tags []string
	for _, field := range fields {
		if !strings.HasPrefix(field, "-") || field == "-" || field == "--" {
			return nil, fmt.Errorf("parse effective GOFLAGS: non-flag %q", field)
		}
		tagValue := ""
		switch {
		case strings.HasPrefix(field, "-tags="):
			tagValue = strings.TrimPrefix(field, "-tags=")
		case strings.HasPrefix(field, "--tags="):
			tagValue = strings.TrimPrefix(field, "--tags=")
		}
		tags = append(tags, strings.FieldsFunc(tagValue, func(r rune) bool { return r == ',' || r == ' ' })...)
	}
	slices.Sort(tags)
	return slices.Compact(tags), nil
}

// splitQuotedFields follows the Go command's GOFLAGS quoting rules: quotes are
// recognized only when they surround a complete whitespace-delimited token.
func splitQuotedFields(value string) ([]string, error) {
	var fields []string
	for len(value) > 0 {
		value = strings.TrimLeft(value, " \t\n\r")
		if value == "" {
			break
		}
		if value[0] == '\'' || value[0] == '"' {
			quote := value[0]
			value = value[1:]
			end := strings.IndexByte(value, quote)
			if end < 0 {
				return nil, fmt.Errorf("unterminated %c string", quote)
			}
			fields = append(fields, value[:end])
			value = value[end+1:]
			continue
		}
		end := strings.IndexAny(value, " \t\n\r")
		if end < 0 {
			fields = append(fields, value)
			break
		}
		fields = append(fields, value[:end])
		value = value[end:]
	}
	return fields, nil
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
