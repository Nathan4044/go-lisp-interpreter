// Run the lisp interpreter.
package main

import (
	"fmt"
	"lisp/interpreter"
	"lisp/object"
	"lisp/repl"
	"os"
)

func main() {
	switch len(os.Args) {
	// if there are no args provided, evaluate from stdin
	case 1:
		repl.Start(os.Stdin, os.Stdout)
		// if a filename is provided, evaluate the code within the file
	case 2:
		fileContents, err := os.ReadFile(os.Args[1])

		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
			return
		}

		env := object.NewEnvironment(nil)
		result, errors := interpreter.Run(string(fileContents), env)

		if len(errors) > 0 {
			for _, err := range errors {
				fmt.Fprintf(os.Stderr, err)
			}
		} else {
			err, ok := result.(*object.ErrorObject)

			if ok {
				fmt.Fprintf(os.Stderr, err.Inspect())
			}
		}
		// Currently, execution of only one file is supported.
		// There are also no command options.
	default:
		fmt.Fprintf(os.Stderr, "expected only 1 filename")
	}
}
