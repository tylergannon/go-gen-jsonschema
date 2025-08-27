package builder

import (
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"

	"hash/fnv"

	"slices"

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
		TypeProvidersMap:  map[string][]FieldProvider{},
		IfaceV1: make(map[string]map[string]struct {
			Impls []syntax.TypeID
			Disc  string
		}),
		EnumV1: make(map[string]map[string]struct {
			Mode  string
			Names map[string]string
		}),
		RenderedTypes: []string{},
		Rendered:      map[string]bool{},
	}
	// First, collect providers so they're available during mapping
	var foundNewInterfaceOpts bool
	collectOpts := func(recvName string, opts []syntax.SchemaMethodOptionInfo) {
		if len(opts) == 0 {
			return
		}
		for _, opt := range opts {
			switch string(opt.Kind) {
			case "WithRenderProviders":
				builder.RenderedTypes = append(builder.RenderedTypes, recvName)
				builder.Rendered[recvName] = true
				continue
			case "WithInterface", "WithInterfaceImpls", "WithDiscriminator":
				foundNewInterfaceOpts = true
				continue
			case "WithEnum", "WithEnumMode", "WithEnumName":
				// Enum options don't create providers, they're handled inline
				continue
			}
			builder.TypeProvidersMap[recvName] = append(builder.TypeProvidersMap[recvName], FieldProvider{
				FieldName:        opt.FieldName,
				Kind:             string(opt.Kind),
				ProviderName:     opt.ProviderName,
				ProviderIsMethod: opt.ProviderIsMethod,
			})
		}
	}
	for _, m := range data.SchemaMethods {
		collectOpts(m.Receiver.TypeName, m.Options)
	}
	for _, f := range data.SchemaFuncs {
		collectOpts(f.Receiver.TypeName, f.Options)
	}
	// Disallow mixing legacy NewInterfaceImpl with new interface options in same package
	if foundNewInterfaceOpts && len(data.Interfaces) > 0 {
		return builder, fmt.Errorf("invalid configuration: cannot mix legacy NewInterfaceImpl with v1 interface options in package %s", data.Pkg.PkgPath)
	}

	// Collect v1 interface options per receiver/field and enum options
	applyInterfaceOpts := func(recv string, opts []syntax.SchemaMethodOptionInfo) {
		for _, opt := range opts {
			switch string(opt.Kind) {
			case "WithInterface":
				if builder.IfaceV1[recv] == nil {
					builder.IfaceV1[recv] = make(map[string]struct {
						Impls []syntax.TypeID
						Disc  string
					})
				}
				if _, ok := builder.IfaceV1[recv][opt.FieldName]; !ok {
					builder.IfaceV1[recv][opt.FieldName] = struct {
						Impls []syntax.TypeID
						Disc  string
					}{Impls: nil, Disc: ""}
				}
			case "WithEnum":
				if builder.EnumV1[recv] == nil {
					builder.EnumV1[recv] = make(map[string]struct {
						Mode  string
						Names map[string]string
					})
				}
				if _, ok := builder.EnumV1[recv][opt.FieldName]; !ok {
					builder.EnumV1[recv][opt.FieldName] = struct {
						Mode  string
						Names map[string]string
					}{Mode: "", Names: map[string]string{}}
				}
			case "WithEnumMode":
				if builder.EnumV1[recv] == nil {
					builder.EnumV1[recv] = make(map[string]struct {
						Mode  string
						Names map[string]string
					})
				}
				cfg := builder.EnumV1[recv][opt.FieldName]
				cfg.Mode = opt.EnumMode
				builder.EnumV1[recv][opt.FieldName] = cfg
			case "WithEnumName":
				if builder.EnumV1[recv] == nil {
					builder.EnumV1[recv] = make(map[string]struct {
						Mode  string
						Names map[string]string
					})
				}
				cfg := builder.EnumV1[recv][opt.FieldName]
				if cfg.Names == nil {
					cfg.Names = map[string]string{}
				}
				if opt.EnumConst.TypeName != "" {
					cfg.Names[opt.EnumConst.TypeName] = opt.EnumName
				}
				builder.EnumV1[recv][opt.FieldName] = cfg
			case "WithDiscriminator":

				if builder.IfaceV1[recv] == nil {
					builder.IfaceV1[recv] = make(map[string]struct {
						Impls []syntax.TypeID
						Disc  string
					})
				}
				curr := builder.IfaceV1[recv][opt.FieldName]
				curr.Disc = opt.Discriminator
				builder.IfaceV1[recv][opt.FieldName] = curr
			case "WithInterfaceImpls":
				if builder.IfaceV1[recv] == nil {
					builder.IfaceV1[recv] = make(map[string]struct {
						Impls []syntax.TypeID
						Disc  string
					})
				}
				curr := builder.IfaceV1[recv][opt.FieldName]
				curr.Impls = slices.Clone(opt.ImplTypes)
				builder.IfaceV1[recv][opt.FieldName] = curr
			}
		}
	}
	for _, m := range data.SchemaMethods {
		applyInterfaceOpts(m.Receiver.TypeName, m.Options)
	}
	for _, f := range data.SchemaFuncs {
		applyInterfaceOpts(f.Receiver.TypeName, f.Options)
	}

	// Build TypeProviders slice for template convenience, computing JSON names for fields

	for typeName, providers := range builder.TypeProvidersMap {
		// compute json names from type spec
		if ts, ok := builder.Scan.LocalNamedTypes[typeName]; ok {
			if st, ok2 := ts.Type().Expr().(*dst.StructType); ok2 {
				stWrap := syntax.NewStructType(st, ts)
				for i := range providers {
					for _, f := range stWrap.Fields() {
						for _, name := range f.Field.Names {
							if name.Name == providers[i].FieldName {
								jsonNames := f.PropNames()
								if len(jsonNames) > 0 {
									providers[i].JSONName = jsonNames[0]
								}
							}
						}
					}
				}
			}
		}
		builder.TypeProviders = append(builder.TypeProviders, TypeProviders{TypeName: typeName, Providers: providers})
	}
	// Now map types
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
	TypeName           string
	PkgPath            string
	Pointer            bool
}

