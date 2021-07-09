package resolver

import (
	"fmt"

	"github.com/Ahmed-Sermani/prolang/interpreter"
	"github.com/Ahmed-Sermani/prolang/interpreter/callableenum"
	"github.com/Ahmed-Sermani/prolang/interpreter/classenum"
	"github.com/Ahmed-Sermani/prolang/parser/expressions"
	"github.com/Ahmed-Sermani/prolang/parser/statements"
	"github.com/Ahmed-Sermani/prolang/reporting"
)

type scope map[string]bool
type stack []scope

// static analysis.
// resolve variables to solve the problem of dynamic changing environments
// that apperas with closures.
// after the parser produces the syntax tree,
// but before the interpreter starts executing, it’ll do a single walk over the tree to resolve all of the variables it contains
// aka persisting static scope.
// implements statements and expressions visitor interface.
type Resolver struct {
	inter  *interpreter.Interpreter
	scopes stack
	// track if it's in a function or global scope
	curft int
	// track if it's in a method in class
	curcls int
}

func New(inter *interpreter.Interpreter) *Resolver {
	return &Resolver{
		inter: inter,
		// track the stack of scopes currently in scope.
		// Each element in the stack is a Map representing a single block scope.
		// Keys are variable names. The values are Booleans.
		// the boolean value to track the definition.
		// false if the variable declared but not yet defined, true if both
		scopes: stack{},
		// set default start scope type as global scope
		curft: callableenum.NONE,
	}
}

// resolving blocks
func (resolver *Resolver) VisitBlockStmt(stmt statements.BlockStatement) error {
	resolver.beginScope()
	resolver.Resolve(stmt.Statements)
	resolver.endScope()
	return nil
}

// walks the statements and resolves each one.
func (resolver *Resolver) Resolve(stmts []statements.Statement) {
	for _, stmt := range stmts {
		resolver.resolveStmt(stmt)
	}
}

// apply the statement visitor
func (resolver *Resolver) resolveStmt(stmt statements.Statement) {
	stmt.Accept(resolver)

}

// apply the expression visitor
func (resolver *Resolver) resolveExpr(expr expressions.Experssion) {
	expr.Accept(resolver)
}

func (resolver *Resolver) VisitVarDecStmt(stmt statements.VarDecStatement) error {
	resolver.declare(stmt.Token)
	if stmt.Initializer != nil {
		resolver.resolveExpr(stmt.Initializer)
	}
	resolver.define(stmt.Token)
	return nil
}

func (resolver *Resolver) VisitVairable(expr expressions.Variable) (interface{}, error) {
	// if the scopes is not empty and the variable is not defined
	if len(resolver.scopes) != 0 {
		flag, exists := resolver.scopes[len(resolver.scopes)-1][expr.Token.Lexeme]
		if !flag && exists {
			// Make it an error if reference a variable in its initializer
			// e.g. let a = 8;
			// let a = a;
			reporting.ReportError(expr.Token.Line, fmt.Sprintf("Can't read local variable %s in its own initializer", expr.Token.Lexeme))
			return nil, nil
		}
	}
	resolver.resolveLocalVar(expr, expr.Token.Lexeme)
	return nil, nil
}

// First, it resolve the expression for the assigned value in case it also contains references to other variables.
// Then it resolve the variable that’s being assigned to.
func (resolver *Resolver) VisitAssgin(expr expressions.Assgin) (interface{}, error) {
	resolver.resolveExpr(expr.Value)
	resolver.resolveLocalVar(expr, expr.Token.Lexeme)
	return nil, nil
}

func (resolver *Resolver) VisitThis(expr expressions.This) (interface{}, error) {

	// check if 'this' is used outside of a method body
	if resolver.curcls == classenum.NONE {
		reporting.ReportError(expr.Keywork.Line, "Can't use 'this' keyword outside of a class")
		return nil, nil
	}
	resolver.resolveLocalVar(expr, expr.Keywork.Lexeme)
	return nil, nil
}
func (resolver *Resolver) VisitFunctionStmt(stmt statements.FunctionStatement) error {
	// define the declare the function before resolving to let the function refer to itself
	resolver.declare(stmt.Name)
	resolver.define(stmt.Name)
	resolver.resolveFunction(stmt, callableenum.FUNCTION)
	return nil
}

