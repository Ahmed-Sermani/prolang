package parser

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/Ahmed-Sermani/prolang/parser/expressions"
	"github.com/Ahmed-Sermani/prolang/parser/statements"
	"github.com/Ahmed-Sermani/prolang/reporting"
	"github.com/Ahmed-Sermani/prolang/scanner"
)

var ErrorParsing = errors.New("error while parsing expressions")
var ErrorPrinterVisitor = errors.New("error in printer visitor")
var ErrorInvalidAssginTarget = errors.New("error invalid assgin target")

type Parser struct {
	tokens  []expressions.Token
	current int
}

func New(tokens []expressions.Token) *Parser {

	return &Parser{
		tokens:  tokens,
		current: 0,
	}
}

// start parsing
// prog           → statement* EOF ;
func (p *Parser) Parse() []statements.Statement {
	statements := []statements.Statement{}
	for !p.isAtEnd() {
		dec := p.declaration()
		statements = append(statements, dec)
	}
	return statements
}

func (p *Parser) block() ([]statements.Statement, error) {
	stmts := []statements.Statement{}

	for !p.check(scanner.RIGHT_BRACE) && !p.isAtEnd() {
		stmt := p.declaration()
		stmts = append(stmts, stmt)
	}
	_, err := p.consume(scanner.RIGHT_BRACE, "Expect '}' after block.")
	if err != nil {
		return stmts, err
	}

	return stmts, nil
}

// declaration     → varDeclaration | statement | funcDeclaration ;
func (p *Parser) declaration() statements.Statement {
	var err error
	if p.match(scanner.FUNC) {
		stmt, err := p.function("function")
		if err != nil {
			p.synchronize()
			return nil
		}
		return stmt
	}
	if p.match(scanner.LET) {
		stmt, err := p.varDeclaration()
		// recovery
		if err != nil {
			p.synchronize()
			return nil
		}
		return stmt
	}
	stmt, err := p.statement()

	// recovery
	if err != nil {
		p.synchronize()
		return nil
	}
	return stmt

}

func (p *Parser) function(kind string) (statements.Statement, error) {
	name, err := p.consume(scanner.IDENTIFIER, "Expect "+kind+" name.")
	if err != nil {
		return nil, err
	}
	_, err2 := p.consume(scanner.LEFT_PAREN, "Expect '(' after "+kind+" name.")
	if err2 != nil {
		return nil, err2
	}
	paramenters := []expressions.Token{}

	if !p.check(scanner.RIGHT_PAREN) {
		for {
			param, err := p.consume(scanner.IDENTIFIER, "Expect parameter name.")
			if err != nil {
				return nil, err
			}
			paramenters = append(paramenters, param)
			if !p.match(scanner.COMMA) {
				break
			}
		}
	}
	_, err1 := p.consume(scanner.RIGHT_PAREN, "Expect ')' after parameters")
	if err1 != nil {
		return nil, err1
	}
	_, err3 := p.consume(scanner.LEFT_BRACE, "Expect '{' before "+kind+" body.")
	if err3 != nil {
		return nil, err3
	}
	body, err := p.block()
	if err != nil {
		return nil, err
	}

	return statements.FunctionStatement{
		Name: name,
		Args: paramenters,
		Body: body,
	}, nil
}

func (p *Parser) varDeclaration() (statements.Statement, error) {
	name, err := p.consume(scanner.IDENTIFIER, "Expect variable name.")
	if err != nil {
		return nil, err
	}
	var initializer expressions.Experssion
	if p.match(scanner.EQUAL) {
		initializer, err = p.experssion()
		if err != nil {
			return nil, err
		}
	}
	_, err1 := p.consume(scanner.SEMICOLON, "Expect ';' after variable declaration.")
	if err1 != nil {
		return nil, err1
	}
	return statements.VarDecStatement{
		Token:       name,
		Initializer: initializer,
	}, nil

}

// statement       → exprStatement | printStatement | block | ifStatement | whileStatement | forStatement | returnStatement;
func (p *Parser) statement() (statements.Statement, error) {
	if p.match(scanner.FOR) {
		return p.forStatement()
	}
	if p.match(scanner.IF) {
		return p.ifStatement()
	}
	if p.match(scanner.PRINT) {
		return p.printStatement()
	}
	if p.match(scanner.RETURN) {
		return p.returnStatement()
	}
	if p.match(scanner.WHILE) {
		return p.whileStatement()
	}
	if p.match(scanner.LEFT_BRACE) {
		stmts, err := p.block()
		return statements.BlockStatement{Statements: stmts}, err
	}

	return p.experssionStatement()
}

