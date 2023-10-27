package interpreter

import (
	"lisp/evaluator"
	"lisp/lexer"
	"lisp/object"
	"lisp/parser"
)

func Run(source string, env *object.Environment) (object.Object, []string) {
    l := lexer.New(source)
    p := parser.New(l)
    program := p.ParseProgram()

    if len(p.Errors) > 0 {
        return nil, p.Errors
    } else {
        return evaluator.Evaluate(program, env), nil
    }
}
