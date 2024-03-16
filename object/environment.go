// Definition of the Environment type.
package object

import "fmt"

// Environment is the data structure which holds values
// that are used during program evaluation.
type Environment struct {
	outer  *Environment      // The enclosing Environment, where the current Environment was defined.
	values map[string]Object // A map holding each of the objects defined in the Environment.
}

// Return the object from the Environment that is associated
// with the provided identifier.
//
// If the identifier is not defined in the Environment, it will
// query the enclosing Environment.
//
// If there is no enclosing Environment and the identifier isn't
// found, an Error Object is returned.
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

// Store the provided Object in the Environment, with its key
// being the provided identifier string.
func (e *Environment) Set(ident string, obj Object) {
	e.values[ident] = obj
}

// Create a new Environment object and return its address.
//
// If an outer Environment is provided, use it to enclose the
// new Environment.
func NewEnvironment(outer *Environment) *Environment {
	e := Environment{
		values: make(map[string]Object),
	}

	if outer != nil {
		e.outer = outer
	}

	return &e
}
