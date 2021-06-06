package environment

import (
	"fmt"

	"github.com/Ahmed-Sermani/prolang/parser/expressions"
	"github.com/Ahmed-Sermani/prolang/reporting"
)

// variable manager

type ErrorUndefinedVairable struct {
	msg string
}

func (e ErrorUndefinedVairable) Error() string {
	if e.msg == "" {
		return "Undefined Vairable"
	}
	return e.msg
}

type Environment struct {
	values map[string]interface{}
	// scoping support
	// refer to the outer scope variables
	enclosing *Environment
}

func New(enclosing *Environment) Environment {
	return Environment{
		values:    map[string]interface{}{},
		enclosing: enclosing,
	}
}

func (env *Environment) Define(name string, value interface{}) {
	env.values[name] = value
}

func (env *Environment) Get(t expressions.Token) (interface{}, error) {
	v, ok := env.values[t.Lexeme]
	if ok {
		return v, nil
	}

	// recursive lookup the variable into the outer scopes
	if env.enclosing != nil {
		return env.enclosing.Get(t)
	}
	err := ErrorUndefinedVairable{msg: fmt.Sprintf("Undefined Variable '%s'.", t.Lexeme)}
	reporting.ReportRuntimeError(err)
	return nil, err

}

func (env *Environment) Assgin(t expressions.Token, value interface{}) error {
	if _, ok := env.values[t.Lexeme]; ok {
		env.values[t.Lexeme] = value
		return nil
	}

	// recursive lookup into the outer scopes to the variable to assgin
	if env.enclosing != nil {
		return env.enclosing.Assgin(t, value)
	}
	err := ErrorUndefinedVairable{msg: fmt.Sprintf("Undefined Variable '%s'.", t.Lexeme)}
	reporting.ReportRuntimeError(err)
	return err
}
