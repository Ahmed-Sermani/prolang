package interpreter

import (
	"fmt"

	"github.com/Ahmed-Sermani/prolang/interpreter/environment"
	"github.com/Ahmed-Sermani/prolang/parser/expressions"
	"github.com/Ahmed-Sermani/prolang/parser/statements"
)

type Callable interface {
	Call(*Interpreter, []interface{}) (interface{}, error)
	ArgsNum() int
}

type FunctionCallable struct {
	Declaration statements.FunctionStatement
	Closure     *environment.Environment
	// flags the current function is an initializer
	IsInit bool
}

// implementing the callable interface
func (f *FunctionCallable) Call(inter *Interpreter, args []interface{}) (interface{}, error) {
	// define function environment
	// to handle recursion the environment for function created on
	// the call not on the function deleration
	environment := environment.New(f.Closure)

	// add args to the function environment
	for i, param := range f.Declaration.Args {
		environment.Define(param.Lexeme, args[i])
	}
	// execute function body
	err := inter.executeBlock(f.Declaration.Body, environment)

	// handling the unwind of return statement
	returnValue, isReturn := err.(ErrorHandleReturn)
	// handle empty return from initializer default to 'this'
	if isReturn && f.IsInit {
		return f.Closure.GetAt(0, expressions.Token{Lexeme: "this"})
	}
	if isReturn {
		return returnValue.value, nil
	}
	// if the function initializer then flag it.
	// initializer always return 'this' if it called directly
	// can't rely on 'init' to determind if it's a initializer because maybe some function named with the same name
	if f.IsInit {
		return f.Closure.GetAt(0, expressions.Token{Lexeme: "this"})
	}
	// default return value is nil
	return nil, err
}

func (f *FunctionCallable) ArgsNum() int {
	return len(f.Declaration.Args)
}

func (f *FunctionCallable) bind(i *Instance) *FunctionCallable {
	environment := environment.New(f.Closure)
	environment.Define("this", i)
	return &FunctionCallable{Declaration: f.Declaration, Closure: environment, IsInit: f.IsInit}
}

func (f *FunctionCallable) String() string {
	return "<func " + f.Declaration.Name.Lexeme + ">"
}

type ClassCallable struct {
	name       string
	methods    map[string]*FunctionCallable
	superclass *ClassCallable
}

type Instance struct {
	class  *ClassCallable
	fields map[string]interface{}
}

func (c *ClassCallable) Call(inter *Interpreter, args []interface{}) (interface{}, error) {
	instance := &Instance{class: c, fields: map[string]interface{}{}}
	// checking the initializer method and calling it if exists
	init := instance.class.lookForMethod("init")
	if init != nil {
		// bind 'this' then call the initializer method
		init.bind(instance).Call(inter, args)
	}
	return instance, nil
}

func (c *ClassCallable) lookForMethod(name string) *FunctionCallable {
	method, ok := c.methods[name]
	if ok {
		return method
	}
	// rescursive lookup in the superclass
	if c.superclass != nil {
		return c.superclass.lookForMethod(name)
	}
	return nil
}

// the number of arguments of class is the same as the number of arguments on the initializer
func (c *ClassCallable) ArgsNum() int {
	init := c.lookForMethod("init")
	// default to zero there's no initializer
	if init == nil {
		return 0
	}
	return init.ArgsNum()
}

func (c *ClassCallable) String() string {
	return "<class " + c.name + ">"
}

// lookup a property on an instance
func (i *Instance) Get(name expressions.Token) (interface{}, error) {
	property, exists := i.fields[name.Lexeme]
	if exists {
		return property, nil
	}

	method := i.class.lookForMethod(name.Lexeme)
	// return the method and bind the current instance to 'this'
	if method != nil {
		return method.bind(i), nil
	}

	return nil, &UndefinedProperty{
		InterpretationError: InterpretationError{
			msg: fmt.Sprintf("Undefined property '%s' on object of '%s'", name.Lexeme, i.class.name),
		},
	}
}

// set a field on an instance
// creation of new field freely is allowed
func (i *Instance) Set(name expressions.Token, value interface{}) {
	i.fields[name.Lexeme] = value
}

func (i *Instance) String() string {
	return "<instance of " + i.class.name + ">"
}
