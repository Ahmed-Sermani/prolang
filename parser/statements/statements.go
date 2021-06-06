package statements

import "github.com/Ahmed-Sermani/prolang/parser/expressions"

type Statement interface {
	Accept(StatementVisitor) error
}

type StatementVisitor interface {
	VisitPrintStmt(PrintStatement) error
	VisitExprStmt(ExperssionStatement) error
	VisitVarDecStmt(VarDecStatement) error
	VisitBlockStmt(BlockStatement) error
}

type PrintStatement struct {
	Statement
	Expr expressions.Experssion
}

func (p PrintStatement) Accept(visitor StatementVisitor) error {
	return visitor.VisitPrintStmt(p)
}

type ExperssionStatement struct {
	Statement
	Expr expressions.Experssion
}

func (e ExperssionStatement) Accept(visitor StatementVisitor) error {
	return visitor.VisitExprStmt(e)
}

type VarDecStatement struct {
	Token       expressions.Token
	Initializer expressions.Experssion
}

func (v VarDecStatement) Accept(visitor StatementVisitor) error {
	return visitor.VisitVarDecStmt(v)
}

type BlockStatement struct {
	Statements []Statement
}

func (s BlockStatement) Accept(visitor StatementVisitor) error {
	return visitor.VisitBlockStmt(s)
}
