package evaluator

import (
	"lisp/ast"
	"lisp/object"
)

func numsEqual(first float64, env *object.Environment, rest ...ast.Expression) *object.BooleanObject {
	for _, arg := range rest {
		var num float64
		obj := Evaluate(arg, env)

		switch obj := obj.(type) {
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

func stringsEqual(first *object.String, env *object.Environment, rest ...ast.Expression) *object.BooleanObject {
	for _, arg := range rest {
		obj := Evaluate(arg, env)
		str, ok := obj.(*object.String)

		if !ok {
			return FALSE
		}

		if str.Value != first.Value {
			return FALSE
		}
	}

	return TRUE
}

func boolEqual(first *object.BooleanObject, env *object.Environment, rest ...ast.Expression) *object.BooleanObject {
	for _, arg := range rest {
		obj := Evaluate(arg, env)
		boolean, ok := obj.(*object.BooleanObject)

		if !ok {
			return FALSE
		}

		if boolean.Value != first.Value {
			return FALSE
		}
	}

	return TRUE
}

func lambdasEqual(first *object.LambdaObject, env *object.Environment, rest ...ast.Expression) *object.BooleanObject {
	for _, arg := range rest {
		obj := Evaluate(arg, env)
		lambda, ok := obj.(*object.LambdaObject)

		if !ok {
			return FALSE
		}

		if lambda != first {
			return FALSE
		}
	}

	return TRUE
}

func functionsEqual(first *object.FunctionObject, env *object.Environment, rest ...ast.Expression) *object.BooleanObject {
	for _, arg := range rest {
		obj := Evaluate(arg, env)
		function, ok := obj.(*object.FunctionObject)

		if !ok {
			return FALSE
		}

		if function.Name != first.Name {
			return FALSE
		}
	}

	return TRUE
}
