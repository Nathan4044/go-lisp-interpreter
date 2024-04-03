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
func RunCompiled(source string, constants *[]object.Object, symTable *compiler.SymbolTable, globals []object.Object) (object.Object, []string) {
	l := lexer.New(source)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors) > 0 {
		return nil, p.Errors
	}

	c := compiler.NewWithState(*constants, symTable)
	err := c.Compile(program)

	if err != nil {
		return nil, []string{err.Error()}
	}

	// ensure constants persist between runs
	*constants = c.Bytecode().Constants

	v := vm.NewWithState(c.Bytecode(), globals)

	err = v.Run()

	if err != nil {
		return nil, []string{err.Error()}
	}

	return v.LastPoppedStackElem(), nil
}
