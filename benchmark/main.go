package main

import (
	"flag"
	"fmt"
	"lisp/compiler"
	"lisp/evaluator"
	"lisp/lexer"
	"lisp/object"
	"lisp/parser"
	"lisp/vm"
	"os"
	"time"
)

var engine *string = flag.String("engine", "vm", "select 'eval' or 'vm'")

var input string = `
(def fibonacci (lambda (n)
    (if (or (= n 0)
            (= n 1))
        n
        (+ (fibonacci (- n 1))
           (fibonacci (- n 2))))))
(fibonacci 35)
`

func main() {
	flag.Parse()

	var duration time.Duration
	var result object.Object

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if *engine == "vm" {
		start := time.Now()

		c := compiler.New()
		err := c.Compile(program)

		if err != nil {
			fmt.Fprintf(os.Stderr, "compiler error: %s", err)
		}

		vm := vm.New(c.Bytecode())
		err = vm.Run()

		if err != nil {
			fmt.Fprintf(os.Stderr, "vm error: %s", err)
		}

		duration = time.Since(start)
		result = vm.LastPoppedStackElem()
	} else {
		start := time.Now()

		env := object.NewEnvironment(nil)
		result = evaluator.Evaluate(program, env)

		duration = time.Since(start)
	}

	fmt.Printf("engine=%s result=%s duration=%s\n",
		*engine, result.Inspect(), duration)
}
