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
			input:             "1 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
				code.Make(code.OpConstant, 1),
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

func TestBooleanExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "true",
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "false",
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpFalse),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestConditionals(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "(if true 4) 5",
			expectedConstants: []interface{}{4, 5},
			expectedInstructions: []code.Instructions{
				// 0000
				code.Make(code.OpTrue),
				// 0001
				code.Make(code.OpJumpWhenFalse, 10),
				// 0004
				code.Make(code.OpConstant, 0),
				// 0007
				code.Make(code.OpJump, 11),
				// 0010
				code.Make(code.OpNull),
				// 0011
				code.Make(code.OpPop),
				// 0008
				code.Make(code.OpConstant, 1),
				// 0011
				code.Make(code.OpPop),
			},
		},
		{
			input:             "(if true 4 10) 5",
			expectedConstants: []interface{}{4, 10, 5},
			expectedInstructions: []code.Instructions{
				// 0000
				code.Make(code.OpTrue),
				// 0001
				code.Make(code.OpJumpWhenFalse, 10),
				// 0004
				code.Make(code.OpConstant, 0),
				// 0007
				code.Make(code.OpJump, 13),
				// 0010
				code.Make(code.OpConstant, 1),
				// 0013
				code.Make(code.OpPop),
				// 0014
				code.Make(code.OpConstant, 2),
				// 0017
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestGlobalDefExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "(def a 10)",
			expectedConstants: []interface{}{10},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "(def a 10) (def b 20)",
			expectedConstants: []interface{}{10, 20},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpPop),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpSetGlobal, 1),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "(def a 10) (def b a)",
			expectedConstants: []interface{}{10},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpPop),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 1),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestLocalDefExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `
            (def x 10)
            (lambda () x)
            `,
			expectedConstants: []interface{}{
				10,
				[]code.Instructions{
					code.Make(code.OpGetGlobal, 0),
					code.Make(code.OpReturn),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpPop),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
            (lambda ()
              (def x 10)
              x)
            `,
			expectedConstants: []interface{}{
				10,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpSetLocal, 0),
					code.Make(code.OpPop),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpReturn),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 1),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
            (lambda ()
              (def x 10)
              (def y 15)
              x)
            `,
			expectedConstants: []interface{}{
				10,
				15,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpSetLocal, 0),
					code.Make(code.OpPop),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpSetLocal, 1),
					code.Make(code.OpPop),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpReturn),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 2),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestStringExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "\"string\"",
			expectedConstants: []interface{}{"string"},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestFunctions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: "(lambda () 5)",
			expectedConstants: []interface{}{
				5,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpReturn),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 1),
				code.Make(code.OpPop),
			},
		},
		{
			input: "(lambda () 5 10)",
			expectedConstants: []interface{}{
				5,
				10,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpPop),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpReturn),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 2),
				code.Make(code.OpPop),
			},
		},
		{
			input: "(lambda ())",
			expectedConstants: []interface{}{
				[]code.Instructions{
					code.Make(code.OpNull),
					code.Make(code.OpReturn),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestLambdaCalls(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: "((lambda () 9))",
			expectedConstants: []interface{}{
				9,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpReturn),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 1),
				code.Make(code.OpCall, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: "(def func (lambda () 9)) (func)",
			expectedConstants: []interface{}{
				9,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpReturn),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 1),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpPop),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpCall, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
            (def oneArg (lambda (a) a))
            (oneArg 9)
            `,
			expectedConstants: []interface{}{
				[]code.Instructions{
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpReturn),
				},
				9,
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpPop),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpCall, 1),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
            (def manyArgs (lambda (a b c) a b c))
            (manyArgs 1 2 3)
            `,
			expectedConstants: []interface{}{
				[]code.Instructions{
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpPop),
					code.Make(code.OpGetLocal, 1),
					code.Make(code.OpPop),
					code.Make(code.OpGetLocal, 2),
					code.Make(code.OpReturn),
				},
				1, 2, 3,
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpPop),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpCall, 3),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestCompilerScopes(t *testing.T) {
	compiler := New()

	if compiler.scopeIndex != 0 {
		t.Errorf("scopeIndex wrong: got=%d want=%d", compiler.scopeIndex, 0)
	}

	compiler.emit(code.OpTrue)

	globalScope := compiler.symbolTable

	compiler.enterScope()

	if compiler.scopeIndex != 1 {
		t.Errorf("scopeIndex wrong: got=%d want=%d", compiler.scopeIndex, 1)
	}

	compiler.emit(code.OpFalse)

	if len(compiler.scopes[compiler.scopeIndex].instructions) != 1 {
		t.Errorf(
			"instructions length wrong: got=%d want=%d",
			len(compiler.scopes[compiler.scopeIndex].instructions),
			1,
		)
	}

	last := compiler.scopes[compiler.scopeIndex].lastInstruction

	if last.Opcode != code.OpFalse {
		t.Errorf(
			"lastInstruction.OpCode wrong: got=%d want=%d",
			last, code.OpFalse,
		)
	}

	if compiler.symbolTable.outer != globalScope {
		t.Errorf("compiler did not enter new symbol table")
	}

	compiler.leaveScope()

	if compiler.symbolTable != globalScope {
		t.Errorf("compiler modified globalScope incorrectly")
	}

	if compiler.scopeIndex != 0 {
		t.Errorf("scopeIndex wrong: got=%d want=%d", compiler.scopeIndex, 0)
	}

	compiler.emit(code.OpNull)

	if len(compiler.scopes[compiler.scopeIndex].instructions) != 2 {
		t.Errorf(
			"instructions length wrong: got=%d want=%d",
			len(compiler.scopes[compiler.scopeIndex].instructions),
			2,
		)
	}

	last = compiler.scopes[compiler.scopeIndex].lastInstruction

	if last.Opcode != code.OpNull {
		t.Errorf(
			"lastInstruction.OpCode wrong: got=%d want=%d",
			last, code.OpNull,
		)
	}

	previous := compiler.scopes[compiler.scopeIndex].previousInstruction

	if previous.Opcode != code.OpTrue {
		t.Errorf(
			"previousInstruction.OpCode wrong: got=%d want=%d",
			previous, code.OpJump,
		)
	}
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
		case string:
			err := testStringObject(constant, actual[i])

			if err != nil {
				return fmt.Errorf("constant %d - testStringObject failed: %s", i, err)
			}
		case []code.Instructions:
			lambda, ok := actual[i].(*object.CompiledLambda)

			if !ok {
				return fmt.Errorf("constant %d - not a function: %T", actual[i], actual[i])
			}

			err := testInstructions(constant, lambda.Instructions)

			if err != nil {
				return fmt.Errorf("constant %d - testInstructions failed: %s", i, err)
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

func testStringObject(expected string, actual object.Object) error {
	result, ok := actual.(*object.String)

	if !ok {
		return fmt.Errorf("object is not String: got=%T(%+v)", actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object has wrong value: got=%s want=%s", result.Value, expected)
	}

	return nil
}