func (resolver *Resolver) VisitExprStmt(stmt statements.ExperssionStatement) error {
	resolver.resolveExpr(stmt.Expr)
	return nil
}

func (resolver *Resolver) VisitIfStmt(stmt statements.IfStatement) error {
	resolver.resolveExpr(stmt.Condition)
	resolver.resolveStmt(stmt.ThenBranch)
	if stmt.ElseBranch != nil {
		resolver.resolveStmt(stmt.ElseBranch)
	}
	return nil
}

func (resolver *Resolver) VisitPrintStmt(stmt statements.PrintStatement) error {
	resolver.resolveExpr(stmt.Expr)
	return nil
}

func (resolver *Resolver) VisitReturnStmt(stmt statements.ReturnStatement) error {
	// disallow return out side function scope type
	if resolver.curft == callableenum.NONE {
		reporting.ReportError(stmt.Keyword.Line, "Return is not allowed outside of a function")
	} else if resolver.curft == callableenum.INITIALIZER {
		reporting.ReportError(stmt.Keyword.Line, "Can't use 'return' in an initializer")
	}
	if stmt.Value != nil {
		resolver.resolveExpr(stmt.Value)
	}
	return nil
}

func (resolver *Resolver) VisitWhileStmt(stmt statements.WhileStatement) error {
	resolver.resolveExpr(stmt.Condition)
	resolver.resolveStmt(stmt.Body)
	return nil
}

func (resolver *Resolver) VisitClassStmt(stmt statements.ClassStatement) error {
	curcls := resolver.curcls
	resolver.curcls = classenum.CLASS
	defer func() {
		resolver.curcls = curcls
	}()
	// defining class directly to allow the methods to reference it's class
	resolver.declare(stmt.Name)
	resolver.define(stmt.Name)

	// detect if the class try to extends itself
	if stmt.Superclass.Token.Lexeme == stmt.Name.Lexeme {
		reporting.ReportError(stmt.Name.Line, "Class can't extends itself")
	}

	// resolve the superclass if exists
	if stmt.Superclass.Token.Lexeme != "" {
		resolver.curcls = classenum.SUPCLASS
		resolver.resolveExpr(stmt.Superclass)
	}

	// If the class declaration has a superclass,
	// create a new scope surrounding all of its methods. In that scope, it define "super"
	if stmt.Superclass.Token.Lexeme != "" {
		resolver.beginScope()
		scope := resolver.scopes[len(resolver.scopes)-1]
		resolver.scopes = resolver.scopes[:len(resolver.scopes)-1]
		scope["super"] = true
		resolver.scopes = append(resolver.scopes, scope)
	}

	// define “this”
	resolver.beginScope()
	scope := resolver.scopes[len(resolver.scopes)-1]
	resolver.scopes = resolver.scopes[:len(resolver.scopes)-1]
	scope["this"] = true
	resolver.scopes = append(resolver.scopes, scope)

	// resolver class methods
	for _, method := range stmt.Methods {
		ft := callableenum.METHOD
		if method.Name.Lexeme == "init" {
			ft = callableenum.INITIALIZER
		}
		resolver.resolveFunction(method, ft)
	}
	resolver.endScope()
	// discard super scope
	if stmt.Superclass.Token.Lexeme != "" {
		resolver.endScope()
	}
	return nil
}

func (resolver *Resolver) VisitBinary(expr expressions.Binary) (interface{}, error) {
	resolver.resolveExpr(expr.Left)
	resolver.resolveExpr(expr.Right)
	return nil, nil
}

func (resolver *Resolver) VisitCall(expr expressions.Call) (interface{}, error) {
	resolver.resolveExpr(expr.Callee)
	for _, arg := range expr.Args {
		resolver.resolveExpr(arg)
	}
	return nil, nil
}

func (resolver *Resolver) VisitGrouping(expr expressions.Grouping) (interface{}, error) {
	resolver.resolveExpr(expr.Expr)
	return nil, nil
}

