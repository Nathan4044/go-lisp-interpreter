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
    case 1:
        repl.Start(os.Stdin, os.Stdout)
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
    default:
        fmt.Fprintf(os.Stderr, "expected only 1 filename")
    }
}
