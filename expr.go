package velty

import (
	"fmt"
	"github.com/viant/velty/internal/ast"
	aexpr "github.com/viant/velty/internal/ast/expr"
	"github.com/viant/velty/internal/est/op"
)

func (p *Planner) compileExpr(e ast.Expression) (*op.Expression, error) {
	switch actual := e.(type) {
	case *aexpr.Literal:
		return p.literalExpr(actual)
	case *aexpr.Select:
		return p.selectorExpr(actual)
	case *aexpr.Binary:
		return p.compileBinary(actual)
	case *aexpr.Parentheses:
		return p.compileExpr(actual.P)
	}

	return nil, fmt.Errorf("unsupported expr: %T", e)
}
