package builder

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/tylergannon/go-gen-jsonschema/internal/common"
	"github.com/tylergannon/go-gen-jsonschema/internal/syntax"
)

//go:embed schemas.go.tmpl
var schemasTemplate string

const maxNestingDepth = 100 // This is not the JSON Schema nesting depth but recursion depth...
const defaultSubdir = "jsonschema"

func New(pkg *decorator.Package) (SchemaBuilder, error) {
	data, err := syntax.LoadPackage(pkg)
	if err != nil {
		return SchemaBuilder{}, err
	}
	var builder = SchemaBuilder{
		Scan:              data,
		schemas:           schemaMap{},
		customTypes:       map[string][]InterfaceProp{},
		Subdir:            defaultSubdir,
		BuildTag:          syntax.BuildTag,
		DiscriminatorProp: DefaultDiscriminatorPropName,
	}
	for _, m := range data.SchemaMethods {
		if err = builder.mapType(m.Receiver, syntax.SeenTypes{}); err != nil {
			return builder, err
		}
	}
	for _, f := range data.SchemaFuncs {
		if err = builder.mapType(f.Receiver, syntax.SeenTypes{}); err != nil {
			return builder, err
		}
	}

	return builder, nil
}

type CustomMarshaledType struct {
	Name           string
	InterfaceProps []InterfaceProp
	Star           string
	Initial        string
}

type InterfaceOptionInfo struct {
	TypeNameWithPrefix string
	Discriminator      string
	Pointer            bool
}

type InterfaceInfo struct {
	TypeNameWithPrefix string
	UnmarshalerFunc    string
	Options            []InterfaceOptionInfo
}

func (c *CustomMarshaledType) UnmarshalJSON(data []byte) (err error) {
	type Wrapper struct {
		*CustomMarshaledType
		Star    json.RawMessage
		Initial json.RawMessage
	}
	wrapper := Wrapper{
		CustomMarshaledType: c,
	}
	if err = json.Unmarshal(data, &wrapper); err != nil {
		return err
	}
	if err = json.Unmarshal(wrapper.Initial, &c.Initial); err != nil {
		return err
	}
	if err = json.Unmarshal(wrapper.Star, &c.Star); err != nil {
		return err
	}
	return nil
}

type SchemaBuilder struct {
	Scan              syntax.ScanResult
	schemas           schemaMap
	customTypes       map[string][]InterfaceProp
	Subdir            string
	Pretty            bool
	BuildTag          string
	Imports           []string
	SpecialTypes      []CustomMarshaledType
	Interfaces        []InterfaceInfo
	DiscriminatorProp string
}

func (s SchemaBuilder) HaveInterfaces() bool {
	return len(s.Interfaces) > 0
}

func (s SchemaBuilder) imports() *ImportMap {
	importMap := NewImportMap(s.Scan.Pkg)
	// For each type that has any special interface handling,
	// need a
	for _, interfaceProps := range s.customTypes {
		for _, prop := range interfaceProps {
			importMap.AddPackage(prop.Interface.TypeSpec.Pkg())
			for _, implType := range prop.Interface.Impls {
				if scan, ok := s.Scan.GetPackage(implType.PkgPath); !ok {
					panic("internal error: no package found for " + implType.PkgPath)
				} else {
					importMap.AddPackage(scan.Pkg)
				}
			}
		}
	}
	return importMap
}

func (s SchemaBuilder) SchemaMethods() []syntax.SchemaMethod {
	return s.Scan.SchemaMethods
}

type schemaMap map[string]map[string]JSONSchema

func (m schemaMap) Set(pkgPath, typeName string, schema JSONSchema) {
	if m[pkgPath] == nil {
		m[pkgPath] = make(map[string]JSONSchema)
	}
	m[pkgPath][typeName] = schema
}
func (m schemaMap) Get(pkgPath, typeName string) (schema JSONSchema, ok bool) {
	var _m map[string]JSONSchema
	if _m, ok = m[pkgPath]; !ok {
		return
	}
	schema, ok = _m[typeName]
	return
}

