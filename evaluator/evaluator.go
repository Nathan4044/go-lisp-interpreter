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

func Evaluate(program ast.Program, env *object.Environment) object.Object {
	var result object.Object

	for _, expression := range program.Expressions {
		result = eval(expression, env)
	}

	return result
}

func eval(e ast.Expression, env *object.Environment) object.Object {
	switch e := e.(type) {
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

	fnExpression := eval(e.Fn, env)

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
		obj := eval(arg, env)

		err, ok := obj.(*object.ErrorObject)

		if ok {
			return err
		}

		lambdaEnv.Set(
			lambda.Args[i],
			eval(arg, env),
		)
	}

	lastIndex := len(lambda.Body) - 1

	for _, exp := range lambda.Body[:lastIndex] {
		obj := eval(exp, lambdaEnv)

		err, ok := obj.(*object.ErrorObject)

		if ok {
			return err
		}
	}

	return eval(lambda.Body[lastIndex], lambdaEnv)
}
