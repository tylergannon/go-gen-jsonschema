package schemabuilder

import (
	"fmt"
	"golang.org/x/tools/go/packages"
	"slices"
)

type importMapTuple struct {
	alias string
	pkg   *packages.Package
}

// ImportMap helps with code generation by storing a list of packages
// along with aliases.  The alias can be looked up using the package object,
// which is stored alongside our AST nodes in QuestImpl and TaskImpl objects.
type ImportMap struct {
	localPackage *packages.Package
	aliasCount   int
	types        []importMapTuple
}

func NewImportMap(localPackage *packages.Package) *ImportMap {
	return &ImportMap{localPackage: localPackage}
}

func (m *ImportMap) FindPackageContainingDeclaration(name string) *packages.Package {
	for _, tuple := range m.types {
		pkg := tuple.pkg
		for _, info := range pkg.TypesInfo.Types {

		}
	}
}

func (m *ImportMap) GetPackage(path string) *packages.Package {
	if m.localPackage.PkgPath == path {
		return m.localPackage
	}
	idx := slices.IndexFunc(m.types, func(tuple importMapTuple) bool {
		return tuple.pkg.PkgPath == path
	})
	if idx < 0 {
		return nil
	}
	return m.types[idx].pkg
}

// AddPackage inserts the package into the map, unless it is already contained
// there.  Adds an alias if the alias has already been found.  Keeps a simple
// counter for creating very simple aliases.
func (m *ImportMap) AddPackage(pkg *packages.Package) {
	if m.localPackage.ID == pkg.ID {
		return
	}
	haveName := false

	newObj := struct {
		alias string
		pkg   *packages.Package
	}{pkg: pkg}

	for _, t := range m.types {
		if t.pkg.ID == pkg.ID {
			return
		}
		if t.pkg.Name == pkg.Name {
			haveName = true
		}
	}
	if haveName {
		m.aliasCount++
		newObj.alias = fmt.Sprintf("%s%d", pkg.Name, m.aliasCount)
	}
	m.types = append(m.types, newObj)
}

// PrefixExpr is a function that should be added to the template funcs when
// building a template object.  It correctly prints a type name or call
// expression using the right package name prefix/alias (or none if the
// expression refers to an identifier defined in the local package).
func (m *ImportMap) PrefixExpr(expr string, pkg *packages.Package) string {
	if pkg.ID == m.localPackage.ID {
		return expr
	}
	return fmt.Sprintf("%s.%s", m.Alias(pkg), expr)
}

func (m *ImportMap) ImportStatements() []string {
	var result []string
	for _, t := range m.types {
		// Note that we'll use `goimports` on this file later so imports will be
		// cleaned up and ordered.  Don't worry about the extra whitespace here.
		result = append(result, fmt.Sprintf("%s \"%s\"", t.pkg.Name, t.pkg.PkgPath))
	}

	return result
}

func (m *ImportMap) Alias(pkg *packages.Package) string {
	for _, t := range m.types {
		if t.pkg.ID == pkg.ID {
			if t.alias == "" {
				return pkg.Name
			}
			return t.alias
		}
	}
	panic("called Alias with package that's not registered")
}
