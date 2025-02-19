package velty

import (
	"github.com/viant/velty/internal/est"
	"github.com/viant/velty/internal/est/op"
	"github.com/viant/velty/internal/parser"
	"unsafe"
)

type evaluator struct {
	x       *op.Operand
	cache   *cache
	control est.Control

	parent *Planner
}

func (e *evaluator) compute(state *est.State) unsafe.Pointer {
	varValue := *(*string)(e.x.Exec(state))
	if cacheValue, ok := e.cache.expression(varValue); ok {
		newState, err := e.newState(cacheValue.planner, state)
		if err != nil {
			return est.EmptyStringPtr
		}
		return cacheValue.compute(newState)
	}

	block, err := parser.Parse([]byte(varValue))
	if err != nil {
		return est.EmptyStringPtr
	}

	evaluatorPlanner, err := e.evaluatorPlanner(state)

	if err != nil {
		return est.EmptyStringPtr
	}

	exec, err := evaluatorPlanner.newCompute(block)
	if err != nil {
		return est.EmptyStringPtr
	}

	newState, err := e.newState(evaluatorPlanner, state)
	if err != nil {
		return est.EmptyStringPtr
	}

	e.cache.put(varValue, evaluatorPlanner, exec)

	return exec(newState)
}

func (e *evaluator) evaluatorPlanner(state *est.State) (*Planner, error) {
	evaluatorScope := New()
	scopeType := est.NewType()
	evaluatorScope.Type = scopeType

	var err error
	for _, selector := range e.parent.selectors.Selectors() {
		if selector.Parent != nil {
			continue
		}

		if err = evaluatorScope.DefineVariable(selector.Name, selector.Value(state.MemPtr)); err != nil {
			return nil, err
		}
	}

	return evaluatorScope, nil
}

func (e *evaluator) newState(planner *Planner, state *est.State) (*est.State, error) {
	newState := planner.stateProvider()()

	var err error
	for _, selector := range e.parent.selectors.Selectors() {
		if selector.Parent != nil {
			continue
		}

		if err = newState.SetValue(selector.ID, selector.Value(state.MemPtr)); err != nil {
			return nil, err
		}
	}

	newState.Buffer = state.Buffer
	return newState, nil
}

func evaluate(expr *op.Expression, cache *cache, parent *Planner) (est.New, error) {
	return func(control est.Control) (est.Compute, error) {
		x, err := expr.Operand(control)
		if err != nil {
			return nil, err
		}

		return (&evaluator{
			x:       x,
			cache:   cache,
			control: control,
			parent:  parent,
		}).compute, nil
	}, nil
}
