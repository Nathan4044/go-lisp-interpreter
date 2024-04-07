# Lisp Interpreter

This is a project to learn interpreter concepts. It implements a simple lisp language.
The contents of this project were built using what I learned from the Thorsten Ball's 'Building a X in go' books.

## Basic Usage

Note: it is recommended that you use `rlwrap` when running the repl.

### Build

Build an artefact with `go build` to produce a binary of the project.

### Run

Run the repl with `./lisp`, implemented commands are:
```
+, *, -, /, rem, =, <, >, not, and, or, list, dict, first, rest,
len, push, if, def, lambda, str, print, get, set
```

Run the interpreter with a source file by passing the file as an argument: `./lisp [file]`.
An example file is available in the examples directory.

By default, lisp will now run in the `vm` engine by default, instead of the previous `eval` engine.
The `eval` engine is still usable by passing it as a flag during execution:

`./lisp -engine=eval`

#### Engines

##### Eval
Eval is the tree walking interpreter engine this project originated with.

##### VM
VM compiles the AST produced by the parser into bytecode, which is then executed on in a virtual machine.

#### Examples

Small example files of lisp programs have been written and added to the `examples` directory.

#### Benchmark

A simple benchmark has been written to demonstrate the difference in execution speed between the original tree walking interpreter and the compiled solution. You can run it with `go run benchmark/main.go` to see the difference in time it takes to calculate the 35th fibonacci number between the two methods.

### Test

Run all the tests with `go test ./...`.
Clear previous test results with `go clean -testcache`.
