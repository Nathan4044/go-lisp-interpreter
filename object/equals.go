// A collection of functions for calculating equality between objects.
package object

// Compare list of objects to ensure all have
// the same value as the initially given number.
func numsEqual(first float64, rest ...Object) *BooleanObject {
	for _, arg := range rest {
		var num float64

		switch obj := arg.(type) {
		case *Number:
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
func stringsEqual(first *String, rest ...Object) *BooleanObject {
	for _, arg := range rest {
		str, ok := arg.(*String)

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
func boolEqual(first *BooleanObject, rest ...Object) *BooleanObject {
	for _, arg := range rest {
		boolean, ok := arg.(*BooleanObject)

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
func lambdasEqual(first *LambdaObject, rest ...Object) *BooleanObject {
	for _, arg := range rest {
		lambda, ok := arg.(*LambdaObject)

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
func functionsEqual(first *FunctionObject, rest ...Object) *BooleanObject {
	for _, arg := range rest {
		function, ok := arg.(*FunctionObject)

		if !ok {
			return FALSE
		}

		if function.Name != first.Name {
			return FALSE
		}
	}

	return TRUE
}
