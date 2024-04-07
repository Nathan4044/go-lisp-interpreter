// Collection of functions for creating common error objects.
package object

import (
	"fmt"
)

func BadTypeError(fn string, obj Object) *ErrorObject {
	err := fmt.Sprintf("attempted to call %s with unsupported type %s (%s)",
		fn, obj.Type(), obj.Inspect())

	return &ErrorObject{Error: err}
}

func BadKeyError(obj Object) *ErrorObject {
	err := fmt.Sprintf("attempted to use unsupported type as dict key %s (%s)",
		obj.Type(), obj.Inspect())

	return &ErrorObject{Error: err}
}

func NoArgsError(fn string) *ErrorObject {
	err := fmt.Sprintf("attempted to call %s with no arguments", fn)
	return &ErrorObject{Error: err}
}

func WrongNumOfArgsError(fn string, expected string, got int) *ErrorObject {
	err := fmt.Sprintf("attempted to call %s with incorrect number of arguments: expected %s, got=%d",
		fn, expected, got)
	return &ErrorObject{Error: err}
}
