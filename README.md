# Lisp Interpreter

This is a project to learn interpreter concepts. It implements a simple lisp language.

## Basic Usage

Note: it is recommended that you use `rlwrap` for running the repl.

### Build

Build an artefact with `go build` to produce a binary of the project.

### Run

Run the repl with `./lisp`, implemented commands are:
```
+, *, -, /, rem, =, <, >, not, and, or, list, dict, first, rest, len, push, push!, pop!, if, def, lambda, str, print, get, set
```

Run the interpreter with a source file by passing the file as an argument: `./lisp [file]`.
An example file is available in the examples directory.

### Test

Run all the tests with `go test ./...`.
Clear previous test results with `go clean -testcache`.
