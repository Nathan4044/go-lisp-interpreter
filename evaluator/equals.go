// A collection of functions for calculating equality between objects.
package evaluator

import (
	"lisp/object"
)

// Compare list of objects to ensure all have
// the same value as the initially given number.
func numsEqual(first float64, rest ...object.Object) *object.BooleanObject {
	for _, arg := range rest {
		var num float64

		switch obj := arg.(type) {
		case *object.Integer:
			num = float64(obj.Value)
		case *object.Float:
			num = obj.Value
		default:
			return FALSE
		}

		if num != first {
			return FALSE
		}
	}

	return TRUE
}

// Compare list of objects to ensure all have
// the same value as the initially given string.
func stringsEqual(first *object.String, rest ...object.Object) *object.BooleanObject {
	for _, arg := range rest {
		str, ok := arg.(*object.String)

		if !ok {
			return FALSE
		}

		if str.Value != first.Value {
			return FALSE
		}
	}

	return TRUE
}

// Compare list of objects to ensure all have
// the same value as the initially given bool.
func boolEqual(first *object.BooleanObject, rest ...object.Object) *object.BooleanObject {
	for _, arg := range rest {
		boolean, ok := arg.(*object.BooleanObject)

		if !ok {
			return FALSE
		}

		if boolean.Value != first.Value {
			return FALSE
		}
	}

	return TRUE
}

// Compare list of objects to ensure all have
// the same value as the initially given lambda.
func lambdasEqual(first *object.LambdaObject, rest ...object.Object) *object.BooleanObject {
	for _, arg := range rest {
		lambda, ok := arg.(*object.LambdaObject)

		if !ok {
			return FALSE
		}

		if lambda != first {
			return FALSE
		}
	}

	return TRUE
}

// Compare list of objects to ensure all have
// the same value as the initially given function.
func functionsEqual(first *object.FunctionObject, rest ...object.Object) *object.BooleanObject {
	for _, arg := range rest {
		function, ok := arg.(*object.FunctionObject)

		if !ok {
			return FALSE
		}

		if function.Name != first.Name {
			return FALSE
		}
	}

	return TRUE
}
