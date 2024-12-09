package typeregistry

import (
	"github.com/dave/dst"
)

func isRelevantFunc(funcDecl *dst.FuncDecl) bool {
	// Check function arguments
	params := funcDecl.Type.Params
	results := funcDecl.Type.Results

	if funcDecl.Recv == nil {
		if len(params.List) != 1 {
			return false
		}
	} else if _, ok := funcDecl.Recv.List[0].Type.(*dst.StarExpr); ok {
		return false
	} else if len(params.List) != 0 {
		return false
	}

	if results == nil || len(results.List) != 2 {
		return false
	}
	var ident, ok = results.List[1].Type.(*dst.Ident)
	return ok && ident.Name == "error"
}