type FieldProvider struct {
	FieldName        string
	JSONName         string
	Kind             string
	ProviderName     string
	ProviderIsMethod bool
}

type TypeProviders struct {
	TypeName  string
	IsPointer bool
	Providers []FieldProvider
}

type InterfaceInfo struct {
	TypeNameWithPrefix    string
	TypeName              string
	PkgPath               string
	UnmarshalerFunc       string
	DiscriminatorPropName string
	Options               []InterfaceOptionInfo
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
	NumTestSamples    int
	BuildTag          string
	Imports           []string
	SpecialTypes      []CustomMarshaledType
	Interfaces        []InterfaceInfo
	DiscriminatorProp string

	// Field provider options per type (by receiver type name)
	TypeProvidersMap map[string][]FieldProvider
	TypeProviders    []TypeProviders

	// V1 interface options: receiver -> field -> config
	IfaceV1 map[string]map[string]struct {
		Impls []syntax.TypeID
		Disc  string
	}

	// Enum options: receiver -> field -> config
	EnumV1 map[string]map[string]struct {
		Mode  string            // "strings" or ""
		Names map[string]string // const ident -> name override
	}

	// Types requesting rendered provider execution
	RenderedTypes []string
	Rendered      map[string]bool
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
	// Merge methods and funcs, then filter out invalid receiver base types (underlying pointer/interface)
	var out []syntax.SchemaMethod
	appendIfValid := func(m syntax.SchemaMethod) {
		if ts, ok := s.Scan.LocalNamedTypes[m.Receiver.TypeName]; ok {
			switch ts.Type().Expr().(type) {
			case *dst.StarExpr, *dst.InterfaceType:
				return
			}
		}
		out = append(out, m)
	}
	for _, m := range s.Scan.SchemaMethods {
		appendIfValid(m)
	}
	for _, f := range s.Scan.SchemaFuncs {
		appendIfValid(syntax.SchemaMethod(f))
	}
	return out
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

	// Determine if this is a string or int based enum
	isIntEnum := false

	// Check if there are any values to examine
	if len(enum.Values) > 0 && len(enum.Values[0].Value().Values) > 0 {
		// Check if the first value is an iota or integer
		switch enum.Values[0].Value().Values[0].(type) {
		case *dst.Ident: // iota
			isIntEnum = true
		}
	}

	var (
		sb            strings.Builder
		countComments int
	)

	if isIntEnum {
		// Handle integer enum
		propType := PropertyNode[int]{
			TypeID_: enum.TypeSpec.ID(),
			Typ:     "integer",
		}

		for i, opt := range enum.Values {
			var (
				intValue int
				comment  = opt.Comments()
			)

			// For iota enums, use the index as the value
			intValue = i

			if len(comment) > 0 {
				if countComments > 0 {
					sb.WriteString("\n\n")
				}
				countComments++
				sb.WriteString(fmt.Sprintf("%d", intValue))
				sb.WriteString(": \n")
				sb.WriteString(comment)
			}
			propType.Enum = append(propType.Enum, intValue)
		}

		if len(enum.TypeSpec.Comments()) > 0 {
			propType.Desc = enum.TypeSpec.Comments()
		}
		s.AddSchema(enum.TypeSpec.ID(), propType)
	} else {
		// Handle string enum
		propType := PropertyNode[string]{
			TypeID_: enum.TypeSpec.ID(),
			Typ:     "string",
		}

		for i, opt := range enum.Values {
			var (
				newValue string
				comment  = opt.Comments()
			)

			// Handle different types of enum values
			if len(opt.Value().Values) > 0 {
				switch v := opt.Value().Values[0].(type) {
				case *dst.BasicLit:
					// String literal enum value
					newValue = strings.Trim(v.Value, "\"")
				default:
					// Shouldn't happen for string enums, but fallback to index
					newValue = fmt.Sprintf("%d", i)
				}
			} else {
				// No explicit value, shouldn't happen for string enums
				newValue = fmt.Sprintf("%d", i)
			}

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

		if len(enum.TypeSpec.Comments()) > 0 {
			propType.Desc = enum.TypeSpec.Comments()
		}
		s.AddSchema(enum.TypeSpec.ID(), propType)
	}

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
		return nil, fmt.Errorf("mapType/chanType not allowed %s at %s", t.Name(), t.Position())
	case *dst.StructType:
		return s.renderStructSchema(syntax.NewStructType(node, *t.TypeSpec), description, seen)
	case *dst.InterfaceType:
		return nil, fmt.Errorf("interface types are not supported. Found on %s at %s", t.ID(), t.Position())
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

func (s SchemaBuilder) writeSchema(t syntax.TypeID, targetDir string, noChanges bool) (wroteNew bool, err error) {
	var (
		ok       bool
		filePath string
		sumPath  string
		tmpFile  *os.File
	)

	// Decide target file extension based on whether schema is templated
	if _, templated := s.TypeProvidersMap[t.TypeName]; templated {
		filePath = filepath.Join(targetDir, fmt.Sprintf("%s.json.tmpl", t.TypeName))
	} else {
		filePath = filepath.Join(targetDir, fmt.Sprintf("%s.json", t.TypeName))
	}
	sumPath = filePath + ".sum"

	// Create temp file in same directory to ensure same filesystem

	if tmpFile, err = os.CreateTemp(targetDir, fmt.Sprintf("%s.*.json.tmp", t.TypeName)); err != nil {
		return false, fmt.Errorf("could not create temp file: %w", err)
	}
	defer func() {
		if fCloseErr := tmpFile.Close(); fCloseErr != nil && !errors.Is(fCloseErr, os.ErrClosed) {
			err = errors.Join(err, fmt.Errorf("could not close temp file: %w", fCloseErr))
		}
		// Clean up temp file if we're returning with an error or if we didn't use it
		_, statErr := os.Stat(tmpFile.Name())
		if os.IsNotExist(statErr) {
			return
		} else if statErr != nil {
			err = errors.Join(err, fmt.Errorf("could not stat temp file: %w", statErr))
			return
		}
		if rmErr := os.Remove(tmpFile.Name()); rmErr != nil && !errors.Is(rmErr, os.ErrNotExist) {
			err = errors.Join(err, fmt.Errorf("could not remove temp file: %w", rmErr))
		}
	}()

	var schema json.Marshaler
	if schema, ok = s.GetSchema(t); !ok {
		return false, fmt.Errorf("unknown type %s", t)
	}

	hash := fnv.New64a()
	writer := io.MultiWriter(tmpFile, hash)
	// If this type uses providers, the schema is a template; write raw without pretty/validation
	if _, templated := s.TypeProvidersMap[t.TypeName]; templated {
		var b []byte
		if b, err = schema.MarshalJSON(); err != nil {
			return false, fmt.Errorf("could not marshal template schema: %w", err)
		}
		if _, err = writer.Write(b); err != nil {
			return false, fmt.Errorf("could not write template schema: %w", err)
		}
		if _, err = writer.Write([]byte("\n")); err != nil {
			return false, fmt.Errorf("could not write newline: %w", err)
		}
	} else {
		encoder := json.NewEncoder(writer)
		if s.Pretty {
			encoder.SetIndent("", "  ")
		}
		if err = encoder.Encode(schema); err != nil {
			return false, fmt.Errorf("could not encode schema: %w", err)
		}
	}

	newChecksum := hex.EncodeToString(hash.Sum(nil))

	// Check if content actually changed by comparing with old checksum
	wroteNew = true
	if oldSum, err := os.ReadFile(sumPath); err == nil {
		wroteNew = string(oldSum) != newChecksum
	}

	// If content changed and we're in noChanges mode, return without writing anything
	if wroteNew && noChanges {
		return true, nil
	}

	// Move temp file into place and write new checksum
	if err = tmpFile.Close(); err != nil {
		return false, fmt.Errorf("could not close temp file: %w", err)
	}
	if err = os.Rename(tmpFile.Name(), filePath); err != nil {
		return false, fmt.Errorf("could not move temp file into place: %w", err)
	}
	if err = os.WriteFile(sumPath, []byte(newChecksum), 0644); err != nil {
		return false, fmt.Errorf("could not write checksum file: %w", err)
	}

	return wroteNew, nil
}

func (s *SchemaBuilder) RenderGoCode() (err error) {
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
		for _, ifaceProp := range itsProps {
			var (
				discriminators = map[string]bool{}
				ifacePkg       = ifaceProp.Interface.TypeSpec.Pkg()
			)
			var opts []InterfaceOptionInfo
			for _, option := range ifaceProp.Interface.Impls {
				pkg, ok := s.Scan.GetPackage(option.PkgPath)
				if !ok {
					panic("could not find package at RenderGoCode: " + option.PkgPath)
				}
				var (
					disc = option.TypeName
					i    = 1
				)
				for discriminators[disc] {
					disc = strings.TrimSuffix(disc, strconv.Itoa(i-1))
					disc = fmt.Sprintf("%s%d", disc, i)
				}
				discriminators[disc] = true
				opts = append(opts, InterfaceOptionInfo{
					TypeNameWithPrefix: importMap.PrefixExpr(option.TypeName, pkg.Pkg),
					Discriminator:      disc,
					TypeName:           option.TypeName,
					PkgPath:            option.PkgPath,
					Pointer:            option.Indirection == syntax.Pointer,
				})
			}
			// Determine discriminator property name for this field-specific unmarshaler (only if overridden)
			discProp := ifaceProp.DiscPropName
			s.Interfaces = append(s.Interfaces, InterfaceInfo{

				TypeNameWithPrefix:    importMap.PrefixExpr(ifaceProp.Interface.TypeSpec.Name(), ifacePkg),
				TypeName:              ifaceProp.Interface.TypeSpec.Name(),
				PkgPath:               ifacePkg.PkgPath,
				UnmarshalerFunc:       ifaceProp.UnmarshalerFunc(),
				DiscriminatorPropName: discProp,
				Options:               opts,
			})
		}
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

func (s SchemaBuilder) RenderSchemas(noChanges, force bool) (changedSchemas map[string]bool, err error) {
	var targetDir = filepath.Join(s.Scan.Pkg.Dir, s.Subdir)
	changedSchemas = make(map[string]bool)

	if err = os.MkdirAll(targetDir, 0755); err != nil {
		return nil, fmt.Errorf("could not create subdir %s: %w", targetDir, err)
	}
	for _, method := range s.Scan.SchemaMethods {
		var changed bool
		if changed, err = s.writeSchema(method.Receiver, targetDir, noChanges); err != nil {
			return nil, err
		}
		changedSchemas[method.Receiver.TypeName] = changed || force
	}
	for _, fn := range s.Scan.SchemaFuncs {
		var changed bool
		if changed, err = s.writeSchema(fn.Receiver, targetDir, noChanges); err != nil {
			return nil, err
		}
		changedSchemas[fn.Receiver.TypeName] = changed || force
	}
	return changedSchemas, nil
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
			return syntax.NoStructType, fmt.Errorf("embedded ident should be alias or struct type %s at %s", ts.Details(), ts.Position())
		}

	case *dst.StarExpr:
		return s.resolveEmbeddedType(t.Derive(expr.X), seen)
	case *dst.ParenExpr:
		return s.resolveEmbeddedType(t.Derive(expr.X), seen)
	default:
		fmt.Println(string(debug.Stack()))
		return syntax.NoStructType, fmt.Errorf("unsupported embedded field %T at %s", expr, t.Position())
	}
}

func (s SchemaBuilder) renderStructProps(t syntax.StructType, seenProps syntax.SeenProps, seen syntax.SeenTypes) (props ObjectPropSet, err error) {
	var myProps = slices.Clone(seenProps)
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
			if embeddedType, err = s.resolveEmbeddedType(prop.TypeExpr, seen); err != nil {
				return nil, fmt.Errorf("resolving embedded type: %w", err)
			} else if tempProps, err = s.renderStructProps(embeddedType, myProps, seen); err != nil {
				return nil, fmt.Errorf("rendering embedded type: %w", err)
			}
		} else if tempProps, err = s.renderStructField(t, prop, seen); err != nil {
			return nil, fmt.Errorf("rendering struct field: %w", err)
		}
		props = append(props, tempProps...)
	}
	return props, nil
}

