package interpreter

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/Ahmed-Sermani/prolang/parser/expressions"
	"github.com/Ahmed-Sermani/prolang/parser/statements"
	"github.com/Ahmed-Sermani/prolang/reporting"
	"github.com/Ahmed-Sermani/prolang/scanner"
)

// Runtime errors
type ErrorExpressionInterpretation struct {
	token expressions.Token
	msg   string
}

// implementing the error interface
func (e *ErrorExpressionInterpretation) Error() string {
	if e.msg == "" {
		e.msg = "error while interpreting expressions"
	}
	return e.msg + fmt.Sprintf("[line %d]", e.token.Line)
}

type ErrorOpNumMismatch struct {
	ErrorExpressionInterpretation
}

// implement expression visitor and statement visitor interface
type Interpreter struct{}

func New() *Interpreter {
	return &Interpreter{}
}

func (inter *Interpreter) Interpret(stmts []statements.Statement) error {

	for _, stmt := range stmts {
		err := inter.execute(stmt)
		if err != nil {
			reporting.ReportRuntimeError(err)
			return err
		}
	}

	return nil
}

func (inter *Interpreter) execute(stmt statements.Statement) error {
	return stmt.Accept(inter)
}

// the runtime value was produced during scanning and it in the token.
// The parser took the value and stuck it in the literal tree node,
// so to evaluate a literal pull it back out.
func (inter *Interpreter) VisitLiteral(expr expressions.Literal) (interface{}, error) {
	return expr.Value, nil
}

// grouping node has a reference to an inner node for the expression contained inside the parentheses.
// recursively evaluate that subexpression and return it.
func (inter *Interpreter) VisitGrouping(expr expressions.Grouping) (interface{}, error) {
	return inter.evaluate(expr.Expr)
}

// evaluate the operand expression. Then apply the unary operator itself to the result of that.
// There are two different unary expressions, identified by the type of the operator token.
func (inter *Interpreter) VisitUnary(expr expressions.Unary) (interface{}, error) {
	right, err := inter.evaluate(expr.Right)
	if err != nil {
		return nil, err
	}

	switch expr.Operator.Kind {
	case scanner.MINUS:
		refVal := reflect.ValueOf(right)
		val, err := reflectFloat64(refVal, expr.Operator)
		if err != nil {
			return nil, err
		}
		// negating the right operand in case of minus operator
		return -val, nil
	case scanner.BANG:
		refVal := reflect.ValueOf(right)
		// applying the ! operator
		return !isTruthy(refVal), nil
	}

	return nil, &ErrorExpressionInterpretation{token: expr.Operator}
}