func (s SchemaBuilder) GetSchema(t syntax.TypeID) (schema JSONSchema, ok bool) {
	return s.schemas.Get(t.PkgPath, t.TypeName)
}

func (s SchemaBuilder) AddSchema(t syntax.TypeID, schema JSONSchema) {
	ty := t.Concrete()
	s.schemas.Set(ty.PkgPath, ty.TypeName, schema)
}

// loadScanResult gets the scan result associated with the given syntax.TypeID
func (s SchemaBuilder) loadScanResult(t syntax.TypeID) (syntax.ScanResult, error) {
	if t.PkgPath == "" {
		panic("empty package path in loadScanResult")
	}
	if res, ok := s.Scan.GetPackage(t.PkgPath); ok {
		return res, nil
	}
	panic("package was not loaded: " + t.PkgPath)
}

func (s SchemaBuilder) find(t syntax.TypeID) (token.Position, error) {
	sb, err := s.loadScanResult(t)
	if err != nil {
		return token.Position{}, err
	}
	typeSpec, ok := sb.LocalNamedTypes[t.TypeName]
	if !ok {
		return token.Position{}, fmt.Errorf("SchemaBuilder.find: type %s not found", t.TypeName)
	}
	return typeSpec.Position(), nil
}

func (s SchemaBuilder) mapInterface(iface syntax.IfaceImplementations, seen syntax.SeenTypes) error {
	if seen.Seen(iface.TypeSpec.ID()) {
		return fmt.Errorf("circular dependency found for type %s, defined at %s", iface.TypeSpec.ID(), iface.TypeSpec.Position())
	}
	seen = seen.See(iface.TypeSpec.ID())
	if err := s.checkSeen(seen); err != nil {
		return err
	}

	node := UnionTypeNode{
		TypeID_: iface.TypeSpec.ID(),
	}
	for _, opt := range iface.Impls {
		if err := s.mapType(opt, seen); err != nil {
			return err
		}
		optSchema, ok := s.GetSchema(opt)
		if !ok {
			return fmt.Errorf("type %s is not a known schema", opt)
		}
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
	s.AddSchema(iface.TypeSpec.ID(), node)
	return nil
}

func (s SchemaBuilder) mapEnumType(enum *syntax.EnumSet, seen syntax.SeenTypes) error {
	seen = seen.See(enum.TypeSpec.ID())
	if err := s.checkSeen(seen); err != nil {
		return err
	}

	propType := PropertyNode[string]{
		TypeID_: enum.TypeSpec.ID(),
		Typ:     "string",
	}
	var (
		sb            strings.Builder
		countComments int
	)

	for _, opt := range enum.Values {
		var (
			newValue = strings.Trim(opt.Value().Values[0].(*dst.BasicLit).Value, "\"")
			comment  = opt.Comments()
		)
		if len(comment) > 0 {
			if countComments > 0 {
				sb.WriteString("\n\n")
			}
			countComments++
			sb.WriteString(newValue)
			sb.WriteString(": \n")
			sb.WriteString(comment)
		}
		propType.Enum = append(propType.Enum, newValue)
	}
	if enum.TypeSpec.Pkg() == nil {
		panic("oh heck")
	}

	var comment = enum.TypeSpec.Comments()
	if len(comment) > 0 {
		if sb.Len() > 0 {
			propType.Desc = comment + "\n\n" + sb.String()
		} else {
			propType.Desc = comment
		}
	} else if sb.Len() > 0 {
		propType.Desc = sb.String()
	}
	s.AddSchema(enum.TypeSpec.ID(), propType)
	return nil
}

// mapType
func (s SchemaBuilder) mapType(t syntax.TypeID, seen syntax.SeenTypes) error {
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

func (s SchemaBuilder) checkSeen(seen syntax.SeenTypes) error {
	if len(seen) > maxNestingDepth {
		pos, _ := s.find(seen[0])
		return fmt.Errorf("max nesting depth exceeded at %s", pos)
	}
	return nil
}

func (s SchemaBuilder) mapNamedType(t syntax.TypeID, seen syntax.SeenTypes) error {
	scanResult, err := s.loadScanResult(t)
	if err != nil {
		return err
	}
	typeSpec, ok := scanResult.LocalNamedTypes[t.TypeName]
	if !ok {
		return fmt.Errorf("mapNamedType: type %s not found", t.TypeName)
	}
	if seen.Seen(t) {
		return fmt.Errorf("circular dependency found for type %s at %s", t.TypeName, typeSpec.Position())
	}
	if structType, ok := typeSpec.Type().Expr().(*dst.StructType); ok {
		if props, err := s.resolveLocalInterfaceProps(syntax.NewStructType(structType, typeSpec), nil); err != nil {
			return err
		} else if len(props) > 0 {
			s.customTypes[t.TypeName] = props
		}
	}
	if schema, err := s.renderSchema(typeSpec.Derive(), typeSpec.Comments(), seen); err != nil {
		return err
	} else {
		s.AddSchema(t, schema)
	}
	return nil
}

func (s SchemaBuilder) renderSchema(t syntax.TypeExpr, description string, seen syntax.SeenTypes) (JSONSchema, error) {
	switch node := t.Excerpt.(type) {
	case *dst.Ident:
		switch node.Name {
		case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
			return PropertyNode[int]{Desc: description, Typ: "integer", TypeID_: t.ID()}, nil
		case "string":
			return PropertyNode[string]{Desc: description, Typ: "string", TypeID_: t.ID()}, nil
		case "bool":
			return PropertyNode[bool]{Desc: description, Typ: "boolean", TypeID_: t.ID()}, nil
		case "float32", "float64":
			return PropertyNode[float64]{Desc: description, Typ: "number", TypeID_: t.ID()}, nil
		default:
			// Means it is another named type.
			// Find it.
			newType := syntax.TypeID{TypeName: node.Name, PkgPath: node.Path}
			if newType.PkgPath == "" {
				newType.PkgPath = t.Pkg().PkgPath
			}
			if err := s.mapType(newType, seen.See(t.ID())); err != nil {
				return nil, err
			}
			if schema, ok := s.GetSchema(newType); !ok {
				panic("mapType apparently didn't map the type! " + newType.String())
			} else {
				if description == "" {
					return schema, nil
				}
				if _schemaNode, ok := schema.(schemaNode); !ok {
					return schema, nil
				} else {
					return _schemaNode.setDescription(description), nil
				}
			}
		}
	case *dst.StarExpr:
		return s.renderSchema(t.Derive(node.X), description, seen)
	case *dst.ParenExpr:
		return s.renderSchema(t.Derive(node.X), description, seen)
	case *dst.ArrayType:
		var (
			err    error
			schema = ArrayNode{Desc: description, TypeID_: t.ID()}
		)
		if schema.Items, err = s.renderSchema(t.Derive(node.Elt), "", seen); err != nil {
			return nil, err
		}
		return schema, nil
	case *dst.MapType, *dst.ChanType:
		return nil, fmt.Errorf("unsupported type %s at %s", t.Name(), t.Position())
	case *dst.StructType:
		return s.renderStructSchema(syntax.NewStructType(node, *t.TypeSpec), description, seen)
	case *dst.InterfaceType:
		return nil, fmt.Errorf("Interface types are not supported. Found on %s at %s\n", t.ID(), t.Position())
	default:
		fmt.Printf("Node mapper found unrecognized node type %s at %s\n", t.ToExpr().Details(), t.ToExpr().Position())
		return nil, errors.New("unhandled node type")
	}
}

func (s SchemaBuilder) renderStructSchema(t syntax.StructType, description string, seen syntax.SeenTypes) (node ObjectNode, err error) {
	node = ObjectNode{
		Desc:          description,
		Discriminator: t.Name(),
		TypeID_:       t.ID(),
	}
	node.Properties, err = s.renderStructProps(t, nil, seen)
	return node, err
}

func (s SchemaBuilder) writeSchema(t syntax.TypeID, targetDir string) (err error) {
	var (
		file     *os.File
		ok       bool
		filePath = filepath.Join(targetDir, fmt.Sprintf("%s.json", t.TypeName))
	)
	if file, err = os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644); err != nil {
		return fmt.Errorf("could not open file %s: %w", filePath, err)
	}
	defer common.LogClose(file)
	encoder := json.NewEncoder(file)
	var schema json.Marshaler
	if schema, ok = s.GetSchema(t); !ok {
		return fmt.Errorf("unknown type %s", t)
	}
	if s.Pretty {
		encoder.SetIndent("", "  ")
	}
	if err = encoder.Encode(schema); err != nil {
		return fmt.Errorf("could not encode schema: %w", err)
	}
	return nil
}

func (s SchemaBuilder) RenderGoCode() (err error) {
	var interfaces = map[syntax.TypeID]InterfaceProp{}
	importMap := s.imports()
	s.Imports = importMap.ImportStatements()

	// for _, poop := range s.SchemaMethods() {
	// 	t := s.Scan.LocalNamedTypes[poop.Receiver.TypeName]
	// 	if st, ok := t.Type().Expr().(*dst.StructType); ok {
	// 		_st := syntax.NewStructType(st, t.Derive())
	// 		foo, err := _st.Flatten(
	// 			func(ident syntax.IdentExpr) (syntax.Expr, error) {
	// 				var (
	// 					newType syntax.TypeSpec
	// 					ok      bool
	// 				)
	// 				if ident.Concrete.Path == "" {
	// 					if newType, ok = s.Scan.LocalNamedTypes[ident.Concrete.Name]; !ok {
	// 						panic(fmt.Sprintf("unknown type %s", ident.Concrete.Name))
	// 					}
	// 				} else {
	// 					if scan, ok := s.Scan.GetPackage(ident.Concrete.Path); !ok {
	// 						panic(fmt.Sprintf("unknown type %s", ident.Concrete.Name))
	// 					}

	// 				}

	// 			},
	// 			nil,
	// 		)

	// 	}
	// }

	for n, itsProps := range s.customTypes {
		s.SpecialTypes = append(s.SpecialTypes, CustomMarshaledType{
			Name:           n,
			InterfaceProps: itsProps,
			Initial:        strings.ToLower(n[0:1]),
		})
		for _, prop := range itsProps {
			interfaces[prop.Interface.TypeSpec.ID()] = prop
		}
	}
	for _, ifaceProp := range interfaces {
		var (
			discriminators = map[string]bool{}
			ifacePkg       = ifaceProp.Interface.TypeSpec.Pkg()
		)
		var props []InterfaceOptionInfo
		for _, option := range ifaceProp.Interface.Impls {
			pkg, ok := s.Scan.GetPackage(option.PkgPath)
			if !ok {
				panic("could not find package at RenderGoCode: " + option.PkgPath)
			}
			var (
				discriminator = option.TypeName
				i             = 1
			)
			for discriminators[discriminator] {
				discriminator = strings.TrimSuffix(discriminator, strconv.Itoa(i-1))
				discriminator = fmt.Sprintf("%s%d", discriminator, i)
			}
			discriminators[discriminator] = true
			props = append(props, InterfaceOptionInfo{
				TypeNameWithPrefix: importMap.PrefixExpr(option.TypeName, pkg.Pkg),
				Discriminator:      discriminator,
				Pointer:            option.Indirection == syntax.Pointer,
			})
		}
		s.Interfaces = append(s.Interfaces, InterfaceInfo{
			TypeNameWithPrefix: importMap.PrefixExpr(ifaceProp.Interface.TypeSpec.Name(), ifacePkg),
			UnmarshalerFunc:    ifaceProp.UnmarshalerFunc(),
			Options:            props,
		})
	}
	data, err := RenderTemplate(schemasTemplate, s)
	if err != nil {
		return err
	}
	result, err := FormatCodeWithGoimports(data.Bytes())
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath.Join("jsonschema_gen.go"), result, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (s SchemaBuilder) RenderSchemas() (err error) {
	var targetDir = filepath.Join(s.Scan.Pkg.Dir, s.Subdir)

	if err = os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("could not create subdir %s: %w", targetDir, err)
	}
	for _, method := range s.Scan.SchemaMethods {
		if err = s.writeSchema(method.Receiver, targetDir); err != nil {
			return err
		}
	}
	return nil
}

func (s SchemaBuilder) resolveEmbeddedType(t syntax.TypeExpr, seen syntax.SeenTypes) (syntax.StructType, error) {
	switch expr := t.Excerpt.(type) {
	case *dst.Ident:
		if syntax.BasicTypes[expr.Name] {
			return syntax.NoStructType, fmt.Errorf("basic type %s is unsupported for embedding at %s", expr.Name, t.Position())
		}
		var pkgPath = expr.Path
		if pkgPath == "" {
			pkgPath = t.Pkg().PkgPath
		}
		if scan, ok := s.Scan.GetPackage(pkgPath); !ok {
			return syntax.NoStructType, fmt.Errorf("could not resolve package for type %s at %s", expr, t.Position())
		} else if ts, ok := scan.LocalNamedTypes[expr.Name]; !ok {
			return syntax.NoStructType, fmt.Errorf("could not resolve type %s at %s", expr, t.Position())
		} else {
			typeExpr := ts.Derive()
			switch _expr := typeExpr.Excerpt.(type) {
			case *dst.StructType:
				return syntax.NewStructType(_expr, *typeExpr.TypeSpec), nil
			case *dst.Ident:
				return s.resolveEmbeddedType(typeExpr, seen)
			}
			return syntax.NoStructType, fmt.Errorf("unsupported type %s at %s", ts.Details(), ts.Position())
		}

	case *dst.StarExpr:
		return s.resolveEmbeddedType(t.Derive(expr.X), seen)
	case *dst.ParenExpr:
		return s.resolveEmbeddedType(t.Derive(expr.X), seen)
	default:
		return syntax.NoStructType, fmt.Errorf("unsupported type %s", expr)
	}
}

func (s SchemaBuilder) renderStructProps(t syntax.StructType, seenProps syntax.SeenProps, seen syntax.SeenTypes) (props ObjectPropSet, err error) {
	var myProps = append(syntax.SeenProps{}, seenProps...)
	for _, prop := range t.Fields() {
		if prop.Skip() {
			continue
		}
		for _, name := range prop.PropNames() {
			myProps = myProps.See(name)
		}
	}
	for _, prop := range t.Fields() {
		var tempProps ObjectPropSet
		if prop.Skip() {
			continue
		}
		if prop.Embedded() {
			var embeddedType syntax.StructType
			if embeddedType, err = s.resolveEmbeddedType(t.Derive(), seen); err != nil {
				return nil, fmt.Errorf("resolving embedded type: %w", err)
			} else if tempProps, err = s.renderStructProps(embeddedType, myProps, seen); err != nil {
				return nil, fmt.Errorf("rendering embedded type: %w", err)
			}
		} else if tempProps, err = s.renderStructField(prop, seen); err != nil {
			return nil, fmt.Errorf("rendering struct field: %w", err)
		}
		props = append(props, tempProps...)
	}
	return props, nil
}

func (s SchemaBuilder) renderStructField(f syntax.StructField, seen syntax.SeenTypes) (props []ObjectProp, err error) {
	var (
		schema JSONSchema
		name   string
	)
	if schema, err = s.renderSchema(f.Derive(f.Type()), f.Comments(), seen); err != nil {
		return nil, fmt.Errorf("rendering schema: %w", err)
	}
	for _, name = range f.PropNames() {
		props = append(props, ObjectProp{
			Name:     name,
			Schema:   schema,
			Optional: false,
		})
	}
	return props, nil
}

type InterfaceProp struct {
	Field     syntax.StructField
	Interface syntax.IfaceImplementations
}

func (s InterfaceProp) UnmarshalerFunc() string {
	return fmt.Sprintf("__jsonUnmarshal__%s__%s", s.Interface.TypeSpec.Pkg().Name, s.Interface.TypeSpec.Name())
}

func (i InterfaceProp) FieldNames() string {
	var names []string
	for _, name := range i.Field.Field.Names {
		names = append(names, name.Name)
	}
	return strings.Join(names, ", ")
}

func (i InterfaceProp) StructTag() string {
	if i.Field.Field.Tag == nil {
		return ""
	}
	return i.Field.Field.Tag.Value
}

// resolveLocalInterfaceProps finds registered interface properties anywhere in the
// given struct field, as long as the struct type is from the local package.
// If the interface field is the very type of one of the struct's types (or of
// one of its embedded structs), it will be returned.
// If there are any registered interface types that are embedded deeper in the
// type expressions, they are considered illegal and will result in an error.
//
// Valid:
// ```
// type MyInterface interface{}
//
//	type struct Foo {
//	  Bar MyInterface `json:"bar"`
//	}
//
// ```
// Invalid examples:
// ```
// type (
//
//	MyInterface interface{}
//	struct Foo {
//	  Bar (MyInterface) `json:"bar"`
//	  Baz []MyInterface `json:"baz"`
//	  Bap struct { // Inline structs are permissible, but they cannot contain interfaces.
//	    Rap MyInterface `json:"rap"`
//	  }
//	}
//
// )
// ```
func (s SchemaBuilder) resolveLocalInterfaceProps(t syntax.StructType, seenProps syntax.SeenProps) (props []InterfaceProp, err error) {
	if t.Pkg().PkgPath != s.Scan.Pkg.PkgPath {
		return nil, nil
	}
	for _, prop := range t.Fields() {
		if prop.Embedded() {
			continue
		}
		if ident, ok := prop.Field.Type.(*dst.Ident); ok {
			ok = false
			for _, name := range prop.PropNames() {
				if !seenProps.Seen(name) {
					ok = true
				}
				seenProps = seenProps.See(name)
			}
			if !ok {
				continue
			}
			iface, ok := s.findInterfaceImpl(ident, s.Scan.Pkg)
			if !ok {
				continue
			}
			if len(prop.Field.Names) != 1 {
				return nil, fmt.Errorf("interface prop %s has more than one field name at %s", strings.Join(prop.PropNames(), ","), prop.Position())
			}
			props = append(props, InterfaceProp{
				Field:     prop,
				Interface: iface,
			})
			continue
		}
		dst.Inspect(prop.Field.Type, func(node dst.Node) bool {
			if ident, ok := node.(*dst.Ident); !ok {
				return true
			} else if _, ok = s.findInterfaceImpl(ident, s.Scan.Pkg); ok {
				pos := prop.Derive(ident).Position()
				err = fmt.Errorf("found registered interface type %s in an unsupported location at %s", ident.Name, pos)
				return false
			}
			return true
		})
		if err != nil {
			return nil, err
		}
	}
	for _, prop := range t.Fields() {
		if !prop.Embedded() {
			continue
		}
		if _t, err := s.resolveEmbeddedType(t.Derive(), nil); err != nil {
			return nil, fmt.Errorf("resolving embedded type: %w", err)
		} else if propsTemp, err := s.resolveLocalInterfaceProps(_t, seenProps); err != nil {
			return nil, fmt.Errorf("resolving embedded local interface properties: %w", err)
		} else {
			props = append(props, propsTemp...)
		}
	}
	return props, nil
}

func (s SchemaBuilder) findInterfaceImpl(ident *dst.Ident, localPkg *decorator.Package) (iface syntax.IfaceImplementations, ok bool) {
	var pkgPath = ident.Path
	if pkgPath == "" {
		pkgPath = localPkg.PkgPath
	}
	scan, ok := s.Scan.GetPackage(pkgPath)
	if !ok {
		panic(fmt.Sprintf("resolveLocalInterfaceProps could not resolve package for type %s at %s", ident.Name, pkgPath))
	}
	iface, ok = scan.Interfaces[ident.Name]
	return iface, ok
}
