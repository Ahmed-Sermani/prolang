package resolver

import (
	"fmt"

	"github.com/Ahmed-Sermani/prolang/interpreter"
	"github.com/Ahmed-Sermani/prolang/parser/expressions"
	"github.com/Ahmed-Sermani/prolang/parser/statements"
	"github.com/Ahmed-Sermani/prolang/reporting"
)

type scope map[string]bool
type stack []scope

func (s *stack) push(sc scope) {
	*s = append(*s, sc)
}

func (s *stack) pop() (scope, bool) {
	if s.isEmpty() {
		return scope{}, false
	} else {
		index := len(*s) - 1
		element := (*s)[index]
		*s = (*s)[:index]
		return element, true
	}
}

func (s *stack) peak() scope {
	return (*s)[len(*s)-1]
}

func (s *stack) isEmpty() bool {
	return len(*s) == 0
}

// static analysis.
// resolve variables to solve the problem of dynamic changing environments
// that apperas with closures.
// after the parser produces the syntax tree,
// but before the interpreter starts executing, it’ll do a single walk over the tree to resolve all of the variables it contains
// aka persisting static scope.
// implements statements and expressions visitor interface.
type Resolver struct {
	inter  interpreter.Interpreter
	scopes stack
}

func New(inter interpreter.Interpreter) *Resolver {
	return &Resolver{
		inter: inter,
		// track the stack of scopes currently in scope.
		// Each element in the stack is a Map representing a single block scope.
		// Keys are variable names. The values are Booleans.
		// the boolean value to track the definition.
		// false if the variable declared but not yet defined, true if both
		scopes: stack{},
	}
}

// resolving blocks
func (resolver *Resolver) VisitBlockStmt(stmt statements.BlockStatement) error {
	resolver.beginScope()
	resolver.resolve(stmt.Statements)
	resolver.endScope()
	return nil
}

// walks the statements and resolves each one.
func (resolver *Resolver) resolve(stmts []statements.Statement) {
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
	resolver.declare(stmt.Token.Lexeme)
	if stmt.Initializer != nil {
		resolver.resolveExpr(stmt.Initializer)
	}
	resolver.define(stmt.Token.Lexeme)
	return nil
}

func (resolver *Resolver) VisitVairable(expr expressions.Variable) (interface{}, error) {
	scope, notEmpty := resolver.scopes.pop()

	state, exists := scope[expr.Token.Lexeme]
	// if the scopes is not empty and the variable is not defined
	if notEmpty && exists && !state {
		// Make it an error to reference a variable in its initializer
		// e.g. let a = 8;
		// let a = a;
		reporting.ReportError(expr.Token.Line, fmt.Sprintf("Can't read local variable %s in its own initializer", expr.Token.Lexeme))
		return nil, nil
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

func (resolver *Resolver) VisitFunctionStmt(stmt statements.FunctionStatement) error {
	// define the declare the function before resolving to let the function refer to itself
	resolver.declare(stmt.Name.Lexeme)
	resolver.define(stmt.Name.Lexeme)
	resolver.resolveFunction(stmt.Body)
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

// initialize the scope
func (resolver *Resolver) beginScope() {
	resolver.scopes.push(scope{})
}

func (resolver *Resolver) endScope() {
	resolver.scopes.pop()
}

// adds the variable to the innermost scope so that it shadows
// any outer one and so that it knows the variable exists.
// marks it as 'not yet ready' by binding its name to false.
// The value represents whether or not it have finished resolving that variable’s initializer.
func (resolver *Resolver) declare(name string) {
	scope, ok := resolver.scopes.pop()
	if !ok {
		return
	}
	scope[name] = false
	resolver.scopes.push(scope)
}

// resolve the variable in that same scope where the variable exists but is unavailable
func (resolver *Resolver) define(name string) {
	scope, ok := resolver.scopes.pop()
	if !ok {
		return
	}
	scope[name] = true
	resolver.scopes.push(scope)
}

func (resolver *Resolver) resolveLocalVar(expr expressions.Experssion, name string) {
	// starts at the innermost scope and work outwards,
	// looking in each map for a matching name.
	// If it find the variable, it resolve it, passing in the number of scopes
	// between the current innermost scope and the scope where the variable was found
	// if it never finds the variable, it will leave it unresolved and assume it’s global.
	for i := len(resolver.scopes) - 1; i >= 0; i-- {
		scope := resolver.scopes[i]
		_, exists := scope[name]
		if exists {
			resolver.inter.Resolve(expr, len(resolver.scopes)-1-i)
		}
	}
}

// creates a new scope for the body and then binds variables for each of the parameter
// then it resolves the function body in that scope
func (resolver *Resolver) resolveFunction(function statements.FunctionStatement) {
	resolver.beginScope()
	for _, arg := range function.Args {
		resolver.declare(arg.Lexeme)
		resolver.define(arg.Lexeme)
	}
	resolver.resolve(function.Body)
	resolver.endScope()
}
