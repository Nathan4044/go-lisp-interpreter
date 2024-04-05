// Run the lisp interpreter.
package main

import (
	"flag"
	"fmt"
	"lisp/compiler"
	"lisp/evaluator"
	"lisp/lexer"
	"lisp/object"
	"lisp/parser"
	"lisp/repl"
	"lisp/vm"
	"os"
)

var engine *string = flag.String("engine", "vm", "enter 'vm' or 'eval'")

func main() {
	flag.Parse()
	fmt.Printf("%+v\n", flag.Args())
	switch len(flag.Args()) {
	// if there are no args provided, evaluate from stdin
	case 0:
		if *engine == "eval" {
			repl.Start(os.Stdin, os.Stdout)
		} else {
			repl.StartCompiled(os.Stdin, os.Stdout)
		}
		// if a filename is provided, evaluate the code within the file
	case 1:
		// Currently, execution of only one file is supported.
		// There are also no command options.
		fileContents, err := os.ReadFile(flag.Arg(0))

		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
			return
		}

		if *engine == "eval" {
			runFile(string(fileContents))
		} else {
			runCompiled(string(fileContents))
		}
	default:
		fmt.Fprintf(os.Stderr, "expected only 1 filename")
	}
}

// Convert the provided program into an AST, then evluate it.
func runFile(source string) {
	l := lexer.New(source)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors) > 0 {
		for _, err := range p.Errors {
			fmt.Fprintf(os.Stderr, err)
		}

		return
	}

	env := object.NewEnvironment(nil)
	result := evaluator.Evaluate(program, env)

	fmt.Println(result.Inspect())
}

// Compile the expressions in the provided program into bytecode, then
// execute the bytecode on a VM.
func runCompiled(source string) {
	l := lexer.New(source)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors) > 0 {
		for _, err := range p.Errors {
			fmt.Fprintf(os.Stderr, err)
		}

		return
	}

	c := compiler.New()
	err := c.Compile(program)

	if err != nil {
		fmt.Fprintf(os.Stderr, "compiler error: %s", err)
	}

	v := vm.New(c.Bytecode())
	err = v.Run()

	if err != nil {
		fmt.Fprintf(os.Stderr, "vm error: %s", err)
	}

	fmt.Println(v.LastPoppedStackElem())
}
