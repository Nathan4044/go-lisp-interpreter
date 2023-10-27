package object

import "fmt"

type Environment struct {
	outer  *Environment
	values map[string]Object
}

func (e *Environment) Get(ident string) Object {
	result, ok := e.values[ident]

	if ok {
		return result
	}

	if e.outer != nil {
		return e.outer.Get(ident)
	}

	err := fmt.Sprintf("No such item: %s", ident)
	return &ErrorObject{Error: err}
}

func (e *Environment) Set(ident string, obj Object) {
	e.values[ident] = obj
}

func NewEnvironment(outer *Environment) *Environment {
	e := Environment{
		values: make(map[string]Object),
	}

	if outer != nil {
		e.outer = outer
	}

	return &e
}
