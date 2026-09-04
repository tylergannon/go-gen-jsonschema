package builder

import (
	"fmt"
	"go/token"
	"go/types"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/tylergannon/go-gen-jsonschema/internal/syntax"
)

type InspectArgs struct {
	TargetDir string
	TypeNames []string
}

type PackageInspection struct {
	PackagePath  string
	Roots        []RootInspection
	Unregistered []string
}

type RootInspection struct {
	TypePath                string
	Position                token.Position
	HasProviders            bool
	RequiresUnionCodec      bool
	RequiresStringEnumCodec bool
	Findings                []InspectionFinding
	Err                     error
}

type InspectionFinding struct {
	Code      string
	Certainty string
	Message   string
	Remedy    string
	TypePath  string
	FieldPath string
	Position  token.Position
}

type InspectionError struct {
	Code      string
	Certainty string
	Message   string
	Remedy    string
	TypePath  string
	FieldPath string
	Position  token.Position
	Cause     error
}

func (e *InspectionError) Error() string { return e.Message }
func (e *InspectionError) Unwrap() error { return e.Cause }

// Inspect performs the same package scan and schema mapping used by generation,
// one registered root at a time. It does not render, invoke providers, or write.
func Inspect(args InspectArgs) (PackageInspection, error) {
	pkgs, err := syntax.LoadReadonly(args.TargetDir)
	if err != nil {
		return PackageInspection{}, fmt.Errorf("load package %s: %w", args.TargetDir, err)
	}
	if len(pkgs) == 0 {
		return PackageInspection{}, fmt.Errorf("no packages found in %s", args.TargetDir)
	}
	scan, err := syntax.LoadPackage(pkgs[0])
	if err != nil {
		return PackageInspection{}, &InspectionError{
			Code:      "scan_unclassified",
			Certainty: "unknown",
			Message:   fmt.Sprintf("could not prove the v1 model boundary while scanning package %s: %v", pkgs[0].PkgPath, err),
			Remedy:    "review the scanner error and use a documented supported model shape",
			Cause:     err,
		}
	}
	registered := make(map[string]token.Position)
	for _, method := range scan.SchemaMethods {
		registered[method.Receiver.TypeName] = method.MarkerCall.CallExpr.Position()
	}
	for _, function := range scan.SchemaFuncs {
		registered[function.Receiver.TypeName] = function.MarkerCall.CallExpr.Position()
	}

	selected := slices.Clone(args.TypeNames)
	if len(selected) == 0 {
		selected = make([]string, 0, len(registered))
		for name := range registered {
			selected = append(selected, name)
		}
	}
	slices.Sort(selected)
	selected = slices.Compact(selected)

	result := PackageInspection{PackagePath: pkgs[0].PkgPath}
	for _, name := range selected {
		position, ok := registered[name]
		if !ok {
			result.Unregistered = append(result.Unregistered, name)
			continue
		}
		root := RootInspection{
			TypePath: scan.Pkg.PkgPath + "." + name,
			Position: position,
		}
		root.Findings = inspectStaticShape(scan, name)
		mapped, mapErr := inspectRoot(pkgs[0], name)
		if mapErr != nil {
			if typed, ok := mapErr.(*InspectionError); ok {
				root.Err = typed
			} else {
				root.Err = &InspectionError{
					Code:      "mapping_unclassified",
					Certainty: "unknown",
					Message:   fmt.Sprintf("could not prove the v1 mapping for %s: %v", root.TypePath, mapErr),
					Remedy:    "review the reported builder error and use a documented supported shape",
					TypePath:  root.TypePath,
					Position:  position,
					Cause:     mapErr,
				}
			}
		} else {
			for mappedName := range mapped.schemas[mapped.Scan.Pkg.PkgPath] {
				root.HasProviders = root.HasProviders || len(mapped.TypeProvidersMap[mappedName]) > 0
				root.RequiresUnionCodec = root.RequiresUnionCodec || len(mapped.customTypes[mappedName]) > 0 || len(mapped.IfaceV1[mappedName]) > 0
				for _, enum := range mapped.EnumV1[mappedName] {
					if enum.UseStringer {
						root.RequiresStringEnumCodec = true
						break
					}
				}
			}
		}
		result.Roots = append(result.Roots, root)
	}
	return result, nil
}

func inspectRoot(pkg *decorator.Package, name string) (mapped SchemaBuilder, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = &InspectionError{
				Code:      "internal_inspection_panic",
				Certainty: "internal",
				Message:   fmt.Sprintf("panic while inspecting %s: %v", name, recovered),
				Remedy:    "report this diagnostic and the installed tool revision",
			}
		}
	}()
	return NewForTypes(pkg, []string{name})
}