// returnStatement → "return" expression? ";" ;
func (p *Parser) returnStatement() (statements.Statement, error) {
	keyword := p.previous()
	var value expressions.Experssion
	if !p.check(scanner.SEMICOLON) {
		value1, err := p.experssion()
		if err != nil {
			return nil, err
		}
		value = value1
	}
	_, err := p.consume(scanner.SEMICOLON, "Expext ';' after return value.")
	if err != nil {
		return nil, err
	}
	return statements.ReturnStatement{
		Keyword: keyword,
		Value:   value,
	}, nil
}

// forStatement   → "for" "(" ( varDecl | exprStmt | ";" ) expression? ";" expression? ")" statement ;
func (p *Parser) forStatement() (statements.Statement, error) {
	_, err := p.consume(scanner.LEFT_PAREN, "Expect '(' after 'for'")
	if err != nil {
		return nil, err
	}
	var initializer statements.Statement

	// variable declaration
	if p.match(scanner.LET) {
		dec, err := p.varDeclaration()
		initializer = dec
		if err != nil {
			return nil, err
		}
		// the initializer is an expression
	} else {
		expr, err := p.experssionStatement()
		initializer = expr
		if err != nil {
			return nil, err
		}

	}

	var condition expressions.Experssion
	// if the condition omitted
	if !p.check(scanner.SEMICOLON) {
		condition, err = p.experssion()
		if err != nil {
			return nil, err
		}
	}
	_, err1 := p.consume(scanner.SEMICOLON, "Expect ';' after loop condition")
	if err1 != nil {
		return nil, err1
	}

	var increment expressions.Experssion
	// check if the increment part is omitted
	if !p.check(scanner.RIGHT_PAREN) {
		increment, err = p.experssion()
		if err != nil {
			return nil, err
		}
	}
	_, err2 := p.consume(scanner.RIGHT_PAREN, "Expect ')' after for clauses.")
	if err2 != nil {
		return nil, err2
	}

	body, err := p.statement()
	if err != nil {
		return nil, err
	}

	if increment != nil {
		// replacing the body with Block statement that has the increment appended at the end
		body = statements.BlockStatement{
			Statements: []statements.Statement{
				body,
				statements.ExperssionStatement{Expr: increment},
			},
		}

	}

	if condition == nil {
		// if the condition is not set. set it as true literal
		condition = expressions.Literal{Value: true}
	}

	// implement for loop as syntactic sugar of while loop
	body = statements.WhileStatement{
		Condition: condition,
		Body:      body,
	}

	// if the initializer is set. using Block statement. set the initializer as the first statement
	if initializer != nil {
		body = statements.BlockStatement{
			Statements: []statements.Statement{initializer, body},
		}
	}

	return body, nil
}

// whileStmt      → "while" "(" expression ")" statement ;
func (p *Parser) whileStatement() (statements.Statement, error) {
	_, err := p.consume(scanner.LEFT_PAREN, "Expect '(' after 'while'")
	if err != nil {
		return nil, err
	}
	condition, err := p.experssion()
	if err != nil {
		return nil, err
	}
	_, err1 := p.consume(scanner.RIGHT_PAREN, "Expect ')' after condition")
	if err1 != nil {
		return nil, err1
	}
	body, err := p.statement()
	if err != nil {
		return nil, err
	}

	return statements.WhileStatement{
		Condition: condition,
		Body:      body,
	}, nil
}

// ifStatement    → "if" "(" expression ")" statement ( "else" statement )? ;
func (p *Parser) ifStatement() (statements.Statement, error) {
	_, err := p.consume(scanner.LEFT_PAREN, "Expected '(' after 'if'")
	if err != nil {
		return nil, err
	}
	condition, err := p.experssion()
	if err != nil {
		return nil, err
	}

	_, err1 := p.consume(scanner.RIGHT_PAREN, "Expected ')' after condition")
	if err1 != nil {
		return nil, err
	}
	thenBranch, err1 := p.statement()

	if err1 != nil {
		return nil, err
	}

	var elseBranch statements.Statement
	if p.match(scanner.ELSE) {
		elseBranch, err = p.statement()
		if err != nil {
			return nil, err
		}
	}

	return statements.IfStatement{
		Condition:  condition,
		ThenBranch: thenBranch,
		ElseBranch: elseBranch,
	}, nil

}

