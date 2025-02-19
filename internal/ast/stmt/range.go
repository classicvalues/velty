package stmt

import (
	"github.com/viant/velty/internal/ast"
	"github.com/viant/velty/internal/ast/expr"
)

//Range represents regular for loop
type Range struct {
	Init ast.Statement
	Cond ast.Expression
	Body Block
	Post ast.Statement
}

func (r *Range) Statements() []ast.Statement {
	return r.Body.Statements()
}

func (r *Range) AddStatement(statement ast.Statement) {
	r.Body.AddStatement(statement)
}

//ForEach represents for each loop
type ForEach struct {
	Index *expr.Select
	Item  *expr.Select
	Set   *expr.Select
	Body  Block
}

func (f *ForEach) Statements() []ast.Statement {
	return f.Body.Statements()
}

func (f *ForEach) AddStatement(statement ast.Statement) {
	f.Body.AddStatement(statement)
}
