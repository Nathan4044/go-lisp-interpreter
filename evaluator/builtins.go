// Namespace for containing the builtin functions map.
package evaluator

import (
	"lisp/object"
)

// A map of all the built in functions in the interpreter
var builtins = map[string]*object.FunctionObject{
	"+":     object.GetBuiltinByName("+"),
	"*":     object.GetBuiltinByName("*"),
	"-":     object.GetBuiltinByName("-"),
	"/":     object.GetBuiltinByName("/"),
	"rem":   object.GetBuiltinByName("rem"),
	"=":     object.GetBuiltinByName("="),
	"<":     object.GetBuiltinByName("<"),
	">":     object.GetBuiltinByName(">"),
	"not":   object.GetBuiltinByName("not"),
	"and":   object.GetBuiltinByName("and"),
	"or":    object.GetBuiltinByName("or"),
	"list":  object.GetBuiltinByName("list"),
	"dict":  object.GetBuiltinByName("dict"),
	"first": object.GetBuiltinByName("first"),
	"rest":  object.GetBuiltinByName("rest"),
	"last":  object.GetBuiltinByName("last"),
	"len":   object.GetBuiltinByName("len"),
	"push":  object.GetBuiltinByName("push"),
	"str":   object.GetBuiltinByName("str"),
	"print": object.GetBuiltinByName("print"),
	"get":   object.GetBuiltinByName("get"),
	"set":   object.GetBuiltinByName("set"),
}

func evalTruthy(obj object.Object) bool {
	if obj == NULL || obj == FALSE {
		return false
	}

	return true
}

func isInt(num float64) bool {
	return num == float64(int64(num))
}