// exprStatement  → expression ";" ;
func (p *Parser) printStatement() (statements.Statement, error) {
	val, err := p.experssion()
	if err != nil {
		return nil, err
	}
	_, err1 := p.consume(scanner.SEMICOLON, "Expected ';' after expression.")
	if err1 != nil {
		return nil, err
	}
	return statements.PrintStatement{Expr: val}, nil
}

// exprStatement  → expression ";" ;
func (p *Parser) experssionStatement() (statements.Statement, error) {
	val, err := p.experssion()
	if err != nil {
		return nil, err
	}
	_, err1 := p.consume(scanner.SEMICOLON, "Expected ';' after value.")
	if err1 != nil {
		return nil, err
	}
	return statements.ExperssionStatement{Expr: val}, nil
}

// expression     → assignment ;
func (p *Parser) experssion() (expressions.Experssion, error) {
	return p.assignment()
}

// assignment     → IDENTIFIER "=" assignment | logicalOr ;
func (p *Parser) assignment() (expressions.Experssion, error) {
	expr, err := p.logicalOr()
	if err != nil {
		return nil, err
	}
	if p.match(scanner.EQUAL) {
		equals := p.previous()

		// since assignment is right associative, then recursively call assignment() to parse the r-value.
		val, err := p.assignment()
		if err != nil {
			return nil, err
		}

		// look at the left-hand side expression and figure out what kind of assignment target it is
		// convert the r-value expression node into an l-value representation
		if varExpr, ok := expr.(expressions.Variable); ok {
			return expressions.Assgin{Token: varExpr.Token, Value: val}, nil
		}

		reporting.ReportError(equals.Line, ErrorInvalidAssginTarget.Error())
		return nil, ErrorInvalidAssginTarget
	}

	return expr, nil

}

// logicalOr      → logicalAnd ( "or" logicalAnd )* ;
func (p *Parser) logicalOr() (expressions.Experssion, error) {
	expr, err := p.logicalAnd()
	if err != nil {
		return nil, err
	}

	for p.match(scanner.OR) {
		op := p.previous()
		right, err := p.logicalAnd()
		if err != nil {
			return nil, nil
		}
		expr = expressions.Logical{
			Right:    right,
			Left:     expr,
			Operator: op,
		}
	}
	return expr, nil
}

// logicalAnd     → equality ( "and" equality )* ;
func (p *Parser) logicalAnd() (expressions.Experssion, error) {
	expr, err := p.equality()
	if err != nil {
		return nil, err
	}
	for p.match(scanner.AND) {
		op := p.previous()
		right, err := p.equality()
		if err != nil {
			return nil, err
		}
		expr = expressions.Logical{
			Right:    right,
			Left:     expr,
			Operator: op,
		}
	}
	return expr, nil
}

// equality       → comparison ( ( "!=" | "==" ) comparison )* ;
func (p *Parser) equality() (expressions.Experssion, error) {
	expr, err := p.comparison()
	if err != nil {
		return expressions.Binary{}, err
	}

	for p.match(scanner.BANG_EQUAL, scanner.EQUAL_EQUAL) {
		operator := p.previous()
		right, err := p.comparison()
		if err != nil {
			return expressions.Binary{}, err
		}
		expr = expressions.Binary{
			Left:     expr,
			Right:    right,
			Operator: operator,
		}

	}
	return expr, err

}

// comparison     → term ( ( ">" | ">=" | "<" | "<=" ) term )* ;
func (p *Parser) comparison() (expressions.Experssion, error) {
	expr, err := p.term()
	if err != nil {
		return expressions.Binary{}, err
	}

	for p.match(scanner.GREATER, scanner.GREATER_EQUAL, scanner.LESS, scanner.LESS_EQUAL) {
		operator := p.previous()
		right, err := p.term()
		if err != nil {
			return expressions.Binary{}, err
		}
		expr = expressions.Binary{
			Left:     expr,
			Right:    right,
			Operator: operator,
		}
	}
	return expr, err
}

