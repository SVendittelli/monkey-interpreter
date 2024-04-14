package evaluator

import (
	"github.com/SVendittelli/monkey-interpreter/ast"
	"github.com/SVendittelli/monkey-interpreter/object"
)

func quote(node ast.Node) object.Object {
	return &object.Quote{Node: node}
}
