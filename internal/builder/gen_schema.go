package builder

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/tylergannon/go-gen-jsonschema/internal/scanner"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const maxNestingDepth = 100 // This is not the JSON Schema nesting depth but recursion depth...
const defaultSubdir = "jsonschema"

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
		Subdir:   defaultSubdir,
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
	Subdir   string
	Pretty   bool
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
	if schema, err := s.renderSchema(t, typeSpec.AnyTypeSpec, typeSpec.GetDescription(), seen); err != nil {
		return err
	} else {
		s.Schemas[t.Concrete()] = schema
	}
	return nil
}

func (s SchemaBuilder) renderSchema(typeID scanner.TypeID, anyTypeSpec scanner.AnyTypeSpec, description string, seen seenTypes) (JSONSchema, error) {
	switch node := anyTypeSpec.Spec.(type) {
	case *dst.Ident:
		switch node.Name {
		case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
			return PropertyNode[int]{Desc: description, Typ: "integer", TypeID_: typeID}, nil
		case "string":
			return PropertyNode[string]{Desc: description, Typ: "string", TypeID_: typeID}, nil
		case "bool":
			return PropertyNode[bool]{Desc: description, Typ: "boolean", TypeID_: typeID}, nil
		case "float32", "float64":
			return PropertyNode[float64]{Desc: description, Typ: "number", TypeID_: typeID}, nil
		default:
			// Means it is another named type.
			// Find it.
			newType := scanner.TypeID{TypeName: node.Name, PkgPath: node.Path}
			if newType.PkgPath == "" {
				newType.PkgPath = typeID.PkgPath
			}
			if err := s.mapType(newType, seen.See(typeID)); err != nil {
				return nil, err
			}
			if schema, ok := s.Schemas[newType]; !ok {
				panic("mapType apparently didn't map the type! " + newType.String())
			} else {
				if description == "" {
					return schema, nil
				}
				if _schemaNode, ok := schema.(schemaNode); !ok {
					return schema, nil
				} else {
					_schemaNode.SetDescription(description)
					return schema, nil
				}
			}
		}
		return nil, errors.New("Please finish me")
	case *dst.StarExpr:
		return s.renderSchema(typeID, anyTypeSpec.Derive(node.X), description, seen)
	case *dst.ArrayType:
		var (
			err    error
			schema = ArrayNode{
				Desc:    description,
				TypeID_: typeID,
			}
		)
		if schema.Items, err = s.renderSchema(typeID, anyTypeSpec.Derive(node.Elt), description, seen); err != nil {
			return nil, err
		}
		return schema, nil
	default:
		fmt.Printf("Node mapper found unrecognized node type %s (%T) at %s\n", anyTypeSpec.Spec, anyTypeSpec.Spec, anyTypeSpec.Position())
		return nil, errors.New("unhandled node type")
	}

}

func (s SchemaBuilder) localScan() scanner.ScanResult {
	return s.Packages[s.LocalPkg.PkgPath]
}

func (s SchemaBuilder) writeSchema(t scanner.TypeID, targetDir string) (err error) {
	var (
		file     *os.File
		ok       bool
		filePath = filepath.Join(targetDir, fmt.Sprintf("%s.json", t.TypeName))
	)
	if file, err = os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644); err != nil {
		return fmt.Errorf("could not open file %s: %w", filePath, err)
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	var schema json.Marshaler
	if schema, ok = s.Schemas[t.Concrete()]; !ok {
		return fmt.Errorf("unknown type %s", t.TypeName)
	}
	if s.Pretty {
		encoder.SetIndent("", "  ")
	}
	if err = encoder.Encode(schema); err != nil {
		return fmt.Errorf("could not encode schema: %w", err)
	}
	return nil
}

func (s SchemaBuilder) RenderSchemas() (err error) {
	var (
		localScan = s.localScan()
		targetDir = filepath.Join(s.LocalPkg.Dir, s.Subdir)
	)
	if err = os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("could not create subdir %s: %w", targetDir, err)
	}
	for _, method := range localScan.SchemaMethods {
		if err = s.writeSchema(method.Receiver, targetDir); err != nil {
			return err
		}
	}
	return nil
}
