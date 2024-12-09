package typeregistry

import (
	"errors"
	"fmt"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"go/types"
	"strings"
)

const maxRecursionDepth = 500

var errRecursionDepthExceeded = errors.New("recursion depth exceeded (this probably means circular embedded pointer types)")

// resolveType recurses into the given type up to the next named type,
// to build a string name for the given type.
func (r *Registry) resolveType(t types.Type, ts dst.Node, pkg *decorator.Package) (TypeID, error) {
	return r.resolveTypeRecursive(t, ts, pkg, 0)
}

// resolveType recurses into the given type up to the next named type,
// to build a string name for the given type.
func (r *Registry) resolveTypeRecursive(t types.Type, ts dst.Node, pkg *decorator.Package, depth int) (TypeID, error) {
	depth++
	if depth > maxRecursionDepth {
		return "", errRecursionDepthExceeded
	}

	switch t := t.(type) {
	case *types.Named:
		return NewTypeID(t.Obj().Pkg().Path(), t.Obj().Name()), nil
	case *types.Basic:
		_t := ts.(*dst.Ident)
		return TypeID(_t.Name), nil
	case *types.Pointer:
		return r.resolveTypeRecursive(t.Elem(), ts.(*dst.StarExpr).X, pkg, depth)
	case *types.Slice:
		var arrayT = ts.(*dst.ArrayType).Elt
		inner, err := r.resolveTypeRecursive(t.Elem(), arrayT, pkg, depth)
		return TypeID(fmt.Sprintf("%s[]", inner)), err
	case *types.Array:
		var arrayT = ts.(*dst.ArrayType)
		inner, err := r.resolveTypeRecursive(t.Elem(), arrayT.Elt, pkg, depth)
		if lit, ok := arrayT.Len.(*dst.BasicLit); ok {
			return TypeID(fmt.Sprintf("%s[%s]", inner, lit.Value)), err
		}
		panic(fmt.Sprintf("unhandled array len type: %T", arrayT.Len))
	case *types.Struct:
		switch nodeType := ts.(type) {
		case *dst.Ident:
			if _typeSpec, err := r.resolveTypeIdent(nodeType, pkg); err != nil {
				return "", err
			} else if structNode, ok := _typeSpec.GetTypeSpec().Type.(*dst.StructType); ok {
				return r.resolveStructType(t, structNode, pkg, depth)
			} else {
				inspect("t is struct (dst is ident)", t, nodeType)
			}
		}

		return r.resolveStructType(t, ts.(*dst.StructType), pkg, depth)
	}
	return "", fmt.Errorf("unsupported type %T: %w", t, ErrUnsupportedType)
}

func (r *Registry) resolveStructType(t *types.Struct, ts *dst.StructType, pkg *decorator.Package, depth int) (TypeID, error) {
	var (
		sb = strings.Builder{}
	)
	sb.WriteString("struct {")
	_, err := r.resolveFields(t, ts, pkg, 0, &sb, depth)
	if err != nil {
		return "", err
	}
	sb.WriteString("}")
	return TypeID(sb.String()), nil
}

func (r *Registry) resolveFields(t *types.Struct, ts *dst.StructType, pkg *decorator.Package, cntFields int, sb *strings.Builder, depth int) (int, error) {
	depth++
	if depth > maxRecursionDepth {
		return 0, errRecursionDepthExceeded
	}
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
			_ts, ok := r.getType(f.Var.Name(), f.Var.Pkg().Path())
			if !ok {
				return 0, fmt.Errorf("type %s not found in %s", f.Var.Name(), f.Var.Pkg().Path())
			}
			named, ok := _ts.GetType().(*types.Named)
			if !ok {
				return 0, fmt.Errorf("embed issue: should be named but is %T %s", _ts.GetType(), f.PositionString())
			}
			embeddedType, ok := named.Underlying().(*types.Struct)
			if !ok {
				return 0, fmt.Errorf("illegal embed type %T %s", _ts.GetType(), f.PositionString())
			}
			embeddedStructNode, ok := _ts.typeSpec.Type.(*dst.StructType)
			if !ok {
				return 0, fmt.Errorf("illegal embed found struct type but wrong node %T %s", _ts.typeSpec.Type, f.PositionString())
			}
			if cntFields, err = r.resolveFields(embeddedType, embeddedStructNode, _ts.pkg, cntFields, sb, depth); err != nil {
				return 0, err
			}

		} else {
			cntFields++
			if cntFields > 1 {
				sb.WriteString(" ")
			}

			if resolvedType, err = r.resolveTypeRecursive(f.FieldType(), f.Type, pkg, depth); err != nil {
				return 0, err
			}
			resolvedType = resolvedType.shorten(pkg)
			sb.WriteString(fmt.Sprintf("%s %s", f.FieldName, resolvedType))
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
