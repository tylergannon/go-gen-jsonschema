package builder

import (
	"fmt"

	"github.com/dave/dst/decorator"
)

// ImportMap helps with code generation by storing a list of packages
// along with aliases.  The alias can be looked up using the package object,
// which is stored alongside our AST nodes in QuestImpl and TaskImpl objects.
type ImportMap struct {
	localPackage *decorator.Package
	aliasCount   int
	types        []struct {
		alias string
		pkg   *decorator.Package
	}
}

func (m *ImportMap) LocalPkgName() string {
	return m.localPackage.Name
}

func NewImportMap(localPackage *decorator.Package) *ImportMap {
	return &ImportMap{localPackage: localPackage}
}

// AddPackage inserts the package into the map, unless it is already contained
// there.  Adds an alias if the alias has already been found.  Keeps a simple
// counter for creating very simple aliases.
func (m *ImportMap) AddPackage(pkg *decorator.Package) {
	if m.localPackage.ID == pkg.ID {
		return
	}
	haveName := false

	newObj := struct {
		alias string
		pkg   *decorator.Package
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
func (m *ImportMap) PrefixExpr(expr string, pkg *decorator.Package) string {
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

func (m *ImportMap) Alias(pkg *decorator.Package) string {
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
