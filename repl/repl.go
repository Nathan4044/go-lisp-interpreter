// repl contains the function that starts an interactive session.
package repl

import (
	"bufio"
	"fmt"
	"io"
	"lisp/compiler"
	"lisp/evaluator"
	"lisp/lexer"
	"lisp/object"
	"lisp/parser"
	"lisp/vm"
)

const PROMPT = ">>> "

// Starts an interactive interpreter, conventionally in the terminal
// with stdin and stdout as the Reader and Writer.
func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment(nil)

	for {
		fmt.Fprintf(out, PROMPT)
		scanned := scanner.Scan()

		if !scanned {
			return
		}

		if len(scanner.Text()) == 0 {
			continue
		}

		l := lexer.New(scanner.Text())
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors) > 0 {
			for _, err := range p.Errors {
				fmt.Fprintf(out, err)
			}

			return
		}

		result := evaluator.Evaluate(program, env)
		fmt.Fprintln(out, result.Inspect())
	}
}

// Starts an interactive interpreter, conventionally in the terminal
// with stdin and stdout as the Reader and Writer.
func StartCompiled(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	constants := []object.Object{}
	globals := make([]object.Object, vm.GlobalSize)
	symbolTable := compiler.NewSymbolTable()

	for i, v := range object.Builtins {
		symbolTable.DefineBuiltin(i, v.Name)
	}

	for {
		fmt.Fprintf(out, PROMPT)
		scanned := scanner.Scan()

		if !scanned {
			return
		}

		if len(scanner.Text()) == 0 {
			continue
		}

		l := lexer.New(scanner.Text())
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors) > 0 {
			for _, err := range p.Errors {
				fmt.Fprintf(out, err)
			}

			return
		}

		c := compiler.NewWithState(constants, symbolTable)
		err := c.Compile(program)

		if err != nil {
			fmt.Fprintf(out, "compiler error: %s\n", err)
			continue
		}

		// preserve constants between commands
		constants = c.Bytecode().Constants

		v := vm.NewWithState(c.Bytecode(), globals)
		err = v.Run()

		if err != nil {
			fmt.Fprintf(out, "vm error: %s\n", err)
			continue
		}

		result := v.LastPoppedStackElem()

		fmt.Fprintln(out, result.Inspect())
	}
}
