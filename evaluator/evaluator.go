package evaluator

import (
	"fmt"
	"lisp/ast"
	"lisp/object"
)

var (
	TRUE  = &object.BooleanObject{Value: true}
	FALSE = &object.BooleanObject{Value: false}
	NULL  = &object.Null{}
)

func Evaluate(e ast.Expression, env *object.Environment) object.Object {
	switch e := e.(type) {
    case *ast.Program:
	    var result object.Object

	    for _, expression := range e.Expressions {
	    	result = Evaluate(expression, env)
	    }

	    return result
	case *ast.IntegerLiteral:
		return &object.Integer{Value: e.Value}
	case *ast.FloatLiteral:
		return &object.Float{Value: e.Value}
	case *ast.StringLiteral:
		return &object.String{Value: e.Value}
	case *ast.Identifier:
		return evalIdentifier(e, env)
	case *ast.SExpression:
		return evaluateSExpression(e, env)
	default:
		return NULL
	}
}

func evaluateSExpression(e *ast.SExpression, env *object.Environment) object.Object {
	if e.Fn == nil {
		return &object.List{}
	}

	fnExpression := Evaluate(e.Fn, env)

	switch fnExpression := fnExpression.(type) {
	case *object.FunctionObject:
		return fnExpression.Fn(env, e.Args...)
	case *object.LambdaObject:
		return evalLambda(e.Fn.String(), fnExpression, env, e.Args...)
	default:
		err := fmt.Sprintf("%s is not a function", fnExpression.Inspect())
		return &object.ErrorObject{
			Error: err,
		}
	}
}

func evalIdentifier(i *ast.Identifier, env *object.Environment) object.Object {
	if i.String() == "true" {
		return TRUE
	}

	if i.String() == "false" {
		return FALSE
	}

	if i.String() == "null" {
		return NULL
	}

	builtins := makeBuitins()
	fn, ok := builtins[i.String()]

	if ok {
		return &object.FunctionObject{Fn: fn, Name: i.String()}
	}

	return env.Get(i.String())
}

func evalLambda(lambdaName string, lambda *object.LambdaObject, env *object.Environment, args ...ast.Expression) object.Object {
	if len(lambda.Args) != len(args) {
		err := fmt.Sprintf("incorrect number of args for %s: expected=%d got=%d",
			lambdaName, len(lambda.Args), len(args))
		return &object.ErrorObject{Error: err}
	}

	lambdaEnv := object.NewEnvironment(lambda.Env)

	for i, arg := range args {
		obj := Evaluate(arg, env)

		if obj.Type() == object.ERROR_OBJ {
			return obj
		}

		lambdaEnv.Set(
			lambda.Args[i],
			Evaluate(arg, env),
		)
	}

	lastIndex := len(lambda.Body) - 1

	for _, exp := range lambda.Body[:lastIndex] {
		obj := Evaluate(exp, lambdaEnv)

		if obj.Type() == object.ERROR_OBJ {
			return obj
		}
	}

	return Evaluate(lambda.Body[lastIndex], lambdaEnv)
}
