package typeregistry

import (
	"errors"
	"fmt"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/tylergannon/structtag"
	"go/ast"
	"go/token"
	"go/types"
	"log"
	"strings"
)

func (r *Registry) LoadAndValidateNamedType(namedType *types.Named) (TypeSpec, error) {
	typeName := namedType.Obj()

	// Okay we should be able to find a type spec for the given type.
	// Of not then we scan the package.
	if err := r.LoadAndScan(typeName.Pkg().Path()); err != nil {
		return nil, fmt.Errorf("failed to load and scan package %q: %w", typeName.Pkg().Path(), err)
	}
	ts, _, ok := r.getType(typeName.Name(), typeName.Pkg().Path())
	if !ok {
		return nil, fmt.Errorf("failed to find type %q in package %q", typeName.Name(), typeName.Pkg().Path())
	}
	switch namedType.Underlying().(type) {
	case *types.Struct:
		return r.loadAndValidateNamedStruct(ts, namedType)
	}

	return r.LoadAndValidateType(namedType.Underlying())
}

var (
	ErrUnsupportedType = errors.New("unsupported type")
)

func (r *Registry) LoadAndValidateType(typ types.Type) (TypeSpec, error) {
	inspect("LoadAndValidateType", typ)
	switch t := typ.(type) {
	case *types.Basic:
		return resolveBasicType(t)
	case *types.Named:
		return r.LoadAndValidateNamedType(t)
	case *types.Pointer:
		return r.LoadAndValidateType(t.Elem())
	case *types.Chan:
		return nil, fmt.Errorf("chan types are not supported: %w", ErrUnsupportedType)
	case *types.Signature:
		return nil, fmt.Errorf("signature types are not supported: %w", ErrUnsupportedType)
	case *types.Interface:
		return nil, fmt.Errorf("interface types are not supported: %w", ErrUnsupportedType)
	case *types.Alias:
		return r.LoadAndValidateType(t.Underlying())
	}
	inspect("Type", typ)
	return nil, errors.New("unsupported type")
}

func (r *Registry) loadAndValidateNamedStruct(ts *typeSpec, namedType *types.Named) (*NamedStructSpec, error) {
	structType, ok := namedType.Underlying().(*types.Struct)
	if !ok {
		panic("not a struct type (in loadAndValidateNamedStruct)")
	}
	structNode, ok := ts.typeSpec.Type.(*dst.StructType)
	if !ok {
		panic("ast node is not a struct type (in loadAndValidateNamedStruct)")
	}
	var (
		spec = &NamedStructSpec{
			Description: buildComments(ts.Decorations()),
			TypeName:    namedType.Obj(),
			TypeSpec:    ts,
		}
		err error
	)
	if spec.StructTypeSpec, err = r.LoadAndValidateStruct(spec, structType, structNode); err != nil {
		return nil, fmt.Errorf("failed to load and validate struct: %w", err)
	}

	return spec, nil
}

type fieldConfiguration struct {
	parent         *NamedStructSpec
	jsonTagOptions []string
	fieldName      string
	*dst.Field
	*types.Var
}

func (f fieldConfiguration) pos() token.Position {
	return f.parent.Pkg().Fset.Position(f.Pos())
}

func (f fieldConfiguration) ignore() bool {
	return f.fieldName == "-"
}

func (f fieldConfiguration) typeExpr() dst.Expr {
	return f.Field.Type
}

func (f fieldConfiguration) fieldType() types.Type {
	return f.Var.Type()
}

func (spec *NamedStructSpec) parseFieldConf(field *types.Var, fieldNode *dst.Field) (fieldConfiguration, error) {
	inspect("parseFieldConf", spec, field, fieldNode)
	var (
		fieldName      string
		jsonTagOptions []string
	)
	if !field.Embedded() && fieldNode.Tag != nil {
		tagValue := strings.Trim(fieldNode.Tag.Value, "`")
		if tags, err := structtag.Parse(tagValue); err != nil {
			position := spec.Pkg().Fset.Position(field.Pos())

			return fieldConfiguration{}, fmt.Errorf(
				"failed to parse tags for field %q at %s:%d:%d: %w",
				field.Name(), position.Filename, position.Line, position.Column, err,
			)
		} else if tag, err := tags.Get("json"); err != nil || len(tag.Options) == 0 {
			fieldName = field.Name()
		} else {
			fieldName = tag.Options[0]
			jsonTagOptions = tag.Options
		}
	}
	return fieldConfiguration{
		jsonTagOptions: jsonTagOptions,
		fieldName:      fieldName,
		Var:            field,
		Field:          fieldNode,
		parent:         spec,
	}, nil
}

