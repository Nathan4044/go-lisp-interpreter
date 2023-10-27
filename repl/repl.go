package repl

import (
	"bufio"
	"fmt"
	"io"
	"lisp/interpreter"
	"lisp/object"
)

const PROMPT = ">>> "

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

		result, errors := interpreter.Run(scanner.Text(), env)

		if len(errors) > 0 {
			for _, err := range errors {
				fmt.Fprintln(out, err)
			}
		} else {
			fmt.Fprintln(out, result.Inspect())
		}
	}
}