type staticInspector struct {
	findings          []InspectionFinding
	seen              map[syntax.TypeID]bool
	productionHooks   map[syntax.TypeID]token.Position
	productionScanned map[string]bool
}

func inspectStaticShape(scan syntax.ScanResult, rootName string) []InspectionFinding {
	inspector := staticInspector{
		seen:              make(map[syntax.TypeID]bool),
		productionHooks:   make(map[syntax.TypeID]token.Position),
		productionScanned: make(map[string]bool),
	}
	typeSpec, ok := scan.LocalNamedTypes[rootName]
	if !ok {
		return nil
	}
	inspector.walkNamed(scan, typeSpec, rootName, rootName)
	slices.SortFunc(inspector.findings, func(a, b InspectionFinding) int {
		if c := strings.Compare(a.FieldPath, b.FieldPath); c != 0 {
			return c
		}
		return strings.Compare(a.Code, b.Code)
	})
	return inspector.findings
}

func (i *staticInspector) walkNamed(scan syntax.ScanResult, spec syntax.TypeSpec, typePath, fieldPath string) {
	id := spec.ID().Concrete()
	if i.seen[id] {
		i.add("unsupported_recursive_type", "unsupported", "recursive types are outside the v1 contract", "replace the recursive edge with a nonrecursive supported representation", typePath, fieldPath, spec.Position())
		return
	}
	if err := i.loadProductionHooks(scan); err != nil {
		i.add("unknown_production_method_scan", "unknown", fmt.Sprintf("could not inspect production JSON methods: %v", err), "fix the source file error and run inspection again", typePath, fieldPath, spec.Position())
	}
	if position, found := i.productionHooks[spec.ID().Concrete()]; found {
		i.add("unknown_custom_json_hook", "unknown", "custom JSON hooks can change the wire shape and cannot be proven by static inspection", "verify the hook against the generated schema or use a documented supported shape", typePath, fieldPath, position)
	}
	i.seen[id] = true
	if structNode, ok := spec.Type().Expr().(*dst.StructType); ok {
		structType := syntax.NewStructType(structNode, spec)
		for _, field := range structType.Fields() {
			if field.Skip() {
				continue
			}
			fieldNames := goFieldNames(field)
			if len(fieldNames) == 0 {
				fieldNames = []string{"<embedded>"}
			}
			for _, name := range fieldNames {
				nextPath := fieldPath + "." + name
				wrapper, inner, wrapperErr := field.Wrapper()
				if wrapperErr != nil {
					i.add("unsupported_wrapper_shape", "unsupported", wrapperErr.Error(), "use Optional[T] or Nullable[T] directly as a named exported field", typePath, nextPath, field.Position())
					continue
				}
				if wrapper == syntax.WrapperNone && (field.HasJSONOption("omitempty") || field.HasJSONOption("omitzero")) {
					i.add("unsupported_required_omission", "unsupported", "ordinary required fields cannot use omitempty or omitzero within the v1 contract", "use Optional[T] for absence or remove the omission option", typePath, nextPath, field.Position())
				}
				if field.HasJSONOption("string") {
					i.add("unsupported_json_string", "unsupported", "json:,string changes the wire shape and is outside the v1 contract", "use the natural JSON scalar representation or a separate supported wire field", typePath, nextPath, field.Position())
				}
				if inspectionRegisteredInterfaceField(scan, spec.Name(), name) {
					if !supportedRegisteredInterfaceShape(scan, field.Type(), wrapper, inner) {
						i.add("unsupported_interface_shape", "unsupported", "registered interfaces support only direct I, Optional[I], and direct []I fields", "move the union to a direct named field, Optional[I], or one-dimensional []I field", typePath, nextPath, field.Position())
					}
					continue
				}
				if wrapper != syntax.WrapperNone {
					i.walkExpr(scan, inner, typePath, nextPath, field.Position())
				} else {
					i.walkExpr(scan, field.Type(), typePath, nextPath, field.Position())
				}
			}
		}
		delete(i.seen, id)
		return
	}
	i.walkExpr(scan, spec.Type().Expr(), typePath, fieldPath, spec.Position())
	delete(i.seen, id)
}