func (inter *Interpreter) VisitBinary(expr expressions.Binary) (interface{}, error) {

	left, err := inter.evaluate(expr.Left)
	if err != nil {
		return nil, &ErrorExpressionInterpretation{token: expr.Operator}
	}
	right, err := inter.evaluate(expr.Right)
	if err != nil {
		return nil, &ErrorExpressionInterpretation{token: expr.Operator}
	}

	switch expr.Operator.Kind {

	// arithmetic operator
	case scanner.MINUS:
		lRefVal := reflect.ValueOf(left)
		lVal, err := reflectFloat64(lRefVal, expr.Operator)
		if err != nil {
			return nil, err
		}
		rRefVal := reflect.ValueOf(right)
		rVal, err := reflectFloat64(rRefVal, expr.Operator)
		if err != nil {
			return nil, err
		}

		return (lVal - rVal), nil

	case scanner.SLASH:
		lRefVal := reflect.ValueOf(left)
		lVal, err := reflectFloat64(lRefVal, expr.Operator)
		if err != nil {
			return nil, err
		}
		rRefVal := reflect.ValueOf(right)
		rVal, err := reflectFloat64(rRefVal, expr.Operator)
		if err != nil {
			return nil, err
		}

		return (lVal / rVal), nil

	case scanner.STAR:
		lRefVal := reflect.ValueOf(left)
		lVal, err := reflectFloat64(lRefVal, expr.Operator)
		if err != nil {
			return nil, err
		}
		rRefVal := reflect.ValueOf(right)
		rVal, err := reflectFloat64(rRefVal, expr.Operator)
		if err != nil {
			return nil, err
		}

		return (lVal * rVal), nil
	// + supports additions on numbers and concatenation on strings
	case scanner.PLUS:
		lRefVal := reflect.ValueOf(left)
		rRefVal := reflect.ValueOf(right)
		if lRefVal.Kind() == reflect.Float64 && rRefVal.Kind() == reflect.Float64 {
			return lRefVal.Float() + rRefVal.Float(), nil
		} else if lRefVal.Kind() == reflect.String && rRefVal.Kind() == reflect.String {
			return lRefVal.String() + rRefVal.String(), nil
		} else {
			return nil, &ErrorOpNumMismatch{
				ErrorExpressionInterpretation{
					token: expr.Operator,
					msg:   "Operands must be two numbers or two strings",
				},
			}
		}
	// comparison operators
	case scanner.GREATER:
		lRefVal := reflect.ValueOf(left)
		lVal, err := reflectFloat64(lRefVal, expr.Operator)
		if err != nil {
			return nil, err
		}
		rRefVal := reflect.ValueOf(right)
		rVal, err := reflectFloat64(rRefVal, expr.Operator)
		if err != nil {
			return nil, err
		}
		return lVal > rVal, nil
	case scanner.GREATER_EQUAL:
		lRefVal := reflect.ValueOf(left)
		lVal, err := reflectFloat64(lRefVal, expr.Operator)
		if err != nil {
			return nil, err
		}
		rRefVal := reflect.ValueOf(right)
		rVal, err := reflectFloat64(rRefVal, expr.Operator)
		if err != nil {
			return nil, err
		}
		return lVal >= rVal, nil
	case scanner.LESS:
		lRefVal := reflect.ValueOf(left)
		lVal, err := reflectFloat64(lRefVal, expr.Operator)
		if err != nil {
			return nil, err
		}
		rRefVal := reflect.ValueOf(right)
		rVal, err := reflectFloat64(rRefVal, expr.Operator)
		if err != nil {
			return nil, err
		}
		return lVal < rVal, nil
	case scanner.LESS_EQUAL:
		lRefVal := reflect.ValueOf(left)
		lVal, err := reflectFloat64(lRefVal, expr.Operator)
		if err != nil {
			return nil, err
		}
		rRefVal := reflect.ValueOf(right)
		rVal, err := reflectFloat64(rRefVal, expr.Operator)
		if err != nil {
			return nil, err
		}
		return lVal <= rVal, nil
	// equality
	case scanner.EQUAL_EQUAL:
		return isEqual(left, right), nil
	case scanner.BANG_EQUAL:
		return !isEqual(left, right), nil
	}

	return nil, &ErrorExpressionInterpretation{token: expr.Operator}
}

func (inter *Interpreter) VisitExprStmt(expr statements.ExperssionStatement) error {
	_, err := inter.evaluate(expr.Expr)
	return err

}

func (inter *Interpreter) VisitPrintStmt(expr statements.PrintStatement) error {
	val, err := inter.evaluate(expr.Expr)
	if err != nil {
		return err
	}
	fmt.Println(stringify(val))
	return err

}

// sends the expression back into the interpreter’s visitor implementation
func (inter *Interpreter) evaluate(expr expressions.Experssion) (interface{}, error) {
	return expr.Accept(inter)
}

// define what to consider true and false
// currently only false and nil considered falsy
func isTruthy(ref reflect.Value) bool {
	if ref.Kind() == reflect.Ptr && ref.IsNil() {
		return false
	}
	if ref.Kind() == reflect.Bool {
		return ref.Bool()
	}
	return true
}

// Interface values are comparable. Two interface values are equal if they have identical dynamic types and equal dynamic values
// or if both have value nil.
func isEqual(left interface{}, right interface{}) bool {
	return left == right
}

func reflectFloat64(ref reflect.Value, operator expressions.Token) (float64, error) {
	err := checkNumOperands(ref, operator)
	if err != nil {
		return 0.0, err
	}
	return ref.Float(), nil
}

func checkNumOperands(operand reflect.Value, operator expressions.Token) error {
	if operand.Kind() != reflect.Float64 {
		return &ErrorOpNumMismatch{
			ErrorExpressionInterpretation{
				token: operator,
				msg:   "Operand must be a number.",
			},
		}
	}
	return nil
}

// used for debugging. convert obj of type interface{} to string
func stringify(obj interface{}) string {
	refVal := reflect.ValueOf(obj)
	if refVal.Kind() == reflect.Ptr && refVal.IsNil() {
		return "nil"
	}
	if refVal.Kind() == reflect.Float64 {
		text := fmt.Sprintf("%f", refVal.Float())

		return text
	}

	if refVal.Kind() == reflect.Bool {
		return strconv.FormatBool(refVal.Bool())
	}

	return refVal.String()
}
