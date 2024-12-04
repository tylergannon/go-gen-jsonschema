package typeregistry

import (
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"go/types"
)

type edgeType int

const (
	edgeTypeIdentity edgeType = iota
	edgeTypeField
	edgeTypeDefinition
)

type edge struct {
	Type edgeType
	From TypeID
	To   TypeID
}

type Node struct {
	resolvedType  TypeID             // update this after resolving the type
	parent        *Node              // the Node that caused this to go onto the stack
	typ           types.Type         // the types object
	expr          dst.Expr           // the Expr Node representing this Node
	children      []*edge            // before leaving from the first visit, must be initialized to have length equal to number of child nodes
	indexAtParent int                // the index where this Node should reflect its resolved type within `parent.children`.
	processed     bool               // true value means we've already scanned this Node's children
	pkg           *decorator.Package // the decorated package that this Node occurs in
	typeSpec      TypeSpec           // may be nil; only exists for top-level type decls.
}

func (r *Registry) graphTypeForSchema(ts TypeSpec) (*Node, map[TypeID]*Node, error) {
	var (
		nodes    = map[TypeID]*Node{}
		rootNode = &Node{
			typ:      ts.GetType(),
			typeSpec: ts,
			pkg:      ts.Pkg(),
			expr:     ts.GetTypeSpec().Type,
		}
		stack = []*Node{rootNode}
	)

	for len(stack) > 0 {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if nodesTemp, err := r.visitNode(node, nodes); err != nil {
			return nil, nil, err
		} else {
			stack = append(stack, nodesTemp...)
		}
	}
	return rootNode, nodes, nil
}

func (r *Registry) visitNode(node *Node, nodes map[TypeID]*Node) ([]*Node, error) {

	return nil, nil
}