func hasProviderForGoField(list []FieldProvider, goFieldNames []string) bool {
	for _, it := range list {
		if slices.Contains(goFieldNames, it.FieldName) {
			return true
		}
	}
	return false
}

func (s SchemaBuilder) renderStructField(owner syntax.StructType, f syntax.StructField, seen syntax.SeenTypes) (props []ObjectProp, err error) {
	var (
		schema JSONSchema
		name   string
	)
	// Prefer centralized tag parsing
	if f.Field.Tag != nil && f.Field.Tag.Value != "" {
		if tag := common.ParseJSONSchemaTag(f.Field.Tag.Value); tag.HasRef {
			schema = RefNode{Ref: tag.Ref}
		}
	}
	if schema == nil {
		// Enums v1
		if cfgs, okEnum := s.EnumV1[owner.Name()]; okEnum {
			for _, goField := range f.Field.Names {
				cfg, ok2 := cfgs[goField.Name]
				if !ok2 {
					continue
				}
				ident, ok := f.Field.Type.(*dst.Ident)
				if !ok {
					continue
				}
				pkgPath := ident.Path
				if pkgPath == "" {
					pkgPath = s.Scan.Pkg.PkgPath
				}
				scanRes, okp := s.Scan.GetPackage(pkgPath)
				if !okp {
					continue
				}
				enumSet, okE := scanRes.Constants[ident.Name]
				if !okE {
					continue
				}

				// Determine if the enum is string-based by checking the first value
				isStringEnum := false
				if len(enumSet.Values) > 0 && len(enumSet.Values[0].Value().Values) > 0 {
					if _, ok := enumSet.Values[0].Value().Values[0].(*dst.BasicLit); ok {
						// It has a BasicLit value, likely a string
						isStringEnum = true
					}
				}

				// Use string mode if explicitly set or if it's a string-based enum
				if cfg.Mode == "strings" || (cfg.Mode == "" && isStringEnum) {
					var vals []string
					for _, v := range enumSet.Values {
						var value string
						if isStringEnum && len(v.Value().Values) > 0 {
							// Get the actual string value
							if lit, ok := v.Value().Values[0].(*dst.BasicLit); ok {
								value = strings.Trim(lit.Value, "\"")
							} else {
								value = v.Value().Names[0].Name
							}
						} else {
							// For iota enums with string mode, use the constant name
							value = v.Value().Names[0].Name
						}

						// Check for custom name override
						if override, okn := cfg.Names[v.Value().Names[0].Name]; okn {
							value = override
						}
						vals = append(vals, value)
					}
					schema = PropertyNode[string]{Typ: "string", Enum: vals, TypeID_: f.ID()}
				} else {
					var vals []int
					iotaVal := 0
					for _, v := range enumSet.Values {
						if len(v.Value().Values) > 0 {
							if bl, ok := v.Value().Values[0].(*dst.BasicLit); ok && bl.Kind == token.INT {
								if n, err := strconv.Atoi(bl.Value); err == nil {
									iotaVal = n
								}
							}
						}
						vals = append(vals, iotaVal)
						iotaVal++
					}
					schema = PropertyNode[int]{Typ: "integer", Enum: vals, TypeID_: f.ID()}
				}
				break
			}
		}
		// Interfaces v1
		if cfgs, okV1 := s.IfaceV1[owner.Name()]; okV1 {
			for _, goField := range f.Field.Names {
				if cfg, ok2 := cfgs[goField.Name]; ok2 {
					var union = UnionTypeNode{DiscriminatorPropName: cfg.Disc, TypeID_: f.ID()}
					for _, impl := range cfg.Impls {
						if err = s.mapType(impl, seen); err != nil {
							return nil, fmt.Errorf("rendering interface impl: %w", err)
						}
						implSchema, ok := s.GetSchema(impl)
						if !ok {
							return nil, fmt.Errorf("type %s is not a known schema", impl)
						}
						obj, ok2b := implSchema.(ObjectNode)
						if !ok2b {
							return nil, fmt.Errorf("expected %s to be an object-type schema", impl.TypeName)
						}
						obj2 := obj
						obj2.Discriminator = impl.TypeName
						union.Options = append(union.Options, obj2)
					}
					schema = union
					break
				}
			}
		}
		// Providers
		if schema == nil {
			if providers, ok := s.TypeProvidersMap[f.Name()]; ok {
				var goNames []string
				for _, ident := range f.Field.Names {
					goNames = append(goNames, ident.Name)
				}
				if hasProviderForGoField(providers, goNames) {
					jsonNames := f.PropNames()
					if len(jsonNames) > 0 {
						schema = TemplateHoleNode{Name: jsonNames[0]}
					}
				}
			}
		}
		// Fallback
		if schema == nil {
			if schema, err = s.renderSchema(f.Derive(f.Type()), f.Comments(), seen); err != nil {
				return nil, fmt.Errorf("rendering schema: %w", err)
			}
		}
	}
	for _, name = range f.PropNames() {
		props = append(props, ObjectProp{
			Name:     name,
			Schema:   schema,
			Optional: !f.Required(),
		})
	}
	return props, nil
}

