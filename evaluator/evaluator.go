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

// Recursively evaluate a given expression and return a final value.
func Evaluate(e ast.Expression, env *object.Environment) object.Object {
	switch e := e.(type) {
	case *ast.Program:
		var result object.Object

		for _, expression := range e.Expressions {
			result = Evaluate(expression, env)

			if result.Type() == object.ERROR_OBJ {
				return result
			}
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

	// this switch covers evaluating special functions that break the standard
	// function construction
	switch e.Fn.String() {
	case "if":
		return evaluateIfExpression(e, env)
	case "def":
		return evaluateDefExpression(e, env)
	case "lambda":
		return evaluateLambdaExpression(e, env)
	}

	fnExpression := Evaluate(e.Fn, env)

	args := []object.Object{}
	for _, arg := range e.Args {
		obj := Evaluate(arg, env)

		if obj.Type() == object.ERROR_OBJ {
			return obj
		}

		args = append(args, obj)
	}

	switch fnExpression := fnExpression.(type) {
	case *object.FunctionObject:
		return fnExpression.Fn(env, args...)
	case *object.LambdaObject:
		return evalLambda(e.Fn.String(), fnExpression, env, args...)
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

	fn, ok := builtins[i.String()]

	if ok {
		return &object.FunctionObject{Fn: fn, Name: i.String()}
	}

	return env.Get(i.String())
}

func evalLambda(lambdaName string, lambda *object.LambdaObject, env *object.Environment, args ...object.Object) object.Object {
	if len(lambda.Args) != len(args) {
		err := fmt.Sprintf("incorrect number of args for %s: expected=%d got=%d",
			lambdaName, len(lambda.Args), len(args))
		return &object.ErrorObject{Error: err}
	}

	lambdaEnv := object.NewEnvironment(lambda.Env)

	for i, arg := range args {
		lambdaEnv.Set(
			lambda.Args[i],
			arg,
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

func evaluateIfExpression(e *ast.SExpression, env *object.Environment) object.Object {
	if len(e.Args) < 2 || len(e.Args) > 3 {
		return wrongNumOfArgsError("if", "2 or 3", len(e.Args))
	}

	obj := Evaluate(e.Args[0], env)

	if obj.Type() == object.ERROR_OBJ {
		return obj
	}

	condition := evalTruthy(obj)

	if condition {
		return Evaluate(e.Args[1], env)
	}

	if len(e.Args) == 3 {
		return Evaluate(e.Args[2], env)
	}

	return NULL
}

func evaluateDefExpression(e *ast.SExpression, env *object.Environment) object.Object {
	if len(e.Args) != 2 {
		return wrongNumOfArgsError("def", "2", len(e.Args))
	}

	ident, ok := e.Args[0].(*ast.Identifier)

	if !ok {
		err := fmt.Sprintf("cannot assign to non-identifier %s", e.Args[0].String())
		return &object.ErrorObject{Error: err}
	}

	val := Evaluate(e.Args[1], env)

	if val.Type() != object.ERROR_OBJ {
		env.Set(ident.String(), val)
	}

	return val
}

func evaluateLambdaExpression(e *ast.SExpression, env *object.Environment) object.Object {
	args := e.Args

	if len(args) < 2 {
		return wrongNumOfArgsError("lambda", "at least 2", len(args))
	}

	argsList, ok := args[0].(*ast.SExpression)

	if !ok {
		err := fmt.Sprintf("lambda requires list of args, got %s", args[0].String())
		return &object.ErrorObject{Error: err}
	}

	lambdaArgs := []string{}

	if argsList.Fn != nil {
		arg, ok := argsList.Fn.(*ast.Identifier)

		if !ok {
			err := fmt.Sprintf("lambda function must be identifier, got %s", argsList.Fn.String())
			return &object.ErrorObject{Error: err}
		}

		lambdaArgs = append(lambdaArgs, arg.String())
	}

	for _, lambdaArg := range argsList.Args {
		arg, ok := lambdaArg.(*ast.Identifier)

		if !ok {
			err := fmt.Sprintf("lambda args must be identifiers, got %s", arg)
			return &object.ErrorObject{Error: err}
		}

		lambdaArgs = append(lambdaArgs, arg.String())
	}

	return &object.LambdaObject{
		Args: lambdaArgs,
		Env:  env,
		Body: args[1:],
	}
}
