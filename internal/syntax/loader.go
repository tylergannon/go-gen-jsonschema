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
	Mode:  PackageLoadNeeds,
	Tests: false,
}

func Load(path string) ([]*decorator.Package, error) {
	context, err := ResolveBuildContext()
	if err != nil {
		return nil, err
	}
	return LoadWithBuildContext(path, context)
}

// LoadWithBuildContext loads using a previously resolved operation context.
func LoadWithBuildContext(path string, context BuildContext) ([]*decorator.Package, error) {
	return loadFromDirWithBuildContext("", path, context, false)
}

// LoadReadonly loads a package with the jsonschema build tag while forbidding
// the Go command from updating go.mod or go.sum.
func LoadReadonly(path string) ([]*decorator.Package, error) {
	context, err := ResolveBuildContext()
	if err != nil {
		return nil, err
	}
	return LoadReadonlyWithBuildContext(path, context)
}

// LoadReadonlyWithBuildContext loads with a previously resolved operation
// context, avoiding a second Go environment lookup during inspection.
func LoadReadonlyWithBuildContext(path string, context BuildContext) ([]*decorator.Package, error) {
	return LoadTargetWithBuildContext(path, context, true)
}

// LoadTargetWithBuildContext loads and validates the package in targetDir under
// the operation's effective build context. readonly prevents module-file
// updates but does not change which source models are accepted.
func LoadTargetWithBuildContext(targetDir string, context BuildContext, readonly bool) ([]*decorator.Package, error) {
	loaded, err := loadFromDirWithBuildContext(targetDir, ".", context, readonly)
	if err != nil {
		return nil, err
	}
	if err := validateLoadedPackages(loaded); err != nil {
		return nil, err
	}
	return loaded, nil
}

func loadFromDirWithBuildContext(dir, pattern string, context BuildContext, readonly bool) ([]*decorator.Package, error) {
	config := packageConfig(context)
	config.Dir = dir
	if readonly {
		config.BuildFlags = append(config.BuildFlags, "-mod=readonly")
	}
	loaded, err := decorator.Load(&config, pattern)
	if err != nil {
		return nil, err
	}
	return loaded, nil
}

func validateLoadedPackages(loaded []*decorator.Package) error {
	var loadErrors []packages.Error
	missingImport := false
	seen := make(map[string]bool)
	var collectErrors func(*decorator.Package)
	collectErrors = func(pkg *decorator.Package) {
		if pkg == nil || seen[pkg.ID] {
			return
		}
		seen[pkg.ID] = true
		loadErrors = append(loadErrors, pkg.Errors...)
		for _, file := range pkg.Syntax {
			for _, imported := range file.Imports {
				path, unquoteErr := strconv.Unquote(imported.Path.Value)
				if unquoteErr == nil && path != "C" && pkg.Package.Imports[path] == nil {
					missingImport = true
				}
			}
		}
		for _, imported := range pkg.Imports {
			collectErrors(imported)
		}
	}
	for _, pkg := range loaded {
		collectErrors(pkg)
	}
	if len(loadErrors) > 0 {
		return &PackageLoadError{Errors: loadErrors, MissingImport: missingImport}
	}
	return nil
}

func packageConfig(context BuildContext) packages.Config {
	config := *DefaultPackageCfg
	config.BuildFlags = slices.Clone(DefaultPackageCfg.BuildFlags)
	config.BuildFlags = append(config.BuildFlags, context.generationBuildFlags()...)
	return config
}

type PackageLoadError struct {
	Errors        []packages.Error
	MissingImport bool
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

func (e *PackageLoadError) HasToolchainError() bool {
	if e.MissingImport {
		return true
	}
	for _, loadErr := range e.Errors {
		if loadErr.Kind == packages.ListError {
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
