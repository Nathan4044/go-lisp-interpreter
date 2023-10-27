package parser

import (
	"lisp/ast"
	"lisp/lexer"
	"strings"
	"testing"
)

func TestParseInteger(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{
			input:    "5",
			expected: 5,
		},
		{
			input:    "500",
			expected: 500,
		},
		{
			input:    "-5",
			expected: -5,
		},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)

		program := p.ParseProgram()

		if len(program.Expressions) != 1 {
			t.Fatalf("Wrong number of expressions. expected=%d, got=%d", 1, len(program.Expressions))
		}

		testIntegerLiteral(t, program.Expressions[0], tt.expected)
	}
}

func TestParseFloat(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{
			input:    "5.0",
			expected: 5,
		},
		{
			input:    "500.0",
			expected: 500,
		},
		{
			input:    "-5.0",
			expected: -5,
		},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)

		program := p.ParseProgram()

		if len(program.Expressions) != 1 {
			t.Fatalf("Wrong number of expressions. expected=%d, got=%d", 1, len(program.Expressions))
		}

		testFloatLiteral(t, program.Expressions[0], tt.expected)
	}
}

func TestParseString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			`"hello"`,
			"hello",
		},
		{
			`"5"`,
			"5",
		},
		{
			`"(test 'this')"`,
			"(test 'this')",
		},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()

		if len(program.Expressions) != 1 {
			t.Fatalf("Expected 1 expression, got %d", program.Expressions)
		}

		sl, ok := program.Expressions[0].(*ast.StringLiteral)

		if !ok {
			t.Fatalf("Expected expression to be StringLiteral, got %T", program.Expressions[0])
		}

		if sl.String() != tt.expected {
			t.Errorf("Expected %s, got %s", tt.expected, sl.String())
		}
	}
}

func TestParseIdentifier(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "+",
			expected: "+",
		},
		{
			input:    "add",
			expected: "add",
		},
		{
			input:    "add2things",
			expected: "add2things",
		},
		{
			input:    "this-func",
			expected: "this-func",
		},
		{
			input:    "this_func",
			expected: "this_func",
		},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()

		if len(program.Expressions) != 1 {
			t.Errorf("wrong number of expressions. expected=%d, got=%d", 1, len(program.Expressions))
		}

	}
}

func TestParseDict(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			`(dict)`,
			`(dict)`,
		},
        {
            `(dict a b)`,
            `(dict a b)`,
        },
		{
			`{}`,
			`(dict)`,
		},
        {
            `{a b}`,
            `(dict a b)`,
        },
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()

        if len(p.Errors) > 0 {
            t.Fatalf("Parser created errors:\n%s",
                strings.Join(p.Errors, "\n"))
        }

		if len(program.Expressions) != 1 {
			t.Fatalf("Expected %d expressions. got=%d%+v", 1, len(program.Expressions), program.Expressions)
		}

		sExpression, ok := program.Expressions[0].(*ast.SExpression)

		if !ok {
			t.Fatalf("Expected SExpression. got=%T(%+v)",
				program.Expressions[0], program.Expressions[0])
		}

		if sExpression.String() != tt.expected {
			t.Errorf("Expected=%s, got=%s", tt.expected, sExpression.String())
		}
	}
}

func TestSExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			`(list)`,
			`(list)`,
		},
		{
			`(list 1)`,
			`(list 1)`,
		},
		{
			`(list 1 2 3)`,
			`(list 1 2 3)`,
		},
		{
			`(list 1 (+ 1 1))`,
			`(list 1 (+ 1 1))`,
		},
		{
			`()`,
			`()`,
		},
		{
			`(((())))`,
			`(((())))`,
		},
		{
			`((if (> a b) list coll) 1 2)`,
			`((if (> a b) list coll) 1 2)`,
		},
		{
			`'(1 2 3)`,
			`(list 1 2 3)`,
		},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()

		if len(program.Expressions) != 1 {
			t.Fatalf("Expected %d expressions. got=%d%+v", 1, len(program.Expressions), program.Expressions)
		}

		sExpression, ok := program.Expressions[0].(*ast.SExpression)

		if !ok {
			t.Fatalf("Expected SExpression. got=%T(%+v)",
				program.Expressions[0], program.Expressions[0])
		}

		if sExpression.String() != tt.expected {
			t.Errorf("Expected=%s, got=%s", tt.expected, sExpression.String())
		}
	}
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

func testIntegerLiteral(t *testing.T, expr ast.Expression, expected int64) {
	integerLiteral, ok := expr.(*ast.IntegerLiteral)

	if !ok {
		t.Fatalf("wrong ast type. got=%T(%+v)", expr, expr)
	}

	if integerLiteral.Value != expected {
		t.Errorf("wrong integer value. expected=%d, got=%d", expected, integerLiteral.Value)
	}
}

func testFloatLiteral(t *testing.T, expr ast.Expression, expected float64) {
	floatLiteral, ok := expr.(*ast.FloatLiteral)

	if !ok {
		t.Fatalf("wrong ast type. got=%T(%+v)", expr, expr)
	}

	if floatLiteral.Value != expected {
		t.Errorf("wrong float value. expected=%f, got=%f", expected, floatLiteral.Value)
	}
}

func testIdentifier(t *testing.T, expr ast.Expression, expected string) {
	identifier, ok := expr.(*ast.Identifier)

	if !ok {
		t.Fatalf("not identifier, got=%T(%+v)",
			expr, expr)
	}

	if identifier.String() != expected {
		t.Errorf("expected=%s, got=%s", expected, identifier.String())
	}
}