func (r *Registry) LoadAndValidateStruct(parent *NamedStructSpec, typ *types.Struct, str *dst.StructType) (spec StructTypeSpec, err error) {

	inspect("Struct", typ)
	for i := 0; i < typ.NumFields(); i++ {
		var (
			fieldConf fieldConfiguration
			fieldTS   TypeSpec
		)
		if fieldConf, err = parent.parseFieldConf(typ.Field(i), str.Fields.List[i]); err != nil {
			return spec, fmt.Errorf("loading struct %s: %w", parent.TypeName.Name(), err)
		} else if fieldConf.ignore() {
			continue
		} else if !fieldConf.Exported() {
			return spec, fmt.Errorf("exported fields are not supported: %w", ErrUnsupportedType)
		}
	
		switch fieldType := fieldConf.fieldType().(type) {
		case *types.Struct:
			if fieldStructExpr, ok := str.Fields.List[i].Type.(*dst.StructType); !ok {
				position := fieldConf.pos()
				log.Println(position)
				log.Panicf("struct field %q is not a struct type but %T\n", fieldConf.fieldName, str.Fields.List[0].Type)
			} else {
				var inlineStruct StructTypeSpec
				inlineStruct, err = r.LoadAndValidateStruct(parent, fieldType, fieldStructExpr)
				fieldTS = NewInlineStructSpec(parent, inlineStruct)
			}
		default:
			fieldTS, err = r.LoadAndValidateType(fieldConf.fieldType())
		}

		if err != nil {
			fieldPos := fieldConf.pos()
			return spec, fmt.Errorf("loading struct %s, field %s at %s:%d:%d: %w",
				parent.TypeName.Name(), fieldConf.fieldName, fieldPos.Filename, fieldPos.Line, fieldPos.Column, err,
			)
		}
		if !fieldConf.Embedded() {
			spec.Fields = append(spec.Fields, newStructField(fieldConf, fieldTS))
			continue
		}
		switch fieldType := fieldTS.(type) {
		case *NamedStructSpec:
			for _, field := range fieldType.StructTypeSpec.Fields {
				spec.Fields = append(spec.Fields, field)
			}
		default:
			return spec, fmt.Errorf("embedded field type %T are not supported: %w", fieldTS.GetTypeSpec(), ErrUnsupportedType)
		}

	}
	return spec, nil
}

type NamedStructSpec struct {
	StructTypeSpec
	TypeSpec
	TypeName         *types.TypeName
	NumInlineStructs int
	Description      string
	file             *ast.File
}

type InlineStructSpec struct {
	StructTypeSpec
	Parent *NamedStructSpec
	Idx    int
}

func NewInlineStructSpec(parent *NamedStructSpec, spec StructTypeSpec) *InlineStructSpec {
	idx := parent.NumInlineStructs
	parent.NumInlineStructs++
	return &InlineStructSpec{
		Idx:            idx,
		Parent:         parent,
		StructTypeSpec: spec,
	}
}

func (i InlineStructSpec) ID() TypeID {
	return TypeID(string(i.Parent.ID()) + fmt.Sprintf("_inline%d", i.Idx))
}

func (i InlineStructSpec) GetTypeSpec() *dst.TypeSpec {
	panic("Not implemented on inline type")
}

func (i InlineStructSpec) Pkg() *decorator.Package {
	return i.Parent.Pkg()
}

func (i InlineStructSpec) GenDecl() *dst.GenDecl {
	panic("Not Implemented on inline type.")
}

func (i InlineStructSpec) File() *dst.File {
	return i.Parent.File()
}

func (i InlineStructSpec) Decorations() *dst.NodeDecs {
	panic("Not Implemented on inline type.")
}

func (i InlineStructSpec) GetType() *types.TypeName {
	panic("Not Implemented on inline type.")
}

var _ TypeSpec = (*InlineStructSpec)(nil)

type StructTypeSpec struct {
	// First one is the actual struct, following ones are embedded structs,
	// in order of their declaration
	Structs []*types.Struct
	// The resolved struct nodes, meaning that all embedded fields have been
	// tracked from the type name to their declaration.
	StructNodes []*dst.StructType
	// The flattened list of all fields including embedded types
	Fields []StructField
}

type StructField struct {
	JSONName    string
	Type        TypeSpec
	Description string
}

func newStructField(conf fieldConfiguration, ts TypeSpec) StructField {
	return StructField{
		JSONName:    conf.fieldName,
		Type:        ts,
		Description: buildComments(conf.Decorations()),
	}
}

type EnumType int

const (
	EnumTypeInt EnumType = iota
	EnumTypeString
	EnumTypeIota
)

type EnumField struct {
	TypeSpec
	Type        EnumType
	Description string
	Values      []string
}
