package expressions

import (
	"fmt"
)

type TokenType int

type Token struct {
	Kind    TokenType
	Lexeme  string
	Literal Literal
	Line    int
}

func (tok Token) String() string {
	return fmt.Sprintf("%d %s %v", tok.Kind, tok.Lexeme, tok.Literal.Value)
}

type Experssion interface {
	Accept(ExpressionVisitor) (interface{}, error)
}

type ExpressionVisitor interface {
	VisitBinary(Binary) (interface{}, error)
	VisitGrouping(Grouping) (interface{}, error)
	VisitLiteral(Literal) (interface{}, error)
	VisitUnary(Unary) (interface{}, error)
	VisitVairable(Variable) (interface{}, error)
}

type Binary struct {
	Experssion
	Left     Experssion
	Right    Experssion
	Operator Token
}

type Grouping struct {
	Expr Experssion
}

type Literal struct {
	Experssion
	Value interface{}
}

type Unary struct {
	Experssion
	Right    Experssion
	Operator Token
}

// wrapper around the token for the variable name
type Variable struct {
	Token Token
}

func (g Grouping) Accept(visitor ExpressionVisitor) (interface{}, error) {
	return visitor.VisitGrouping(g)
}

func (b Binary) Accept(visitor ExpressionVisitor) (interface{}, error) {
	return visitor.VisitBinary(b)
}

func (l Literal) Accept(visitor ExpressionVisitor) (interface{}, error) {
	return visitor.VisitLiteral(l)
}

func (u Unary) Accept(visitor ExpressionVisitor) (interface{}, error) {
	return visitor.VisitUnary(u)
}

func (u Variable) Accept(visitor ExpressionVisitor) (interface{}, error) {
	return visitor.VisitVairable(u)
}
