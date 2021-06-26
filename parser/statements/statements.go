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
	VisitIfStmt(IfStatement) error
	VisitWhileStmt(WhileStatement) error
	VisitFunctionStmt(FunctionStatement) error
	VisitReturnStmt(ReturnStatement) error
	VisitClassStmt(ClassStatement) error
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

type IfStatement struct {
	Condition  expressions.Experssion
	ThenBranch Statement
	ElseBranch Statement
}

func (i IfStatement) Accept(visitor StatementVisitor) error {
	return visitor.VisitIfStmt(i)
}

type WhileStatement struct {
	Condition expressions.Experssion
	Body      Statement
}

func (w WhileStatement) Accept(visitor StatementVisitor) error {
	return visitor.VisitWhileStmt(w)
}

type FunctionStatement struct {
	Name expressions.Token
	Args []expressions.Token
	Body []Statement
}

func (f FunctionStatement) Accept(visitor StatementVisitor) error {
	return visitor.VisitFunctionStmt(f)
}

// It stores the return keyword token for error reporting if needed, and the value being returned
type ReturnStatement struct {
	Keyword expressions.Token
	Value   expressions.Experssion
}

func (r ReturnStatement) Accept(visitor StatementVisitor) error {
	return visitor.VisitReturnStmt(r)
}

type ClassStatement struct {
	Name    expressions.Token
	Methods []FunctionStatement
}

func (c ClassStatement) Accept(visitor StatementVisitor) error {
	return visitor.VisitClassStmt(c)
}
