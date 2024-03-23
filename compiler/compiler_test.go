package compiler

import (
	"fmt"
	"lisp/ast"
	"lisp/code"
	"lisp/lexer"
	"lisp/object"
	"lisp/parser"
	"slices"
	"testing"
)

type compilerTestCase struct {
	input                string
	expectedConstants    []interface{}
	expectedInstructions []code.Instructions
}

func TestIntegerArithmetic(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "(+ 1 2)",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				// todo: add call for plus
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 2 3",
			expectedConstants: []interface{}{1, 2, 3},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpPop),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func runCompilerTests(t *testing.T, tests []compilerTestCase) {
	t.Helper()

	for _, tt := range tests {
		program := parse(tt.input)
		compiler := New()

		err := compiler.Compile(program)

		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		bytecode := compiler.Bytecode()

		err = testInstructions(tt.expectedInstructions, bytecode.Instructions)

		if err != nil {
			t.Logf(program.String())
			t.Fatalf("testInstructions failed: %s", err)
		}

		err = testConstants(tt.expectedConstants, bytecode.Constants)

		if err != nil {
			t.Fatalf("testConstants failed: %s", err)
		}
	}
}

func parse(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)

	return p.ParseProgram()
}

func testInstructions(
	expected []code.Instructions,
	actual code.Instructions,
) error {
	concatted := slices.Concat(expected...)

	if len(actual) != len(concatted) {
		return fmt.Errorf(
			"wrong instruction length:\n  want=%q\n  got=%q",
			concatted, actual,
		)
	}

	for i, ins := range concatted {
		if actual[i] != ins {
			return fmt.Errorf(
				"wrong instruction at %d:\n  want=%q\n  got=%q",
				i, concatted, actual,
			)

		}
	}

	return nil
}

func testConstants(
	expected []interface{},
	actual []object.Object,
) error {
	if len(actual) != len(expected) {
		return fmt.Errorf(
			"wrong instruction length:\n  want=%q\n  got=%q",
			expected, actual,
		)
	}

	for i, constant := range expected {
		switch constant := constant.(type) {
		case int:
			err := testIntegerObject(int64(constant), actual[i])

			if err != nil {
				return fmt.Errorf("constant %d - testIntegerObject failed: %s", i, err)
			}
		}
	}

	return nil
}

func testIntegerObject(expected int64, actual object.Object) error {
	result, ok := actual.(*object.Integer)

	if !ok {
		return fmt.Errorf("object is not Integer: got=%T(%+v)", actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object has wrong value: got=%d want=%d", result.Value, expected)
	}

	return nil
}
