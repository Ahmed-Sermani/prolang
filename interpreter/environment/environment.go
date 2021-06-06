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
}

func (env *Environment) Define(name string, value interface{}) {
	if env.values == nil {
		env.values = map[string]interface{}{}
	}
	env.values[name] = value
}

func (env *Environment) Get(t expressions.Token) (interface{}, error) {
	v, ok := env.values[t.Lexeme]
	if !ok {
		err := ErrorUndefinedVairable{msg: fmt.Sprintf("Undefined Variable '%s'.", t.Lexeme)}
		reporting.ReportRuntimeError(err)
		return nil, err
	}
	return v, nil
}
