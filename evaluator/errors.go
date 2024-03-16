// Collection of functions for creating common error objects.
package evaluator

import (
	"fmt"
	"lisp/object"
)

func badTypeError(fn string, obj object.Object) *object.ErrorObject {
	err := fmt.Sprintf("attempted to call %s with unsupported type %s (%s)",
		fn, obj.Type(), obj.Inspect())

	return &object.ErrorObject{Error: err}
}

func badKeyError(obj object.Object) *object.ErrorObject {
	err := fmt.Sprintf("attempted to use unsupported type as dict key %s (%s)",
		obj.Type(), obj.Inspect())

	return &object.ErrorObject{Error: err}
}

func noArgsError(fn string) *object.ErrorObject {
	err := fmt.Sprintf("attempted to call %s with no arguments", fn)
	return &object.ErrorObject{Error: err}
}

func wrongNumOfArgsError(fn string, expected string, got int) *object.ErrorObject {
	err := fmt.Sprintf("attempted to call %s with incorrect number of arguments: expected %s, got=%d",
		fn, expected, got)
	return &object.ErrorObject{Error: err}
}