func (resolver *Resolver) VisitLogical(expr expressions.Logical) (interface{}, error) {
	resolver.resolveExpr(expr.Left)
	resolver.resolveExpr(expr.Right)
	return nil, nil
}
func (resolver *Resolver) VisitUnary(expr expressions.Unary) (interface{}, error) {
	resolver.resolveExpr(expr.Right)
	return nil, nil
}
func (resolver *Resolver) VisitLiteral(expr expressions.Literal) (interface{}, error) {
	return nil, nil
}

// Since properties are looked up dynamically,
// they don’t get resolved. During resolution, it recurse only into the expression to the left of the dot.
// The actual property access happens in the interpreter.
func (resolver *Resolver) VisitPropertyAccess(expr expressions.PropertyAccess) (interface{}, error) {
	resolver.resolveExpr(expr.Obj)
	return nil, nil
}

func (resolver *Resolver) VisitPropertyAssignment(expr expressions.PropertyAssignment) (interface{}, error) {
	resolver.resolveExpr(expr.Value)
	resolver.resolveExpr(expr.Obj)
	return nil, nil
}

// initialize the scope
func (resolver *Resolver) beginScope() {
	resolver.scopes = append(resolver.scopes, scope{})
}

func (resolver *Resolver) endScope() {
	resolver.scopes = resolver.scopes[:len(resolver.scopes)-1]
}
func (resolver *Resolver) VisitSuper(expr expressions.Super) (interface{}, error) {
	if resolver.curcls == classenum.NONE {
		reporting.ReportError(expr.Keyword.Line, "Can't use 'super' outside of a class")
	} else if resolver.curcls == classenum.CLASS {
		reporting.ReportError(expr.Keyword.Line, "Can't use 'super' with no superclass")
	}
	resolver.resolveLocalVar(expr, expr.Keyword.Lexeme)
	return nil, nil
}

// adds the variable to the innermost scope so that it shadows
// any outer one and so that it knows the variable exists.
// marks it as 'not yet ready' by binding its name to false.
// The value represents whether or not it have finished resolving that variable’s initializer.
func (resolver *Resolver) declare(name expressions.Token) {
	if len(resolver.scopes) == 0 {
		return
	}

	scope := resolver.scopes[len(resolver.scopes)-1]
	resolver.scopes = resolver.scopes[:len(resolver.scopes)-1]
	_, containsVar := scope[name.Lexeme]
	if containsVar {
		reporting.ReportError(name.Line, fmt.Sprintf("Duplicate decleration of variable '%s' in local scope", name.Lexeme))
	}
	scope[name.Lexeme] = false
	resolver.scopes = append(resolver.scopes, scope)
}

// resolve the variable in that same scope where the variable exists but is unavailable
func (resolver *Resolver) define(name expressions.Token) {
	if len(resolver.scopes) == 0 {
		return
	}
	scope := resolver.scopes[len(resolver.scopes)-1]
	resolver.scopes = resolver.scopes[:len(resolver.scopes)-1]
	scope[name.Lexeme] = true
	resolver.scopes = append(resolver.scopes, scope)
}

func (resolver *Resolver) resolveLocalVar(expr expressions.Experssion, name string) {
	// starts at the innermost scope and work outwards,
	// looking in each map for a matching name.
	// If it find the variable, it resolve it, passing in the number of scopes
	// between the current innermost scope and the scope where the variable was found
	// if it never finds the variable, it will leave it unresolved and assume it’s global.
	for i := len(resolver.scopes) - 1; i >= 0; i-- {
		scope := resolver.scopes[i]
		flag := scope[name]
		if flag {
			resolver.inter.Resolve(expr, len(resolver.scopes)-1-i)
			return
		}
	}
}

// creates a new scope for the body and then binds variables for each of the parameter
// then it resolves the function body in that scope
func (resolver *Resolver) resolveFunction(function statements.FunctionStatement, ft int) {
	// set the current scope type and save the enclosing one
	encloseft := resolver.curft
	resolver.curft = ft
	resolver.beginScope()
	for _, arg := range function.Args {
		resolver.declare(arg)
		resolver.define(arg)
	}
	resolver.Resolve(function.Body)
	resolver.endScope()
	// restore the enclosing scope type
	resolver.curft = encloseft
}
