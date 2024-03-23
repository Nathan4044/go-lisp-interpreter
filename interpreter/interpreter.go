package interpreter

import (
	"lisp/compiler"
	"lisp/evaluator"
	"lisp/lexer"
	"lisp/object"
	"lisp/parser"
	"lisp/vm"
)

// Evaluate the expresions in the provided program,
// using the provided environment as its base.
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

// Compile the expressions in the provided program into bytecode, then
// execute the bytecode on a VM.
func RunCompiled(source string) (object.Object, []string) {
	l := lexer.New(source)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors) > 0 {
		return nil, p.Errors
	}

	c := compiler.New()
	err := c.Compile(program)

	if err != nil {
		return nil, []string{err.Error()}
	}

	v := vm.New(c.Bytecode())

	err = v.Run()

	if err != nil {
		return nil, []string{err.Error()}
	}

	return v.LastPoppedStackElem(), nil
}
