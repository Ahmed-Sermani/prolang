package interpreter

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/Ahmed-Sermani/prolang/interpreter/environment"
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

type ObjNotCallable struct {
	ErrorExpressionInterpretation
}

type ArgsNumMismatch struct {
	ErrorExpressionInterpretation
}

// error to handle unwind for return statement
type ErrorHandleReturn struct {
	value interface{}
}

func (e ErrorHandleReturn) Error() string {
	return ""
}

// implement expression visitor and statement visitor interface
// global refer to the outer most global environment
type Interpreter struct {
	environment *environment.Environment
	global      *environment.Environment
}

func New() *Interpreter {
	envPtr := environment.New(nil)
	return &Interpreter{
		environment: envPtr,
		global:      envPtr,
	}
}

type clock struct{}

func (c *clock) ArgsNum() int {
	return 0
}

func (c *clock) Call(inter Interpreter, args []interface{}) (interface{}, error) {
	return float64(time.Now().UnixNano() / int64(time.Second)), nil
}

func (inter *Interpreter) Interpret(stmts []statements.Statement) error {
	inter.environment.Define("clock", clock{})
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

func (inter *Interpreter) VisitVairable(expr expressions.Variable) (interface{}, error) {
	return inter.environment.Get(expr.Token)
}

// evaluates the r-value.
// then stores it in the named variable
func (inter *Interpreter) VisitAssgin(expr expressions.Assgin) (interface{}, error) {
	value, err := inter.evaluate(expr.Value)
	if err != nil {
		return nil, err
	}
	inter.environment.Assgin(expr.Token, value)
	return value, nil
}

func (inter *Interpreter) VisitLogical(expr expressions.Logical) (interface{}, error) {
	left, err := inter.evaluate(expr.Left)
	if err != nil {
		return nil, err
	}
	refVal := reflect.ValueOf(left)
	if expr.Operator.Kind == scanner.OR && isTruthy(refVal) {
		return left, nil
	} else if !isTruthy(refVal) {
		return left, nil
	}

	return inter.evaluate(expr.Right)
}

func (inter *Interpreter) VisitCall(expr expressions.Call) (interface{}, error) {
	callee, err := inter.evaluate(expr.Callee)
	if err != nil {
		return nil, err
	}
	args := []interface{}{}
	for _, arg := range expr.Args {
		evaluatedArg, err := inter.evaluate(arg)
		if err != nil {
			return nil, err
		}
		args = append(args, evaluatedArg)

	}
	function, ok := callee.(Callable)
	if !ok {
		return nil, &ObjNotCallable{
			ErrorExpressionInterpretation: ErrorExpressionInterpretation{
				msg: fmt.Sprintf("Object %s is not callable", callee),
			},
		}
	}

	if len(args) != function.ArgsNum() {
		return nil, &ObjNotCallable{
			ErrorExpressionInterpretation: ErrorExpressionInterpretation{
				msg: fmt.Sprintf("Function %s does not have the correct number of arguments", callee),
			},
		}
	}
	return function.Call(inter, args)

}

func (inter *Interpreter) VisitExprStmt(stmt statements.ExperssionStatement) error {
	_, err := inter.evaluate(stmt.Expr)
	return err

}

func (inter *Interpreter) VisitPrintStmt(stmt statements.PrintStatement) error {
	val, err := inter.evaluate(stmt.Expr)
	if err != nil {
		return err
	}
	fmt.Println(stringify(val))
	return err

}

// evaluate the initializer of exists
// vairable decleration without initializer is allowed
func (inter *Interpreter) VisitVarDecStmt(stmt statements.VarDecStatement) error {
	var value interface{}
	if stmt.Initializer != nil {
		// evaluate initializer expression
		tvalue, err := inter.evaluate(stmt.Initializer)
		if err != nil {
			return err
		}
		value = tvalue
	}
	inter.environment.Define(stmt.Token.Lexeme, value)
	return nil
}

func (inter *Interpreter) VisitBlockStmt(stmt statements.BlockStatement) error {
	// passing the current environment into the enclosing state of sub-scope
	return inter.executeBlock(stmt.Statements, environment.New(inter.environment))
}

// evaluates the condition. If truthy, executes the then branch.
// Otherwise, if there is an else branch, executes that.
func (inter *Interpreter) VisitIfStmt(stmt statements.IfStatement) error {
	conditionValue, err := inter.evaluate(stmt.Condition)
	if err != nil {
		return err
	}
	if isTruthy(reflect.ValueOf(conditionValue)) {
		err := inter.execute(stmt.ThenBranch)
		if err != nil {
			return err
		}
	} else if stmt.ElseBranch != nil {
		err := inter.execute(stmt.ElseBranch)
		if err != nil {
			return err
		}
	}
	return nil
}

// evaluated the condition . while it's truthy execute the body
func (inter *Interpreter) VisitWhileStmt(stmt statements.WhileStatement) error {
	for {
		condiction, err := inter.evaluate(stmt.Condition)
		if err != nil {
			return err
		}
		if !isTruthy(reflect.ValueOf(condiction)) {
			break
		}
		err1 := inter.execute(stmt.Body)
		if err1 != nil {
			return nil
		}
	}
	return nil
}

// convert function complie time representation to its runtime representation
func (inter *Interpreter) VisitFunctionStmt(stmt statements.FunctionStatement) error {
	// after creating the FunctionCallable,
	// it create a new binding in the current environment and store a reference to it there.
	// binding as *FunctionCallable because FunctionCallable implements Callable interface as pointer receiver
	function := &FunctionCallable{Declaration: stmt, Closure: inter.environment}
	inter.environment.Define(stmt.Name.Lexeme, function)
	return nil
}

func (inter *Interpreter) VisitReturnStmt(stmt statements.ReturnStatement) error {
	var value interface{}
	if stmt.Value != nil {
		value1, err := inter.evaluate(stmt.Value)
		if err != nil {
			return err
		}
		value = value1
	}
	return ErrorHandleReturn{value: value}
}

func (inter *Interpreter) executeBlock(stmts []statements.Statement, innerEnv *environment.Environment) error {
	// save the outer env
	outerEnv := inter.environment
	// replace the outer env with the inner one
	inter.environment = innerEnv
	// defer env restore That way it gets restored even if an error occurs.
	defer func() {
		inter.environment = outerEnv
	}()
	for _, stmt := range stmts {
		err := inter.execute(stmt)
		if err != nil {
			return err
		}
	}
	return nil

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
