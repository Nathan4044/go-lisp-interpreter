package vm

import (
	"fmt"
	"lisp/ast"
	"lisp/compiler"
	"lisp/lexer"
	"lisp/object"
	"lisp/parser"
	"testing"
)

func parse(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)

	return p.ParseProgram()
}

func testIntegerObject(expected int64, actual object.Object) error {
	result, ok := actual.(*object.Integer)

	if !ok {
		return fmt.Errorf("object is not integer: got=%T(%+v)", actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object has wrong value: got=%d want=%d", result.Value, expected)
	}

	return nil
}

func testBooleanObject(expected bool, actual object.Object) error {
	result, ok := actual.(*object.BooleanObject)

	if !ok {
		return fmt.Errorf("object is not boolean: got=%T(%+v)", actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object has wrong value: got=%t want=%t", result.Value, expected)
	}

	return nil
}

func testStringObject(expected string, actual object.Object) error {
	result, ok := actual.(*object.String)

	if !ok {
		return fmt.Errorf("object is not string: got=%T(%+v)", actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object has wrong value: got=%s want=%s", result.Value, expected)
	}

	return nil
}

type vmTestCase struct {
	input    string
	expected interface{}
}

func runVmTests(t *testing.T, tests []vmTestCase) {
	t.Helper()

	for _, tt := range tests {
		program := parse(tt.input)
		comp := compiler.New()

		err := comp.Compile(program)

		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		vm := New(comp.Bytecode())
		err = vm.Run()

		if err != nil {
			t.Fatalf("vm error: %s", err)
		}

		stackElem := vm.LastPoppedStackElem()

		testExpectedObject(t, tt.expected, stackElem)
	}
}

func testExpectedObject(t *testing.T, expected interface{}, actual object.Object) {
	t.Helper()

	switch expected := expected.(type) {
	case int:
		err := testIntegerObject(int64(expected), actual)

		if err != nil {
			t.Errorf("testIntegerObject failed: %s", err)
		}
	case bool:
		err := testBooleanObject(expected, actual)

		if err != nil {
			t.Errorf("testBooleanObject failed: %s", err)
		}
	case string:
		err := testStringObject(expected, actual)

		if err != nil {
			t.Errorf("testStringObject failed: %s", err)
		}
	case *object.Null:
		if actual != Null {
			t.Errorf("object is not null: %T(%+v)", actual, actual)
		}
	}
}

func TestIntegerArithmetic(t *testing.T) {
	tests := []vmTestCase{
		{"1", 1},
		{"2", 2},
		{"1 2", 2}, // deleteme
		// {"(+ 1 2)", 2}, // fixme
	}

	runVmTests(t, tests)
}

func TestBooleanExpressions(t *testing.T) {
	tests := []vmTestCase{
		{"true", true},
		{"false", false},
	}

	runVmTests(t, tests)
}

func TestConditionals(t *testing.T) {
	tests := []vmTestCase{
		{"(if true 10)", 10},
		{"(if false 10)", Null},
		{"(if true 10 20)", 10},
		{"(if false 10 20)", 20},
		{"(if 1 10)", 10},
		{"(if 1 10 20)", 10},
		{"(if (if false 10) 10 20)", 20},
		// todo: uncomment when functions are implemented
		// {"(if (< 1 2) 10)", 10},
		// {"(if (< 1 2) 10 20)", 10},
		// {"(if (> 1 2) 10 20)", 20},
		// {"(not (if false 10))", true},
	}

	runVmTests(t, tests)
}

func TestGlobalDefExpressions(t *testing.T) {
	tests := []vmTestCase{
		{"(def one 1) one", 1},
		{"(def one 1) (def two 2) one", 1},
		{"(def one 1) (def two one) two", 1},
	}

	runVmTests(t, tests)
}

func TestStringExpressions(t *testing.T) {
	tests := []vmTestCase{
		{"\"string\"", "string"},
		{"(def a \"string\") a", "string"},
	}

	runVmTests(t, tests)
}

func TestLambdaCalls(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
            (def func (lambda () 5))
            (func)
            `,
			expected: 5,
		},
		{
			input: `
            (def one (lambda () 1))
            (def two (lambda () (one)))
            (def three (lambda () (two)))
            (three)
            `,
			expected: 1,
		},
		{
			input: `
            (def truth (lambda () true))
            (def two (lambda () (if (truth) 2 1)))
            (two)
            `,
			expected: 2,
		},
		{
			input:    "((lambda ()))",
			expected: Null,
		},
	}

	runVmTests(t, tests)
}