func (i *staticInspector) walkExpr(scan syntax.ScanResult, expr dst.Expr, typePath, fieldPath string, position token.Position) {
	switch node := expr.(type) {
	case *dst.Ident:
		if syntax.BasicTypes[node.Name] {
			return
		}
		pkgPath := node.Path
		if pkgPath == "" || pkgPath == scan.Pkg.PkgPath {
			if _, registeredUnion := scan.Interfaces[node.Name]; registeredUnion {
				return
			}
			if spec, ok := scan.LocalNamedTypes[node.Name]; ok {
				i.walkNamed(scan, spec, typePath, fieldPath)
			}
			return
		}
		if syntax.IsTimeType(pkgPath, node.Name) {
			return
		}
		remote, ok := scan.GetPackage(pkgPath)
		if !ok {
			i.add("unknown_external_type", "unknown", fmt.Sprintf("external type %s.%s has no proven schema/codec mapping", pkgPath, node.Name), "use a documented supported type or provide explicit schema and codec evidence", typePath, fieldPath, position)
			return
		}
		if scan.Pkg.Module == nil || remote.Pkg.Module == nil || scan.Pkg.Module.Path != remote.Pkg.Module.Path {
			i.add("unknown_external_type", "unknown", fmt.Sprintf("external type %s.%s has no proven schema/codec mapping", pkgPath, node.Name), "use a documented supported type or provide explicit schema and codec evidence", typePath, fieldPath, position)
			return
		}
		if spec, ok := remote.LocalNamedTypes[node.Name]; ok {
			i.walkNamed(remote, spec, typePath, fieldPath)
			return
		}
		i.add("unknown_external_type", "unknown", fmt.Sprintf("external type %s.%s could not be resolved for inspection", pkgPath, node.Name), "use a documented supported type or make the type available in the module graph", typePath, fieldPath, position)
	case *dst.SelectorExpr:
		path := node.Sel.Path
		if qualifier, ok := node.X.(*dst.Ident); ok && path == "" {
			path = qualifier.Path
		}
		i.walkExpr(scan, &dst.Ident{Name: node.Sel.Name, Path: path}, typePath, fieldPath, position)
	case *dst.StarExpr:
		i.walkExpr(scan, node.X, typePath, fieldPath, position)
	case *dst.ParenExpr:
		i.walkExpr(scan, node.X, typePath, fieldPath, position)
	case *dst.ArrayType:
		if node.Len == nil && isByteElement(scan, node.Elt) {
			i.add("unsupported_base64_bytes", "unsupported", "byte-like slices use base64 JSON encoding, which is outside the v1 roundtrip contract", "use a string with explicit encoding semantics or a supported numeric array representation", typePath, fieldPath, position)
			return
		}
		i.walkExpr(scan, node.Elt, typePath, fieldPath+"[]", position)
	case *dst.MapType:
		i.add("unsupported_map", "unsupported", "map fields are outside the v1 contract", "replace the map with a struct or a supported list of key/value records", typePath, fieldPath, position)
	case *dst.ChanType:
		i.add("unsupported_channel", "unsupported", "channel fields cannot cross the JSON boundary", "exclude the field from JSON or replace it with a supported value", typePath, fieldPath, position)
	case *dst.FuncType:
		i.add("unsupported_function", "unsupported", "function fields cannot cross the JSON boundary", "exclude the field from JSON", typePath, fieldPath, position)
	case *dst.InterfaceType:
		i.add("unsupported_inline_interface", "unsupported", "inline interfaces are outside the v1 contract", "declare and register a named interface in a supported field form", typePath, fieldPath, position)
	case *dst.StructType:
		for _, field := range node.Fields.List {
			options, ignored := inlineJSONOptions(field)
			if ignored {
				continue
			}
			if len(field.Names) == 0 {
				i.walkExpr(scan, field.Type, typePath, fieldPath, position)
				continue
			}
			for _, name := range field.Names {
				if !name.IsExported() {
					continue
				}
				nextPath := fieldPath + "." + name.Name
				if slices.Contains(options, "string") {
					i.add("unsupported_json_string", "unsupported", "json:,string changes the wire shape and is outside the v1 contract", "use the natural JSON scalar representation or a separate supported wire field", typePath, nextPath, position)
				}
				i.walkExpr(scan, field.Type, typePath, nextPath, position)
			}
		}
	case *dst.IndexExpr:
		i.walkExpr(scan, node.Index, typePath, fieldPath, position)
	case *dst.IndexListExpr:
		for _, index := range node.Indices {
			i.walkExpr(scan, index, typePath, fieldPath, position)
		}
	default:
		i.add("unknown_go_shape", "unknown", fmt.Sprintf("Go shape %T has no proven v1 mapping", expr), "use a documented supported shape", typePath, fieldPath, position)
	}
}

