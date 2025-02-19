package velty

import (
	"github.com/viant/velty/internal/ast/expr"
	eexpr "github.com/viant/velty/internal/est/expr"
	"github.com/viant/velty/internal/est/op"
	"reflect"
)

func (p *Planner) compileBinary(actual *expr.Binary) (*op.Expression, error) {
	x, err := p.compileExpr(actual.X)
	if err != nil {
		return nil, err
	}
	y, err := p.compileExpr(actual.Y)
	if err != nil {
		return nil, err
	}

	resultType := nonEmptyType(actual.Type(), x.Type, y.Type)
	acc := p.accumulator(resultType)
	resultExpr := &op.Expression{Selector: acc, Type: acc.Type}

	computeNew, err := eexpr.Binary(actual.Token, x, y, resultExpr)
	if err != nil {
		return nil, err
	}

	return &op.Expression{
		Type: resultType,
		New:  computeNew,
	}, nil
}

func nonEmptyType(types ...reflect.Type) reflect.Type {
	for _, r := range types {
		if r != nil {
			return r
		}
	}

	return nil
}