type InterfaceProp struct {
	Field         syntax.StructField
	Interface     syntax.IfaceImplementations
	DiscPropName  string
	FuncNameAlias string
}

func (s InterfaceProp) UnmarshalerFunc() string {
	if s.FuncNameAlias != "" {
		return s.FuncNameAlias
	}
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
			// Prefer v1 interface options if present for this receiver field
			if cfgs, okMap := s.IfaceV1[t.Name()]; okMap {
				var handled bool
				for _, goField := range prop.Field.Names {
					if cfg, ok2 := cfgs[goField.Name]; ok2 {
						// Resolve interface TypeSpec for ident
						pkgPath := ident.Path
						if pkgPath == "" {
							pkgPath = s.Scan.Pkg.PkgPath
						}
						scanRes, okp := s.Scan.GetPackage(pkgPath)
						if !okp {
							return nil, fmt.Errorf("could not find package for interface %s", ident.Name)
						}
						ts, okts := scanRes.LocalNamedTypes[ident.Name]
						if !okts {
							return nil, fmt.Errorf("could not resolve interface type %s", ident.Name)
						}
						iface := syntax.IfaceImplementations{TypeSpec: ts, Impls: cfg.Impls}
						if len(prop.Field.Names) != 1 {
							return nil, fmt.Errorf("interface prop %s has more than one field name at %s", strings.Join(prop.PropNames(), ","), prop.Position())
						}
						funcAlias := fmt.Sprintf("__jsonUnmarshal__%s__%s__%s__%s", ts.Pkg().Name, ts.Name(), t.Name(), goField.Name)
						props = append(props, InterfaceProp{Field: prop, Interface: iface, DiscPropName: cfg.Disc, FuncNameAlias: funcAlias})
						handled = true
						break
					}
				}
				if handled {
					continue
				}
			}
			// Legacy interface resolution

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
		if _t, err := s.resolveEmbeddedType(prop.TypeExpr, nil); err != nil {
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
