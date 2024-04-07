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

// A struct for holding the values involved in a compiler test.
type compilerTestCase struct {
	// The source code that will be compiled.
	input string
	// The constant values that should be extracted from the source code.
	expectedConstants []interface{}
	// The instructions that should be built.
	expectedInstructions []code.Instructions
}

// Ensure integer and float literals are comiled correctly.
func TestNumberLiterals(t *testing.T) {
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
		{
			input:             "1.3",
			expectedConstants: []interface{}{1.3},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

// Ensure that boolean literals are resolved to their unique Opcodes.
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

// Ensure if expressions compile as expected.
func TestConditionals(t *testing.T) {
	// Bytecode position numbers provided for debugging clarity.
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

// Test that variables defined in the global scope are compiled correctly.
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

// Ensure variables defined locally are compiled as such, and that global
// variables referenced locally aren't defined as being local.
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
				code.Make(code.OpClosure, 1, 0),
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
				code.Make(code.OpClosure, 1, 0),
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
				code.Make(code.OpClosure, 2, 0),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

// Test that strings are compiled correctly to constants.
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

// Test that lambdas are correctly compiled to Closure objects.
func TestLambdaExpressions(t *testing.T) {
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
				code.Make(code.OpClosure, 1, 0),
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
				code.Make(code.OpClosure, 2, 0),
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
				code.Make(code.OpClosure, 0, 0),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

// Test that lambdas compiled to Closures will be called correctly.
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
				code.Make(code.OpClosure, 1, 0),
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
				code.Make(code.OpClosure, 1, 0),
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
				code.Make(code.OpClosure, 0, 0),
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
				code.Make(code.OpClosure, 0, 0),
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

// Test that references to builtin functions are compiled correctly.
func TestBuiltinReferences(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `
            (+ 1 2)
            (= 1 1 2)
            `,
			expectedConstants: []interface{}{
				1, 2, 1, 1, 2,
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpGetBuiltin, 0),
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpCall, 2),
				code.Make(code.OpPop),
				code.Make(code.OpGetBuiltin, 5),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpCall, 3),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
            (lambda () (+ 2 3))
            `,
			expectedConstants: []interface{}{
				2, 3,
				[]code.Instructions{
					code.Make(code.OpGetBuiltin, 0),
					code.Make(code.OpConstant, 0),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpCall, 2),
					code.Make(code.OpReturn),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 2, 0),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

// Ensure actual closures compile as expected.
func TestClosures(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `
            (lambda (a)
              (lambda (b)
                (+ a b)))
            `,
			expectedConstants: []interface{}{
				[]code.Instructions{
					code.Make(code.OpGetBuiltin, 0),
					code.Make(code.OpGetFree, 0),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpCall, 2),
					code.Make(code.OpReturn),
				},
				[]code.Instructions{
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpClosure, 0, 1),
					code.Make(code.OpReturn),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 1, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
            (lambda (a)
              (lambda (b)
                (lambda (c)
                  (+ a b c))))
            `,
			expectedConstants: []interface{}{
				[]code.Instructions{
					code.Make(code.OpGetBuiltin, 0),
					code.Make(code.OpGetFree, 0),
					code.Make(code.OpGetFree, 1),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpCall, 3),
					code.Make(code.OpReturn),
				},
				[]code.Instructions{
					code.Make(code.OpGetFree, 0),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpClosure, 0, 2),
					code.Make(code.OpReturn),
				},
				[]code.Instructions{
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpClosure, 1, 1),
					code.Make(code.OpReturn),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 2, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
            (def a 1)
            (lambda ()
              (def b 2)
              (lambda ()
                (def c 3)
                (lambda ()
                  (def d 4)
                  (+ a b c d))))
            `,
			expectedConstants: []interface{}{
				1,
				2,
				3,
				4,
				[]code.Instructions{
					code.Make(code.OpConstant, 3),
					code.Make(code.OpSetLocal, 0),
					code.Make(code.OpPop),
					code.Make(code.OpGetBuiltin, 0),
					code.Make(code.OpGetGlobal, 0),
					code.Make(code.OpGetFree, 0),
					code.Make(code.OpGetFree, 1),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpCall, 4),
					code.Make(code.OpReturn),
				},
				[]code.Instructions{
					code.Make(code.OpConstant, 2),
					code.Make(code.OpSetLocal, 0),
					code.Make(code.OpPop),
					code.Make(code.OpGetFree, 0),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpClosure, 4, 2),
					code.Make(code.OpReturn),
				},
				[]code.Instructions{
					code.Make(code.OpConstant, 1),
					code.Make(code.OpSetLocal, 0),
					code.Make(code.OpPop),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpClosure, 5, 1),
					code.Make(code.OpReturn),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpPop),
				code.Make(code.OpClosure, 6, 0),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

// Ensure that recursive references to closures are compiled correctly.
func TestRecursiveClosures(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `
            (def countdown (lambda (n) (countdown (- n 1))))
            `,
			expectedConstants: []interface{}{
				1,
				[]code.Instructions{
					code.Make(code.OpCurrentClosure),
					code.Make(code.OpGetBuiltin, 2),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpConstant, 0),
					code.Make(code.OpCall, 2),
					code.Make(code.OpCall, 1),
					code.Make(code.OpReturn),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 1, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
            (def wrapper (lambda ()
                (def countdown (lambda (n)
                    (countdown (- n 1))))
                (countdown 10)))
            (wrapper)
            `,
			expectedConstants: []interface{}{
				1,
				[]code.Instructions{
					code.Make(code.OpCurrentClosure),
					code.Make(code.OpGetBuiltin, 2),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpConstant, 0),
					code.Make(code.OpCall, 2),
					code.Make(code.OpCall, 1),
					code.Make(code.OpReturn),
				},
				10,
				[]code.Instructions{
					code.Make(code.OpClosure, 1, 0),
					code.Make(code.OpSetLocal, 0),
					code.Make(code.OpPop),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpConstant, 2),
					code.Make(code.OpCall, 1),
					code.Make(code.OpReturn),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 3, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpPop),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpCall, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
            (def exbo (lambda (n)
                (if (= n 1)
                    n
                    (* n (exbo (- n 1))))))
            (exbo 4)
            `,
			expectedConstants: []interface{}{
				1,
				1,
				[]code.Instructions{
					// 0000
					code.Make(code.OpGetBuiltin, 5),
					// 0002
					code.Make(code.OpGetLocal, 0),
					// 0004
					code.Make(code.OpConstant, 0),
					// 0007
					code.Make(code.OpCall, 2),
					// 0009
					code.Make(code.OpJumpWhenFalse, 17),
					// 0012
					code.Make(code.OpGetLocal, 0),
					// 0014
					code.Make(code.OpJump, 35),
					// 0017
					code.Make(code.OpGetBuiltin, 1),
					// 0019
					code.Make(code.OpGetLocal, 0),
					// 0021
					code.Make(code.OpCurrentClosure),
					// 0022
					code.Make(code.OpGetBuiltin, 2),
					// 0024
					code.Make(code.OpGetLocal, 0),
					// 0026
					code.Make(code.OpConstant, 1),
					// 0029
					code.Make(code.OpCall, 2),
					// 0031
					code.Make(code.OpCall, 1),
					// 0033
					code.Make(code.OpCall, 2),
					// 0035
					code.Make(code.OpReturn),
				},
				4,
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 2, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpPop),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpCall, 1),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
            (def reduce (lambda (lst f acc) 
                (if (= 0 (len lst)) 
                    acc
                    (reduce (rest lst) f (f acc (first lst))))))
            (def map (lambda (lst f)
                (def func (lambda (acc n)
                    (push acc (f n))))
                (reduce lst func '())))
            (def l '(1 2 3))
            (map l (lambda (n) (* 2 n)))
            `,
			expectedConstants: []interface{}{
				0,
				[]code.Instructions{
					// 0000
					code.Make(code.OpGetBuiltin, 5),
					// 0002
					code.Make(code.OpConstant, 0),
					// 0005
					code.Make(code.OpGetBuiltin, 16),
					// 0007
					code.Make(code.OpGetLocal, 0),
					// 0009
					code.Make(code.OpCall, 1),
					// 0011
					code.Make(code.OpCall, 2),
					// 0013
					code.Make(code.OpJumpWhenFalse, 21),
					// 0016
					code.Make(code.OpGetLocal, 2),
					// 0018
					code.Make(code.OpJump, 44),
					// 0021
					code.Make(code.OpCurrentClosure),
					// 0022
					code.Make(code.OpGetBuiltin, 14),
					// 0024
					code.Make(code.OpGetLocal, 0),
					// 0026
					code.Make(code.OpCall, 1),
					// 0028
					code.Make(code.OpGetLocal, 1),
					// 0030
					code.Make(code.OpGetLocal, 1),
					// 0032
					code.Make(code.OpGetLocal, 2),
					// 0034
					code.Make(code.OpGetBuiltin, 13),
					// 0036
					code.Make(code.OpGetLocal, 0),
					// 0038
					code.Make(code.OpCall, 1),
					// 0040
					code.Make(code.OpCall, 2),
					// 0042
					code.Make(code.OpCall, 3),
					// 0044
					code.Make(code.OpReturn),
				},
				[]code.Instructions{
					code.Make(code.OpGetBuiltin, 17),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpGetFree, 0),
					code.Make(code.OpGetLocal, 1),
					code.Make(code.OpCall, 1),
					code.Make(code.OpCall, 2),
					code.Make(code.OpReturn),
				},
				[]code.Instructions{
					code.Make(code.OpGetLocal, 1),
					code.Make(code.OpClosure, 2, 1),
					code.Make(code.OpSetLocal, 2),
					code.Make(code.OpPop),
					code.Make(code.OpGetGlobal, 0),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpGetLocal, 2),
					code.Make(code.OpGetBuiltin, 11),
					code.Make(code.OpCall, 0),
					code.Make(code.OpCall, 3),
					code.Make(code.OpReturn),
				},
				1,
				2,
				3,
				2,
				[]code.Instructions{
					code.Make(code.OpGetBuiltin, 1),
					code.Make(code.OpConstant, 7),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpCall, 2),
					code.Make(code.OpReturn),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 1, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpPop),
				code.Make(code.OpClosure, 3, 0),
				code.Make(code.OpSetGlobal, 1),
				code.Make(code.OpPop),
				code.Make(code.OpGetBuiltin, 11),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpConstant, 5),
				code.Make(code.OpConstant, 6),
				code.Make(code.OpCall, 3),
				code.Make(code.OpSetGlobal, 2),
				code.Make(code.OpPop),
				code.Make(code.OpGetGlobal, 1),
				code.Make(code.OpGetGlobal, 2),
				code.Make(code.OpClosure, 8, 0),
				code.Make(code.OpCall, 2),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
            (def reduce (lambda (lst f acc) 
                (if (= 0 (len lst)) 
                    acc
                    (reduce (rest lst) f (f acc (first lst))))))
            (def map (lambda (lst f)
                (reduce lst (lambda (acc n) (push acc (f n))) '())))
            (def l '(1 2 3))
            (map l (lambda (n) (* 2 n)))
            `,
			expectedConstants: []interface{}{
				0,
				[]code.Instructions{
					// 0000
					code.Make(code.OpGetBuiltin, 5),
					// 0002
					code.Make(code.OpConstant, 0),
					// 0005
					code.Make(code.OpGetBuiltin, 16),
					// 0007
					code.Make(code.OpGetLocal, 0),
					// 0009
					code.Make(code.OpCall, 1),
					// 0011
					code.Make(code.OpCall, 2),
					// 0013
					code.Make(code.OpJumpWhenFalse, 21),
					// 0016
					code.Make(code.OpGetLocal, 2),
					// 0018
					code.Make(code.OpJump, 44),
					// 0021
					code.Make(code.OpCurrentClosure),
					// 0022
					code.Make(code.OpGetBuiltin, 14),
					// 0024
					code.Make(code.OpGetLocal, 0),
					// 0026
					code.Make(code.OpCall, 1),
					// 0028
					code.Make(code.OpGetLocal, 1),
					// 0030
					code.Make(code.OpGetLocal, 1),
					// 0032
					code.Make(code.OpGetLocal, 2),
					// 0034
					code.Make(code.OpGetBuiltin, 13),
					// 0036
					code.Make(code.OpGetLocal, 0),
					// 0038
					code.Make(code.OpCall, 1),
					// 0040
					code.Make(code.OpCall, 2),
					// 0042
					code.Make(code.OpCall, 3),
					// 0044
					code.Make(code.OpReturn),
				},
				[]code.Instructions{
					code.Make(code.OpGetBuiltin, 17),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpGetFree, 0),
					code.Make(code.OpGetLocal, 1),
					code.Make(code.OpCall, 1),
					code.Make(code.OpCall, 2),
					code.Make(code.OpReturn),
				},
				[]code.Instructions{
					code.Make(code.OpGetGlobal, 0),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpGetLocal, 1),
					code.Make(code.OpClosure, 2, 1),
					code.Make(code.OpGetBuiltin, 11),
					code.Make(code.OpCall, 0),
					code.Make(code.OpCall, 3),
					code.Make(code.OpReturn),
				},
				1,
				2,
				3,
				2,
				[]code.Instructions{
					code.Make(code.OpGetBuiltin, 1),
					code.Make(code.OpConstant, 7),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpCall, 2),
					code.Make(code.OpReturn),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 1, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpPop),
				code.Make(code.OpClosure, 3, 0),
				code.Make(code.OpSetGlobal, 1),
				code.Make(code.OpPop),
				code.Make(code.OpGetBuiltin, 11),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpConstant, 5),
				code.Make(code.OpConstant, 6),
				code.Make(code.OpCall, 3),
				code.Make(code.OpSetGlobal, 2),
				code.Make(code.OpPop),
				code.Make(code.OpGetGlobal, 1),
				code.Make(code.OpGetGlobal, 2),
				code.Make(code.OpClosure, 8, 0),
				code.Make(code.OpCall, 2),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

// Ensure that scopes are entered and exited correctly during compilation.
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

// Run a compiler test case by:
//  1. Compiling the provided source code, ensuring no errors.
//  2. Testing that the compiled instructions match the expected instructions.
//  3. Testing that the compiled constants match the expected constants.
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

// Helper function for getting a parsed program for testing.
func parse(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)

	return p.ParseProgram()
}

// Test that the compiled instructions match the expected ones.
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

// Test that the compiled constants match the expected ones.
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
		case float64:
			err := testFloatObject(constant, actual[i])

			if err != nil {
				return fmt.Errorf("constant %d - testFloatObject failed: %s", i, err)
			}
		case string:
			err := testStringObject(constant, actual[i])

			if err != nil {
				return fmt.Errorf("constant %d - testStringObject failed: %s", i, err)
			}
			// Test that the constant CompiledLambda matches the expected instructions.
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

func testFloatObject(expected float64, actual object.Object) error {
	result, ok := actual.(*object.Float)

	if !ok {
		return fmt.Errorf("object is not Float: got=%T(%+v)", actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object has wrong value: got=%f want=%f", result.Value, expected)
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
