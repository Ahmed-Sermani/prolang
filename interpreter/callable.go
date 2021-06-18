package interpreter

import (
	"github.com/Ahmed-Sermani/prolang/interpreter/environment"
	"github.com/Ahmed-Sermani/prolang/parser/statements"
)

type Callable interface {
	Call(*Interpreter, []interface{}) (interface{}, error)
	ArgsNum() int
}

type FunctionCallable struct {
	Declaration statements.FunctionStatement
}

// implementing the callable interface

func (f *FunctionCallable) Call(inter *Interpreter, args []interface{}) (interface{}, error) {
	// define function environment
	// to handle recursion the environment for function created on
	// the call not on the function deleration
	environment := environment.New(inter.global)

	// add args to the function environment
	for i, param := range f.Declaration.Args {
		environment.Define(param.Lexeme, args[i])
	}
	// execute function body
	err := inter.executeBlock(f.Declaration.Body, environment)
	return nil, err
}

func (f *FunctionCallable) ArgsNum() int {
	return len(f.Declaration.Args)
}

func (f *FunctionCallable) String() string {
	return "<func " + f.Declaration.Name.Lexeme + ">"
}