func (i *staticInspector) loadProductionHooks(scan syntax.ScanResult) error {
	if i.productionScanned[scan.Pkg.PkgPath] {
		return nil
	}
	i.productionScanned[scan.Pkg.PkgPath] = true
	methods, err := syntax.FindProductionJSONMethods(scan.Pkg.Dir, nil)
	if err != nil {
		return err
	}
	for _, method := range methods {
		i.productionHooks[syntax.TypeID{PkgPath: scan.Pkg.PkgPath, TypeName: method.Receiver}.Concrete()] = method.Position
	}
	return nil
}

func isByteElement(scan syntax.ScanResult, expr dst.Expr) bool {
	ident, ok := expr.(*dst.Ident)
	if !ok {
		return false
	}
	if ident.Name == "byte" || ident.Name == "uint8" {
		return true
	}
	target := scan
	if ident.Path != "" && ident.Path != scan.Pkg.PkgPath {
		var found bool
		target, found = scan.GetPackage(ident.Path)
		if !found {
			return false
		}
	}
	if target.Pkg == nil || target.Pkg.Package == nil || target.Pkg.Types == nil {
		return false
	}
	object := target.Pkg.Types.Scope().Lookup(ident.Name)
	if object == nil {
		return false
	}
	basic, ok := object.Type().Underlying().(*types.Basic)
	return ok && basic.Kind() == types.Uint8
}

func inlineJSONOptions(field *dst.Field) (options []string, ignored bool) {
	if field.Tag == nil {
		return nil, false
	}
	value, err := strconv.Unquote(field.Tag.Value)
	if err != nil {
		return nil, false
	}
	jsonTag := reflect.StructTag(value).Get("json")
	parts := strings.Split(jsonTag, ",")
	if len(parts) > 0 && parts[0] == "-" {
		return nil, true
	}
	if len(parts) > 1 {
		return parts[1:], false
	}
	return nil, false
}

func goFieldNames(field syntax.StructField) []string {
	if len(field.Field.Names) > 0 {
		names := make([]string, 0, len(field.Field.Names))
		for _, name := range field.Field.Names {
			if name.IsExported() {
				names = append(names, name.Name)
			}
		}
		return names
	}
	return []string{embeddedFieldName(field.Type())}
}

func embeddedFieldName(expr dst.Expr) string {
	switch node := expr.(type) {
	case *dst.Ident:
		return node.Name
	case *dst.SelectorExpr:
		return node.Sel.Name
	case *dst.StarExpr:
		return embeddedFieldName(node.X)
	case *dst.ParenExpr:
		return embeddedFieldName(node.X)
	default:
		return "<embedded>"
	}
}

func inspectionRegisteredInterfaceField(scan syntax.ScanResult, receiver, field string) bool {
	check := func(method syntax.SchemaMethod) bool {
		if method.Receiver.TypeName != receiver {
			return false
		}
		for _, option := range method.Options {
			if option.Kind == "WithInterface" && option.FieldName == field {
				return true
			}
		}
		return false
	}
	for _, method := range scan.SchemaMethods {
		if check(method) {
			return true
		}
	}
	for _, function := range scan.SchemaFuncs {
		if check(syntax.SchemaMethod(function)) {
			return true
		}
	}
	return false
}

func supportedRegisteredInterfaceShape(scan syntax.ScanResult, fieldType dst.Expr, wrapper syntax.WrapperKind, inner dst.Expr) bool {
	if wrapper != syntax.WrapperNone {
		return wrapper == syntax.WrapperOptional && isNamedInterface(scan, inner)
	}
	if isNamedInterface(scan, fieldType) {
		return true
	}
	array, ok := fieldType.(*dst.ArrayType)
	return ok && array.Len == nil && isNamedInterface(scan, array.Elt)
}

func isNamedInterface(scan syntax.ScanResult, expr dst.Expr) bool {
	ident, ok := expr.(*dst.Ident)
	if !ok || (ident.Path != "" && ident.Path != scan.Pkg.PkgPath) {
		return false
	}
	if _, ok := scan.Interfaces[ident.Name]; ok {
		return true
	}
	typeSpec, ok := scan.LocalNamedTypes[ident.Name]
	if !ok {
		return false
	}
	_, ok = typeSpec.Type().Expr().(*dst.InterfaceType)
	return ok
}

func (i *staticInspector) add(code, certainty, message, remedy, typePath, fieldPath string, position token.Position) {
	i.findings = append(i.findings, InspectionFinding{
		Code: code, Certainty: certainty, Message: message, Remedy: remedy,
		TypePath: typePath, FieldPath: fieldPath, Position: position,
	})
}