// term           → factor ( ( "-" | "+" ) factor )* ;
func (p *Parser) term() (expressions.Experssion, error) {
	expr, err := p.factor()
	if err != nil {
		return expressions.Binary{}, err
	}

	for p.match(scanner.PLUS, scanner.MINUS) {
		operator := p.previous()
		right, err := p.factor()
		if err != nil {
			return expressions.Binary{}, err
		}
		expr = expressions.Binary{
			Left:     expr,
			Right:    right,
			Operator: operator,
		}
	}
	return expr, err
}

// factor         → unary ( ( "/" | "*" ) unary )* ;
func (p *Parser) factor() (expressions.Experssion, error) {
	expr, err := p.unary()
	if err != nil {
		return expressions.Binary{}, err
	}

	for p.match(scanner.SLASH, scanner.STAR) {
		operator := p.previous()
		right, err := p.unary()
		if err != nil {
			return expressions.Binary{}, err
		}
		expr = expressions.Binary{
			Left:     expr,
			Right:    right,
			Operator: operator,
		}
	}
	return expr, err
}

// unary          → ( "!" | "-" ) unary | call ;
func (p *Parser) unary() (expressions.Experssion, error) {
	if p.match(scanner.BANG, scanner.MINUS) {
		operator := p.previous()
		right, err := p.unary()
		if err != nil {
			return expressions.Unary{}, err
		}
		return expressions.Unary{
			Operator: operator,
			Right:    right,
		}, nil
	}
	return p.call()
}

// call           → primary ( "(" arguments? ")" )* ;
func (p *Parser) call() (expressions.Experssion, error) {
	expr, err := p.primary()
	if err != nil {
		return nil, err
	}

	for {
		if p.match(scanner.LEFT_PAREN) {
			expr1, err := p.followCall(expr)
			expr = expr1
			if err != nil {
				return nil, err
			}
		} else {
			break
		}
	}

	return expr, nil
}

func (p *Parser) followCall(callee expressions.Experssion) (expressions.Experssion, error) {
	args := []expressions.Experssion{}
	// handle zero args case
	if !p.check(scanner.RIGHT_PAREN) {
		for {
			expr, err := p.experssion()
			if err != nil {
				return nil, err
			}
			args = append(args, expr)
			if !p.match(scanner.COMMA) {
				break
			}
		}
	}

	parenth, err := p.consume(scanner.RIGHT_PAREN, "Expect ')' after argument")
	if err != nil {
		return nil, err
	}
	return expressions.Call{
		Callee:  callee,
		Args:    args,
		Parenth: parenth,
	}, nil
}

// primary        → NUMBER | STRING | "true" | "false" | "nil" | "(" expression ")" ;
func (p *Parser) primary() (expressions.Experssion, error) {
	switch {
	case p.match(scanner.FALSE):
		return expressions.Literal{Value: false}, nil
	case p.match(scanner.TRUE):
		return expressions.Literal{Value: true}, nil
	case p.match(scanner.NIL):
		return expressions.Literal{Value: nil}, nil
	case p.match(scanner.NUMBER, scanner.STRING):
		return p.previous().Literal, nil
	case p.match(scanner.LEFT_PAREN):
		{
			expr, err1 := p.experssion()
			if err1 != nil {
				return expressions.Grouping{}, err1
			}
			_, err := p.consume(scanner.RIGHT_PAREN, "Expect ')' after expression.")
			if err != nil {
				return expressions.Grouping{}, err
			}
			return expressions.Grouping{Expr: expr}, nil
		}
	case p.match(scanner.IDENTIFIER):
		return expressions.Variable{Token: p.previous()}, nil
	}
	return expressions.Grouping{}, ErrorParsing
}

// checks the next token against expected token type if not match then report syntax error
// with its respective location
func (p *Parser) consume(tokenType expressions.TokenType, msg string) (expressions.Token, error) {
	if p.check(tokenType) {
		return p.advance(), nil
	}
	if p.peek().Kind == scanner.EOF {
		msg = fmt.Sprintf(" at end %s", msg)
		reporting.ReportError(p.peek().Line, msg)
	} else {
		msg = fmt.Sprintf(" at '%s' %s", p.peek().Lexeme, msg)
		reporting.ReportError(p.peek().Line, msg)
	}

	return expressions.Token{}, ErrorParsing
}

