package parser

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/Ahmed-Sermani/prolang/parser/expressions"
	"github.com/Ahmed-Sermani/prolang/reporting"
	"github.com/Ahmed-Sermani/prolang/scanner"
)

var ErrorParsing = errors.New("error while parsing expressions")
var ErrorPrinterVisitor = errors.New("error in printer visitor")

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
func (p *Parser) Parse() (expressions.Experssion, error) {
	return p.experssion()
}

// expression     → equality ;
func (p *Parser) experssion() (expressions.Experssion, error) {
	return p.equality()
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

// unary          → ( "!" | "-" ) unary | primary
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
	return p.primary()
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
func (p *Parser) Synchronize() {
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
