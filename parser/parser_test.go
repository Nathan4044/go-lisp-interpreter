package parser

import (
	"lisp/ast"
	"lisp/lexer"
	"testing"
)

type parserTest struct {
	input        string
	expected     interface{}
	expectedType string
}

func TestParseInteger(t *testing.T) {
	tests := []parserTest{
		{
			input:    "5",
			expected: float64(5),
		},
		{
			input:    "500",
			expected: float64(500),
		},
		{
			input:    "-5",
			expected: float64(-5),
		},
	}

	runParserTests(t, tests)
}

func TestParseFloat(t *testing.T) {
	tests := []parserTest{
		{
			input:    "5.0",
			expected: float64(5),
		},
		{
			input:    "500.0",
			expected: float64(500),
		},
		{
			input:    "-5.0",
			expected: float64(-5),
		},
		{
			input:    "-5.2",
			expected: float64(-5.2),
		},
	}

	runParserTests(t, tests)
}

func TestParseString(t *testing.T) {
	tests := []parserTest{
		{
			input:        `"hello"`,
			expected:     "hello",
			expectedType: "string",
		},
		{
			input:        `"5"`,
			expected:     "5",
			expectedType: "string",
		},
		{
			input:        `"(test 'this')"`,
			expected:     "(test 'this')",
			expectedType: "string",
		},
	}

	runParserTests(t, tests)
}

func testStringLiteral(t *testing.T, expr ast.Expression, expected string) {
	str, ok := expr.(*ast.StringLiteral)

	if !ok {
		t.Fatalf("Expected expression to be StringLiteral, got %T", expr)
	}

	if str.String() != expected {
		t.Errorf("Expected %s, got %s", expected, str.String())
	}
}

func TestParseIdentifier(t *testing.T) {
	tests := []parserTest{
		{
			input:        "+",
			expected:     "+",
			expectedType: "identifier",
		},
		{
			input:        "add",
			expected:     "add",
			expectedType: "identifier",
		},
		{
			input:        "add2things",
			expected:     "add2things",
			expectedType: "identifier",
		},
		{
			input:        "this-func",
			expected:     "this-func",
			expectedType: "identifier",
		},
		{
			input:        "this_func",
			expected:     "this_func",
			expectedType: "identifier",
		},
	}

	runParserTests(t, tests)
}

func TestParseDict(t *testing.T) {
	tests := []parserTest{
		{
			input:        `(dict)`,
			expected:     `(dict)`,
			expectedType: "sExpression",
		},
		{
			input:        `(dict a b)`,
			expected:     `(dict a b)`,
			expectedType: "sExpression",
		},
		{
			input:        `{}`,
			expected:     `(dict)`,
			expectedType: "sExpression",
		},
		{
			input:        `{a b}`,
			expected:     `(dict a b)`,
			expectedType: "sExpression",
		},
	}

	runParserTests(t, tests)
}

func TestSExpression(t *testing.T) {
	tests := []parserTest{
		{
			input:        `(list)`,
			expected:     `(list)`,
			expectedType: "sExpression",
		},
		{
			input:        `(list 1)`,
			expected:     `(list 1)`,
			expectedType: "sExpression",
		},
		{
			input:        `(list 1 2 3)`,
			expected:     `(list 1 2 3)`,
			expectedType: "sExpression",
		},
		{
			input:        `(list 1 (+ 1 1))`,
			expected:     `(list 1 (+ 1 1))`,
			expectedType: "sExpression",
		},
		{
			input:        `()`,
			expected:     `()`,
			expectedType: "sExpression",
		},
		{
			input:        `(((())))`,
			expected:     `(((())))`,
			expectedType: "sExpression",
		},
		{
			input:        `((if (> a b) list coll) 1 2)`,
			expected:     `((if (> a b) list coll) 1 2)`,
			expectedType: "sExpression",
		},
		{
			input:        `'(1 2 3)`,
			expected:     `(list 1 2 3)`,
			expectedType: "sExpression",
		},
	}

	runParserTests(t, tests)
}

func TestParseMultipleExpressions(t *testing.T) {
	input := `
    (def one 1)
    (def two 2)
    (+ one two)`

	expected := `(def one 1)(def two 2)(+ one two)`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(program.Expressions) != 3 {
		t.Fatalf("Expected %d expressions. got=%d[%+v]", 3, len(program.Expressions), program.Expressions)
	}

	if program.String() != expected {
		t.Errorf("Expressions do not match. Expected=[%s] got=[%s]",
			expected, program.String())
	}
}

func runParserTests(t *testing.T, tests []parserTest) {
	t.Helper()

	for _, tt := range tests {

		l := lexer.New(tt.input)
		p := New(l)

		program := p.ParseProgram()

		if len(program.Expressions) != 1 {
			t.Fatalf("Wrong number of expressions. expected=%d, got=%d", 1, len(program.Expressions))
		}

		switch expected := tt.expected.(type) {
		case float64:
			testFloatLiteral(t, program.Expressions[0], expected)
		case string:
			switch tt.expectedType {
			case "string":
				testStringLiteral(t, program.Expressions[0], expected)
			case "identifier":
				testIdentifier(t, program.Expressions[0], expected)
			case "sExpression":
				testSExpression(t, program.Expressions[0], expected)
			default:
				t.Errorf("invalid expected type %s", tt.expectedType)
			}
		default:
			t.Errorf("invalid expected type %T(%+v)", expected, expected)
		}
	}
}

func testSExpression(t *testing.T, expr ast.Expression, expected string) {
	se, ok := expr.(*ast.SExpression)

	if !ok {
		t.Fatalf("Expected SExpression. got=%T(%+v)", expr, expr)
	}

	if se.String() != expected {
		t.Errorf("Expected=%s, got=%s", expected, se.String())
	}
}

func testFloatLiteral(t *testing.T, expr ast.Expression, expected float64) {
	t.Helper()

	floatLiteral, ok := expr.(*ast.FloatLiteral)

	if !ok {
		t.Fatalf("wrong ast type. got=%T(%+v)", expr, expr)
	}

	if floatLiteral.Value != expected {
		t.Errorf("wrong float value. expected=%f, got=%f", expected, floatLiteral.Value)
	}
}

func testIdentifier(t *testing.T, expr ast.Expression, expected string) {
	t.Helper()

	identifier, ok := expr.(*ast.Identifier)

	if !ok {
		t.Fatalf("not identifier, got=%T(%+v)",
			expr, expr)
	}

	if identifier.String() != expected {
		t.Errorf("expected=%s, got=%s", expected, identifier.String())
	}
}
