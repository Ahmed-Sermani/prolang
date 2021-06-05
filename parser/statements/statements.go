package statements

import "github.com/Ahmed-Sermani/prolang/parser/expressions"

type Statement interface {
	Accept(StatementVisitor) error
}

type StatementVisitor interface {
	VisitPrintStmt(PrintStatement) error
	VisitExprStmt(ExperssionStatement) error
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

func (p ExperssionStatement) Accept(visitor StatementVisitor) error {
	return visitor.VisitExprStmt(p)
}