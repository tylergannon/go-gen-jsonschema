package builder

import (
	"errors"
	"fmt"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/tylergannon/go-gen-jsonschema/internal/scanner"
	"go/token"
	"slices"
	"strings"
)

const maxNestingDepth = 5

type seenTypes []scanner.TypeID

func (s seenTypes) Seen(t scanner.TypeID) bool {
	return slices.Contains(s, t.Concrete())
}

func (s seenTypes) See(t scanner.TypeID) seenTypes {
	return append(seenTypes{t.Concrete()}, s...)
}

func New(pkg *decorator.Package) (SchemaBuilder, error) {
	var builder = SchemaBuilder{
		LocalPkg: pkg,
		Packages: map[string]scanner.ScanResult{},
		Schemas:  map[scanner.TypeID]JSONSchema{},
	}
	data, err := scanner.LoadPackage(pkg)
	if err != nil {
		return builder, err
	}
	for _, m := range data.SchemaMethods {
		if err = builder.mapType(m.Receiver, seenTypes{}); err != nil {
			return builder, err
		}
	}
	for _, f := range data.SchemaFuncs {
		if err = builder.mapType(f.Receiver, seenTypes{}); err != nil {
			return builder, err
		}
	}

	return builder, nil
}

type SchemaBuilder struct {
	LocalPkg *decorator.Package
	Packages map[string]scanner.ScanResult
	Schemas  map[scanner.TypeID]JSONSchema
}

// loadScanResult gets the scan result associated with the given scanner.TypeID
func (s SchemaBuilder) loadScanResult(t scanner.TypeID) (scanner.ScanResult, error) {
	var pkgPath = t.PkgPath
	if pkgPath == "" {
		pkgPath = s.LocalPkg.PkgPath
	}
	if _, ok := s.Packages[pkgPath]; !ok {
		if pkgs, err := scanner.Load(pkgPath); err != nil {
			return scanner.ScanResult{}, err
		} else if s.Packages[pkgPath], err = scanner.LoadPackage(pkgs[0]); err != nil {
			return scanner.ScanResult{}, err
		}
	}
	return s.Packages[pkgPath], nil
}

func (s SchemaBuilder) find(t scanner.TypeID) (token.Position, error) {
	sb, err := s.loadScanResult(t)
	if err != nil {
		return token.Position{}, err
	}
	typeSpec, ok := sb.LocalNamedTypes[t.TypeName]
	if !ok {
		return token.Position{}, fmt.Errorf("type %s not found", t.TypeName)
	}
	return scanner.NodePosition(sb.Pkg, typeSpec.TypeSpec), nil
}

func (s SchemaBuilder) mapInterface(iface scanner.IfaceImplementations, seen seenTypes) error {
	if seen.Seen(iface.TypeID) {
		return fmt.Errorf("circular dependency found for type %s, defined at %s", iface.TypeID, iface.Position)
	}
	seen = seen.See(iface.TypeID)
	if err := s.checkSeen(seen); err != nil {
		return err
	}

	node := UnionTypeNode{
		TypeID_: iface.TypeID,
	}
	for _, opt := range iface.Impls {
		if err := s.mapType(opt, seen); err != nil {
			return err
		}
		optSchema := s.Schemas[opt]
		obj, ok := optSchema.(ObjectNode)
		if !ok {
			pos, err := s.find(obj.TypeID_)
			if err != nil {
				return err
			} else {
				return fmt.Errorf("expected %s to be an object-type schema at %s", obj.TypeID_.TypeName, pos)
			}
		}
		node.Options = append(node.Options, obj)
	}
	s.Schemas[iface.TypeID] = node
	return nil
}

func (s SchemaBuilder) mapEnumType(enum *scanner.EnumSet, seen seenTypes) error {
	seen = seen.See(enum.TypeID)
	if err := s.checkSeen(seen); err != nil {
		return err
	}

	propType := PropertyNode[string]{
		TypeID_: enum.TypeID,
		Typ:     "string",
	}
	for _, opt := range enum.Values {
		propType.Enum = append(propType.Enum, strings.Trim(opt.Decl.Values[0].(*dst.BasicLit).Value, "\""))
	}
	s.Schemas[enum.TypeID] = propType
	return nil
}

// mapType
func (s SchemaBuilder) mapType(t scanner.TypeID, seen seenTypes) error {
	scanResult, err := s.loadScanResult(t)
	if err != nil {
		return err
	}
	if iface, ok := scanResult.Interfaces[t.TypeName]; ok {
		if err = s.mapInterface(iface, seen); err != nil {
			return err
		}
	} else if enum, ok := scanResult.Constants[t.TypeName]; ok {
		if err = s.mapEnumType(enum, seen); err != nil {
			return err
		}
	} else if err = s.mapNamedType(t, seen); err != nil {
		return err
	}

	return nil
}

func (s SchemaBuilder) checkSeen(seen seenTypes) error {
	if len(seen) > maxNestingDepth {
		pos, _ := s.find(seen[0])
		return fmt.Errorf("max nesting depth exceeded at %s", pos)
	}
	return nil
}

func (s SchemaBuilder) mapNamedType(t scanner.TypeID, seen seenTypes) error {
	scanResult, err := s.loadScanResult(t)
	if err != nil {
		return err
	}
	typeSpec, ok := scanResult.LocalNamedTypes[t.TypeName]
	if !ok {
		return fmt.Errorf("type %s not found", t.TypeName)
	}
	if seen.Seen(t) {
		return fmt.Errorf("circular dependency found for type %s at %s", t.TypeName, typeSpec.Position())
	}
	switch node := typeSpec.TypeSpec.Type.(type) {
	default:

		fmt.Printf("Unrecognized node %T %#v type %s at %s\n", node, node, t.TypeName, typeSpec.Position())
		return errors.New("unhandled node type")
	}
	//return nil
}
