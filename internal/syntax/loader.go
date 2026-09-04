package syntax

import (
	"fmt"
	"go/token"
	"slices"
	"strconv"
	"strings"

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

// LoadReadonly loads a package with the jsonschema build tag while forbidding
// the Go command from updating go.mod or go.sum.
func LoadReadonly(path string) ([]*decorator.Package, error) {
	config := *DefaultPackageCfg
	config.Dir = path
	config.BuildFlags = append(slices.Clone(DefaultPackageCfg.BuildFlags), "-mod=readonly")
	loaded, err := decorator.Load(&config, ".")
	if err != nil {
		return nil, err
	}
	var loadErrors []packages.Error
	for _, pkg := range loaded {
		loadErrors = append(loadErrors, pkg.Errors...)
	}
	if len(loadErrors) > 0 {
		return nil, &PackageLoadError{Errors: loadErrors}
	}
	return loaded, nil
}

type PackageLoadError struct {
	Errors []packages.Error
}

func (e *PackageLoadError) Error() string {
	messages := make([]string, 0, len(e.Errors))
	for _, loadErr := range e.Errors {
		if loadErr.Pos == "" {
			messages = append(messages, loadErr.Msg)
		} else {
			messages = append(messages, fmt.Sprintf("%s: %s", loadErr.Pos, loadErr.Msg))
		}
	}
	return strings.Join(messages, "; ")
}

func (e *PackageLoadError) HasSourceError() bool {
	for _, loadErr := range e.Errors {
		if loadErr.Kind == packages.ParseError || loadErr.Kind == packages.TypeError {
			return true
		}
	}
	return false
}

func (e *PackageLoadError) Position() token.Position {
	for _, loadErr := range e.Errors {
		if position, ok := parsePackagePosition(loadErr.Pos); ok {
			return position
		}
	}
	return token.Position{}
}

func parsePackagePosition(value string) (token.Position, bool) {
	columnSeparator := strings.LastIndex(value, ":")
	if columnSeparator < 0 {
		return token.Position{}, false
	}
	lineSeparator := strings.LastIndex(value[:columnSeparator], ":")
	if lineSeparator < 0 {
		return token.Position{}, false
	}
	line, lineErr := strconv.Atoi(value[lineSeparator+1 : columnSeparator])
	column, columnErr := strconv.Atoi(value[columnSeparator+1:])
	if lineErr != nil || columnErr != nil {
		return token.Position{}, false
	}
	return token.Position{Filename: value[:lineSeparator], Line: line, Column: column}, true
}