// discard tokens until the beginning of the next statement
func (p *Parser) synchronize() {
	p.advance()

	for !p.isAtEnd() {
		if p.previous().Kind == scanner.SEMICOLON {
			return
		}

		switch p.peek().Kind {
		case scanner.CLASS:
			return
		case scanner.FUNC:
			return
		case scanner.LET:
			return
		case scanner.FOR:
			return
		case scanner.IF:
			return
		case scanner.WHILE:
			return
		case scanner.PRINT:
			return
		case scanner.RETURN:
			return
		}
		p.advance()
	}
}

// matches the next token and consume it if match
func (p *Parser) match(tokenType ...expressions.TokenType) bool {
	for _, v := range tokenType {
		if p.check(v) {
			p.advance()
			return true
		}
	}

	return false
}

// checks the current peak token against token type
func (p *Parser) check(tokenType expressions.TokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.peek().Kind == tokenType
}

// consume the current token and return it
func (p *Parser) advance() expressions.Token {
	if !p.isAtEnd() {
		p.current++
	}
	return p.previous()
}

// peaks on the current token
func (p *Parser) peek() expressions.Token {
	return p.tokens[p.current]
}

// is EOF token
func (p *Parser) isAtEnd() bool {
	return p.peek().Kind == scanner.EOF
}

// gets the previous token
func (p *Parser) previous() expressions.Token {
	return p.tokens[p.current-1]
}

// Debug Visitor
// implements the expressions.ExpressionVisitor interface
type PrintVisitor struct{}

func (pv PrintVisitor) Print(expr expressions.Experssion) (string, error) {
	v, err := expr.Accept(pv)
	if err != nil {
		return "", err
	}
	refVal := reflect.ValueOf(v)
	if refVal.Kind() != reflect.String {
		return "", ErrorPrinterVisitor
	}

	return refVal.String(), nil

}

// visit binary expression
func (pv PrintVisitor) VisitBinary(expr expressions.Binary) (interface{}, error) {
	return pv.parenthesize(expr.Operator.Lexeme, expr.Left, expr.Right)
}

// visit grouping expression
func (pv PrintVisitor) VisitGrouping(expr expressions.Grouping) (interface{}, error) {
	return pv.parenthesize("group", expr.Expr)
}

// visit literal and reflect its value
func (pv PrintVisitor) VisitLiteral(expr expressions.Literal) (interface{}, error) {
	if expr.Value == nil {
		return "nil", nil
	}
	return expr.Value, nil

}

// visit unary expression
func (pv PrintVisitor) VisitUnary(expr expressions.Unary) (interface{}, error) {
	return pv.parenthesize(expr.Operator.Lexeme, expr.Right)
}

// not implemented
func (pv PrintVisitor) VisitVairable(expr expressions.Variable) (interface{}, error) {
	return nil, nil
}

// not implemented
func (pv PrintVisitor) VisitAssgin(expr expressions.Assgin) (interface{}, error) {
	return nil, nil
}

// not implemented
func (pv PrintVisitor) VisitLogical(expr expressions.Logical) (interface{}, error) {
	return nil, nil
}

// not implemented
func (pv PrintVisitor) VisitCall(expr expressions.Call) (interface{}, error) {
	return nil, nil
}

// stringify the expressions into single string builder and return its accumulated string.
// uses reflection to reflect the expressions value:
// output e.g. (+ 2 3)
func (pv PrintVisitor) parenthesize(name string, expr ...expressions.Experssion) (string, error) {
	var sb strings.Builder
	sb.WriteString("(")
	sb.WriteString(name)

	for _, v := range expr {
		sb.WriteString(" ")
		v, err := v.Accept(pv)
		if err != nil {
			return "", ErrorPrinterVisitor
		}
		refValue := reflect.ValueOf(v)
		if refValue.Kind() == reflect.Float64 {
			val := refValue.Float()
			s := fmt.Sprintf("%f", val)
			sb.WriteString(s)
		} else {
			sb.WriteString(refValue.String())
		}

	}
	sb.WriteString(")")

	return sb.String(), nil
}
