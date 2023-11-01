package evaluator

import (
	"bytes"
	"fmt"
	"lisp/ast"
	"lisp/object"
	"strings"
)

func makeBuitins() map[string]object.Function {
	return map[string]object.Function{
		"+": func(env *object.Environment, args ...ast.Expression) object.Object {
			var result float64 = 0

			for _, arg := range args {
				obj := Evaluate(arg, env)

				switch obj := obj.(type) {
				case *object.Integer:
					result += float64(obj.Value)
				case *object.Float:
					result += obj.Value
				case *object.ErrorObject:
					return obj
				default:
					return badTypeError("+", obj)
				}
			}

			if isInt(result) {
				return &object.Integer{Value: int64(result)}
			}
			return &object.Float{Value: result}
		},
		"*": func(env *object.Environment, args ...ast.Expression) object.Object {
			var result float64 = 1

			for _, arg := range args {
				obj := Evaluate(arg, env)

				switch obj := obj.(type) {
				case *object.Integer:
					result *= float64(obj.Value)
				case *object.Float:
					result *= obj.Value
				case *object.ErrorObject:
					return obj
				default:
					return badTypeError("*", obj)
				}
			}

			if isInt(result) {
				return &object.Integer{Value: int64(result)}
			}
			return &object.Float{Value: result}
		},
		"-": func(env *object.Environment, args ...ast.Expression) object.Object {
			if len(args) == 0 {
				return noArgsError("-")
			}

			var nums []float64

			for _, arg := range args {
				obj := Evaluate(arg, env)

				switch obj := obj.(type) {
				case *object.Integer:
					nums = append(nums, float64(obj.Value))
				case *object.Float:
					nums = append(nums, obj.Value)
				case *object.ErrorObject:
					return obj
				default:
					return badTypeError("-", obj)
				}
			}

			if len(nums) == 1 {
				if isInt(nums[0]) {
					return &object.Integer{Value: -int64(nums[0])}
				}
				return &object.Float{Value: -nums[0]}
			} else {
				result := nums[0]

				for _, num := range nums[1:] {
					result -= num
				}

				if isInt(result) {
					return &object.Integer{Value: int64(result)}
				}
				return &object.Float{Value: result}
			}
		},
		"/": func(env *object.Environment, args ...ast.Expression) object.Object {
			if len(args) == 0 {
				return noArgsError("/")
			}

			var nums []float64

			for _, arg := range args {
				obj := Evaluate(arg, env)

				switch obj := obj.(type) {
				case *object.Integer:
					nums = append(nums, float64(obj.Value))
				case *object.Float:
					nums = append(nums, obj.Value)
				case *object.ErrorObject:
					return obj
				default:
					return badTypeError("/", obj)
				}
			}

			if len(nums) == 1 {
				return &object.Float{Value: 1 / nums[0]}
			} else {
				result := nums[0]

				for _, num := range nums[1:] {
					if num == 0 {
						return &object.ErrorObject{
							Error: "Attempted to divide by 0",
						}
					}
					result /= num
				}

				if isInt(result) {
					return &object.Integer{Value: int64(result)}
				}
				return &object.Float{Value: result}
			}
		},
		"rem": func(env *object.Environment, args ...ast.Expression) object.Object {
			if len(args) != 2 {
				return wrongNumOfArgsError("rem", "2", len(args))
			}

			var ints [2]int64

			for i, arg := range args {
				obj := Evaluate(arg, env)
				switch obj := obj.(type) {
				case *object.Integer:
					ints[i] = obj.Value
				case *object.ErrorObject:
					return obj
				default:
					return badTypeError("/", obj)
				}

			}

			if ints[1] == 0 {
				return &object.ErrorObject{
					Error: "Attempted rem of 0",
				}
			}

			return &object.Integer{Value: ints[0] % ints[1]}
		},
		"=": func(env *object.Environment, args ...ast.Expression) object.Object {
			if len(args) == 0 {
				return TRUE
			}

			obj := Evaluate(args[0], env)

			switch obj := obj.(type) {
			case *object.Integer:
				return numsEqual(float64(obj.Value), env, args[1:]...)
			case *object.Float:
				return numsEqual(obj.Value, env, args[1:]...)
			case *object.String:
				return stringsEqual(obj, env, args[1:]...)
			case *object.BooleanObject:
				return boolEqual(obj, env, args[1:]...)
			case *object.LambdaObject:
				return lambdasEqual(obj, env, args[1:]...)
			case *object.FunctionObject:
				return functionsEqual(obj, env, args[1:]...)
			default:
				return badTypeError("=", obj)
			}
		},
		"<": func(env *object.Environment, args ...ast.Expression) object.Object {
			if len(args) == 0 {
				return wrongNumOfArgsError("<", "at least 1", 0)
			}

			nums := []float64{}

			for _, arg := range args {
				obj := Evaluate(arg, env)

				switch obj := obj.(type) {
				case *object.Integer:
					nums = append(nums, float64(obj.Value))
				case *object.Float:
					nums = append(nums, obj.Value)
				default:
					return badTypeError("<", obj)
				}
			}

			for i, n := range nums[1:] {
				if n <= nums[i] {
					return FALSE
				}
			}

			return TRUE
		},
		">": func(env *object.Environment, args ...ast.Expression) object.Object {
			if len(args) == 0 {
				return wrongNumOfArgsError(">", "at least 1", 0)
			}

			nums := []float64{}

			for _, arg := range args {
				obj := Evaluate(arg, env)

				switch obj := obj.(type) {
				case *object.Integer:
					nums = append(nums, float64(obj.Value))
				case *object.Float:
					nums = append(nums, obj.Value)
				default:
					return badTypeError(">", obj)
				}
			}

			for i, n := range nums[1:] {
				if n >= nums[i] {
					return FALSE
				}
			}

			return TRUE
		},
		"not": func(env *object.Environment, args ...ast.Expression) object.Object {
			if len(args) != 1 {
				return wrongNumOfArgsError("not", "1", len(args))
			}

			obj := Evaluate(args[0], env)

            if obj.Type() == object.ERROR_OBJ {
                return obj
            }

			if evalTruthy(obj) {
				return FALSE
			}
			return TRUE
		},
		"and": func(env *object.Environment, args ...ast.Expression) object.Object {
			for _, arg := range args {
				obj := Evaluate(arg, env)

                if obj.Type() == object.ERROR_OBJ {
                    return obj
                }

				if !evalTruthy(obj) {
					return FALSE
				}
			}

			return TRUE
		},
		"or": func(env *object.Environment, args ...ast.Expression) object.Object {
			for _, arg := range args {
				obj := Evaluate(arg, env)

                if obj.Type() == object.ERROR_OBJ {
					return obj
				}

				if evalTruthy(obj) {
					return TRUE
				}
			}

			return FALSE
		},
		"list": func(env *object.Environment, args ...ast.Expression) object.Object {
			items := []object.Object{}

			for _, arg := range args {
				items = append(items, Evaluate(arg, env))
			}

			return &object.List{
				Values: items,
			}
		},
		"dict": func(env *object.Environment, args ...ast.Expression) object.Object {
			if len(args)%2 != 0 {
				return wrongNumOfArgsError("dict", "even number", len(args))
			}

			items := map[object.HashKey]object.DictPair{}

			for i := 0; i < len(args)-1; i += 2 {
				obj := Evaluate(args[i], env)
				value := Evaluate(args[i+1], env)

				key, ok := obj.(object.Hashable)

				if !ok {
                    if obj.Type() == object.ERROR_OBJ {
						return obj
					}

					return badKeyError(obj)
				}

				items[key.HashKey()] = object.DictPair{
					Key:   obj,
					Value: value,
				}
			}

			return &object.Dict{
				Values: items,
			}
		},
		"first": func(env *object.Environment, args ...ast.Expression) object.Object {
			if len(args) != 1 {
				return wrongNumOfArgsError("first", "1", len(args))
			}

			obj := Evaluate(args[0], env)

			if obj.Type() != object.LIST_OBJ {
				return badTypeError("first", obj)
			}

			list := obj.(*object.List)
			return list.Values[0]
		},
		"rest": func(env *object.Environment, args ...ast.Expression) object.Object {
			if len(args) != 1 {
				return wrongNumOfArgsError("rest", "1", len(args))
			}

			obj := Evaluate(args[0], env)

			if obj.Type() != object.LIST_OBJ {
				return badTypeError("rest", obj)
			}

			list := obj.(*object.List)
			return &object.List{
				Values: list.Values[1:],
			}
		},
		"len": func(env *object.Environment, args ...ast.Expression) object.Object {
			if len(args) != 1 {
				return wrongNumOfArgsError("len", "1", len(args))
			}

			obj := Evaluate(args[0], env)

			if obj.Type() != object.LIST_OBJ {
				return badTypeError("len", obj)
			}

			list := obj.(*object.List)
			return &object.Integer{Value: int64(len(list.Values))}
		},
		"push": func(env *object.Environment, args ...ast.Expression) object.Object {
			if len(args) != 2 {
				return wrongNumOfArgsError("push", "2", len(args))
			}

			listObj := Evaluate(args[0], env)

			if listObj.Type() != object.LIST_OBJ {
                if listObj.Type() == object.ERROR_OBJ {
                    return listObj
                }

				return &object.ErrorObject{
					Error: fmt.Sprintf("first argument to concat should be list. got=%T(%+v)",
						listObj, listObj),
				}
			}

            obj := Evaluate(args[1], env)

			if obj.Type() == object.ERROR_OBJ {
				return obj
			}

			list := listObj.(*object.List)
			newList := make([]object.Object, len(list.Values))

            copy(newList, list.Values)
			newList = append(newList, obj)

			return &object.List{Values: newList}
		},
		"push!": func(env *object.Environment, args ...ast.Expression) object.Object {
			if len(args) != 2 {
				return wrongNumOfArgsError("push", "2", len(args))
			}

			listObj := Evaluate(args[0], env)

			if listObj.Type() != object.LIST_OBJ {
				return &object.ErrorObject{
					Error: fmt.Sprintf("first argument to concat should be list. got=%T(%+v)",
						listObj, listObj),
				}
			}

            obj := Evaluate(args[1], env)

			if obj.Type() == object.ERROR_OBJ {
				return obj
			}

			list := listObj.(*object.List)
			list.Values = append(list.Values, obj)

			return list
		},
		"pop!": func(env *object.Environment, args ...ast.Expression) object.Object {
			if len(args) != 1 {
				return wrongNumOfArgsError("pop", "1", len(args))
			}

			obj := Evaluate(args[0], env)
			if obj.Type() != object.LIST_OBJ {
				return &object.ErrorObject{
					Error: fmt.Sprintf("argument to pop should be list. got=%T(%+v)",
						obj, obj),
				}
			}

			list := obj.(*object.List)

			result := list.Values[len(list.Values)-1]

			list.Values = list.Values[:len(list.Values)-1]

			return result
		},
		"if": func(env *object.Environment, args ...ast.Expression) object.Object {
			if len(args) < 2 || len(args) > 3 {
				return wrongNumOfArgsError("if", "2 or 3", len(args))
			}

			obj := Evaluate(args[0], env)

            if obj.Type() == object.ERROR_OBJ {
                return obj
            }

			condition := evalTruthy(obj)

			if condition {
				return Evaluate(args[1], env)
			}

			if len(args) == 3 {
				return Evaluate(args[2], env)
			}

			return NULL
		},
		"def": func(env *object.Environment, args ...ast.Expression) object.Object {
			if len(args) != 2 {
				return wrongNumOfArgsError("def", "2", len(args))
			}

			ident, ok := args[0].(*ast.Identifier)

			if !ok {
				err := fmt.Sprintf("cannot assign to non-identifier %s", args[0].String())
				return &object.ErrorObject{Error: err}
			}

			val := Evaluate(args[1], env)

            if val.Type() != object.ERROR_OBJ {
			    env.Set(ident.String(), val)
            }


			return val
		},
		"lambda": func(env *object.Environment, args ...ast.Expression) object.Object {
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
		},
		"str": func(env *object.Environment, args ...ast.Expression) object.Object {
			objects := []object.Object{}

			for _, arg := range args {
				obj := Evaluate(arg, env)

                if obj.Type() == object.ERROR_OBJ {
                    return obj
                }

				objects = append(objects, obj)
			}

			var result bytes.Buffer

			for _, obj := range objects {
				result.WriteString(obj.Inspect())
			}

			return &object.String{
				Value: result.String(),
			}
		},
		"print": func(env *object.Environment, args ...ast.Expression) object.Object {
			objects := []string{}

			for _, arg := range args {
				obj := Evaluate(arg, env)

                if obj.Type() == object.ERROR_OBJ {
                    return obj
                }

				objects = append(objects, obj.Inspect())
			}

			fmt.Println(strings.Join(objects, " "))

			return NULL
		},
		"get": func(env *object.Environment, args ...ast.Expression) object.Object {
			if len(args) != 2 {
				wrongNumOfArgsError("get", "2", len(args))
			}

			dictObj := Evaluate(args[0], env)

			if dictObj.Type() != object.DICT_OBJ {
				errobj, ok := dictObj.(*object.ErrorObject)

				if ok {
					return errobj
				}

				err := fmt.Sprintf("attempted to get from %s(%s) instead of dict", dictObj.Type(), dictObj.Inspect())
				return &object.ErrorObject{
					Error: err,
				}
			}

            obj := Evaluate(args[1], env)

            if obj.Type() == object.ERROR_OBJ {
                return obj
            }

			key, ok := obj.(object.Hashable)

			if !ok {
				badKeyError(obj)
			}

			dict := dictObj.(*object.Dict)
			result, ok := dict.Values[key.HashKey()]

			if !ok {
				return NULL
			}

			return result.Value
		},
		"set": func(env *object.Environment, args ...ast.Expression) object.Object {
			if len(args) != 3 {
				wrongNumOfArgsError("get", "3", len(args))
			}

			dictObj := Evaluate(args[0], env)

			if dictObj.Type() != object.DICT_OBJ {
				if dictObj.Type() == object.ERROR_OBJ {
					return dictObj
				}

				err := fmt.Sprintf("attempted to get from %s(%s) instead of dict", dictObj.Type(), dictObj.Inspect())
				return &object.ErrorObject{
					Error: err,
				}
			}

            keyObj := Evaluate(args[1], env)

            if keyObj.Type() == object.ERROR_OBJ {
                return keyObj
            }

			key, ok := keyObj.(object.Hashable)

			if !ok {
				badKeyError(keyObj)
			}

			value := Evaluate(args[2], env)

            if value.Type() == object.ERROR_OBJ {
                return value
            }

			dict := dictObj.(*object.Dict)
			dict.Values[key.HashKey()] = object.DictPair{
				Key:   keyObj,
				Value: value,
			}

			return dict
		},
	}
}

func evalTruthy(obj object.Object) bool {
	boolean, ok := obj.(*object.BooleanObject)

	if !ok {
		if obj == NULL {
			return false
		}

		return true
	}

	return boolean.Value
}

func isInt(num float64) bool {
	return num == float64(int64(num))
}
