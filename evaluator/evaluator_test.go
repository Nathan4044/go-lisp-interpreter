package evaluator

import (
	"lisp/lexer"
	"lisp/object"
	"lisp/parser"
	"testing"
)

type evaluatorTest struct {
	input        string
	expected     interface{}
	expectedType string
}

func TestEvaluateIntegerLiteral(t *testing.T) {
	tests := []evaluatorTest{
		{
			input:    "6",
			expected: float64(6),
		},
		{
			input:    "600",
			expected: float64(600),
		},
		{
			input:    "6 600",
			expected: float64(600),
		},
		{
			input:    "-6",
			expected: float64(-6),
		},
		{
			input:    "(+ 1 2)",
			expected: float64(3),
		},
		{
			input:    "(+ 1 2 3)",
			expected: float64(6),
		},
		{
			input:    "(+)",
			expected: float64(0),
		},
	}

	runEvalTests(t, tests)
}

func TestEvaluateFloatLiteral(t *testing.T) {
	tests := []evaluatorTest{
		{
			input:    "6.0",
			expected: float64(6),
		},
		{
			input:    "600.0",
			expected: float64(600),
		},
		{
			input:    "6 600.0",
			expected: float64(600),
		},
		{
			input:    "-6.0",
			expected: float64(-6),
		},
	}

	runEvalTests(t, tests)
}

func TestEvaluateStringLiteral(t *testing.T) {
	tests := []evaluatorTest{
		{
			input:        `"hello"`,
			expected:     "hello",
			expectedType: "string",
		},
		{
			input:        `"(list 1 2 3)"`,
			expected:     "(list 1 2 3)",
			expectedType: "string",
		},
	}

	runEvalTests(t, tests)
}

func TestEvaluateListCall(t *testing.T) {
	tests := []struct {
		input              string
		expectedValueCount int
		expectedInspect    string
	}{
		{
			"()",
			0,
			"()",
		},
		{
			"(list)",
			0,
			"()",
		},
		{
			"(list 1)",
			1,
			"(1)",
		},
		{
			"(list 1 2)",
			2,
			"(1 2)",
		},
		{
			"(list)",
			0,
			"()",
		},
		{
			"(list (list 1 2 3))",
			1,
			"((1 2 3))",
		},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()
		env := object.NewEnvironment(nil)

		output := Evaluate(program, env)

		result, ok := output.(*object.List)

		if !ok {
			t.Fatalf("expected list, instead got %T(%+v)", output, output)
		}

		if len(result.Values) != tt.expectedValueCount {
			t.Errorf("expected %d values, got %d(%+v)", tt.expectedValueCount, len(result.Values), result.Values)
		}

		if result.Inspect() != tt.expectedInspect {
			t.Errorf("Expected %s, got %s", result.Inspect(), tt.expectedInspect)
		}
	}
}

func TestEvaluateBooleanLiteral(t *testing.T) {
	tests := []evaluatorTest{
		{
			input:    "true",
			expected: true,
		},
		{
			input:    "false",
			expected: false,
		},
		{
			input:    "(= 1 1)",
			expected: true,
		},
		{
			input:    "(= 1 1)",
			expected: true,
		},
		{
			input:    "(not false)",
			expected: true,
		},
		{
			input:    "(not true)",
			expected: false,
		},
		{
			input:    "(not (= 1 1 1))",
			expected: false,
		},
		{
			input:    "(and true true)",
			expected: true,
		},
		{
			input:    "(and true false)",
			expected: false,
		},
		{
			input:    "(and true (= 1 1 1))",
			expected: true,
		},
		{
			input:    "(and true (= 1 2 1) true)",
			expected: false,
		},
		{
			input:    "(and)",
			expected: true,
		},
		{
			input:    "(and 4)",
			expected: true,
		},
		{
			input:    "(or false true)",
			expected: true,
		},
		{
			input:    "(or false false true)",
			expected: true,
		},
		{
			input:    "(or false (= 1 2))",
			expected: false,
		},
		{
			input:    "(or false 1)",
			expected: true,
		},
	}

	runEvalTests(t, tests)
}

func TestIfExpression(t *testing.T) {
	tests := []evaluatorTest{
		{
			input: `(if true
            1)`,
			expected: float64(1),
		},
		{
			input: `(if false
            1)`,
			expected: nil,
		},
		{
			input: `(if 1
            1)`,
			expected: float64(1),
		},
		{
			input: `(if false
            1
            2)`,
			expected: float64(2),
		},
	}

	runEvalTests(t, tests)
}

func TestDefineExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
		envIdent string
		envValue float64
	}{
		{
			"(def x 1) x",
			1,
			"x",
			1,
		},
		{
			"(def x 1)",
			1,
			"x",
			1,
		},
		{
			"(def x 2) 1",
			1,
			"x",
			2,
		},
		{
			`
            (def x 2)
            (def y 3)
            (+ x y)`,
			5,
			"x",
			2,
		},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()
		env := object.NewEnvironment(nil)

		output := Evaluate(program, env)

		result, ok := output.(*object.Number)

		if !ok {
			t.Fatalf("expected Number, instead got %T(%+V)", output, output)
		}

		if result.Value != tt.expected {
			t.Errorf("expected %f, got %f", tt.expected, result.Value)
		}

		entry := env.Get(tt.envIdent)

		val, ok := entry.(*object.Number)

		if !ok {
			t.Fatalf("expected integer, instead got %T(%+V)", entry, entry)
		}

		if val.Value != tt.envValue {
			t.Errorf("expected %f, got %f", tt.envValue, val.Value)
		}
	}
}

func runEvalTests(t *testing.T, tests []evaluatorTest) {
	t.Helper()

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()
		env := object.NewEnvironment(nil)

		result := Evaluate(program, env)

		switch expected := tt.expected.(type) {
		case float64:
			testFloatLiteral(t, result, expected)
		case bool:
			testBooleanLiteral(t, result, expected)
		case string:
			switch tt.expectedType {
			case "string":
				testStringLiteral(t, result, expected)
			default:
				t.Errorf("invalid expected type %s", tt.expectedType)
			}
		case nil:
			if result != NULL {
				t.Errorf("expected NULL, got %q", result)
			}
		default:
			t.Errorf("invalid expected type %T(%+v)", expected, expected)
		}
	}
}

func testStringLiteral(t *testing.T, obj object.Object, expected string) {
	t.Helper()

	string, ok := obj.(*object.String)

	if !ok {
		t.Fatalf("expected string, got=%T(%+v)", obj, obj)
	}

	if string.Value != expected {
		t.Errorf("%s != %s", string.Value, expected)
	}
}

func testBooleanLiteral(t *testing.T, obj object.Object, expected bool) {
	t.Helper()

	result, ok := obj.(*object.BooleanObject)

	if !ok {
		t.Fatalf("expected bool, instead got %T(%+v)", obj, obj)
	}

	if result.Value != expected {
		t.Errorf("expected %t, got %t", expected, result.Value)
	}
}

func testFloatLiteral(t *testing.T, obj object.Object, expected float64) {
	t.Helper()

	float, ok := obj.(*object.Number)

	if !ok {
		t.Fatalf("expected float, got=%T(%+v)", obj, obj)
	}

	if float.Value != expected {
		t.Errorf("%f != %f", float.Value, expected)
	}
}
