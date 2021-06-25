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

func New(enclosing *Environment) *Environment {
	return &Environment{
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

// walks a fixed number of predecessors up the parent chain and
// returns the value of the variable in that environment’s map.
// It doesn’t check if the variable exists because the resolver already found it.
func (env *Environment) GetAt(level int, t expressions.Token) (interface{}, error) {
	predecEnv := env.predecessors(level)
	return predecEnv.values[t.Lexeme], nil

}

func (env *Environment) AssginAt(level int, t expressions.Token, value interface{}) error {
	predecEnv := env.predecessors(level)
	predecEnv.values[t.Lexeme] = value
	return nil
}

func (env *Environment) predecessors(level int) *Environment {
	predecEnv := env
	for i := 0; i < level; i++ {
		predecEnv = predecEnv.enclosing
	}

	return predecEnv
}
