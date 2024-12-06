package typeregistry

import (
	"fmt"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"go/types"
	"log"
	"strings"
)

// resolveType recurses into the given type up to the next named type,
// to build a string name for the given type.
func (r *Registry) resolveType(t types.Type, ts dst.Node, pkg *decorator.Package) (TypeID, error) {
	switch t := t.(type) {
	case *types.Named:
		return NewTypeID(t.Obj().Pkg().Path(), t.Obj().Name()), nil
	case *types.Basic:
		_t := ts.(*dst.Ident)
		return TypeID(_t.Name), nil
	case *types.Pointer:
		return r.resolveType(t.Elem(), ts.(*dst.StarExpr).X, pkg)
	case *types.Slice:
		var arrayT = ts.(*dst.ArrayType).Elt
		inner, err := r.resolveType(t.Elem(), arrayT, pkg)
		return TypeID(fmt.Sprintf("%s[]", inner)), err
	case *types.Array:
		var arrayT = ts.(*dst.ArrayType)
		inner, err := r.resolveType(t.Elem(), arrayT.Elt, pkg)
		if lit, ok := arrayT.Len.(*dst.BasicLit); ok {
			return TypeID(fmt.Sprintf("%s[%s]", inner, lit.Value)), err
		}
		log.Panicf("Can't handle array len type %T", arrayT.Len)
	case *types.Struct:
		return r.resolveStructType(t, ts.(*dst.StructType), pkg)
	}
	return "", fmt.Errorf("unsupported type %T: %w", t, ErrUnsupportedType)
}

func (r *Registry) resolveStructType(t *types.Struct, ts *dst.StructType, pkg *decorator.Package) (TypeID, error) {
	var (
		sb = strings.Builder{}
	)
	sb.WriteString("struct {")
	_, err := r.resolveFields(t, ts, pkg, 0, &sb)
	if err != nil {
		return "", err
	}
	sb.WriteString("}")
	return TypeID(sb.String()), nil
}

func (r *Registry) resolveFields(t *types.Struct, ts *dst.StructType, pkg *decorator.Package, cntFields int, sb *strings.Builder) (int, error) {
	for i := 0; i < t.NumFields(); i++ {
		f, err := parseFieldConf(t.Field(i), ts.Fields.List[i], pkg)
		if err != nil {
			return 0, err
		}
		if f.ignore() {
			continue
		}
		var resolvedType TypeID
		if f.Embedded() {
			//	Must be a named type.  Should be able to look up the type.
			if err = r.LoadAndScan(f.Var.Pkg().Path()); err != nil {
				return 0, err
			}
			_ts, _, ok := r.getType(f.Var.Name(), f.Var.Pkg().Path())
			if !ok {
				return 0, fmt.Errorf("type %s not found in %s", f.Var.Name(), f.Var.Pkg().Path())
			}
			named, ok := _ts.GetType().(*types.Named)
			if !ok {
				return 0, fmt.Errorf("embed issue: should be named but is %T %s", _ts.GetType(), f.posString())
			}
			embeddedType, ok := named.Underlying().(*types.Struct)
			if !ok {
				return 0, fmt.Errorf("illegal embed type %T %s", _ts.GetType(), f.posString())
			}
			embeddedStructNode, ok := _ts.typeSpec.Type.(*dst.StructType)
			if !ok {
				return 0, fmt.Errorf("illegal embed found struct type but wrong node %T %s", _ts.typeSpec.Type, f.posString())
			}
			cntFields, err = r.resolveFields(embeddedType, embeddedStructNode, _ts.pkg, cntFields, sb)
			//cntFields, err = r.resolveFields(_ts.GetType().(*types.Named).Underlying().(*types.Struct), _ts.typeSpec.Type, _ts.pkg, cntFields, sb)

		} else {
			cntFields++
			if cntFields > 1 {
				sb.WriteString(" ")
			}
			if resolvedType, err = r.resolveType(f.fieldType(), f.Type, pkg); err != nil {
				return 0, err
			}
			resolvedType = resolvedType.shorten(pkg)
			//if len(id) > 200 {
			//	id = id.hash()
			//}
			sb.WriteString(fmt.Sprintf("%s %s", f.fieldName, resolvedType))
		}

		if f.Field.Tag != nil && f.Field.Tag.Value != "" {
			sb.WriteString(fmt.Sprintf(" %s", f.Field.Tag.Value))
		}
		if !f.Embedded() {
			sb.WriteString(";")
		}
	}
	return cntFields, nil
}
