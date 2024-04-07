package main

import (
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
	var duration time.Duration
	var result object.Object

	fmt.Println("recursively calculating the 35th fibonacci number:")

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	start := time.Now()

	env := object.NewEnvironment(nil)
	result = evaluator.Evaluate(program, env)

	duration = time.Since(start)

	fmt.Printf("engine=%s result=%s duration=%s\n",
		"eval", result.Inspect(), duration)

	start = time.Now()

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

	fmt.Printf("engine=%s result=%s duration=%s\n",
		"vm", result.Inspect(), duration)
}
