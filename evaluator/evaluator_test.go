package evaluator

import (
	"lisp/lexer"
	"lisp/object"
	"lisp/parser"
	"testing"
)

func TestEvaluateIntegerLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{
			"6",
			6,
		},
		{
			"600",
			600,
		},
		{
			"6 600",
			600,
		},
		{
			"-6",
			-6,
		},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()
		env := object.NewEnvironment(nil)

		result := Evaluate(program, env)

		integer, ok := result.(*object.Integer)

		if !ok {
			t.Fatalf("expected integer, got=%T(%+v)", result, result)
		}

		if integer.Value != tt.expected {
			t.Errorf("%d != %d", integer.Value, tt.expected)
		}
	}
}

func TestEvaluateFloatLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{
			"6.0",
			6,
		},
		{
			"600.0",
			600,
		},
		{
			"6 600.0",
			600,
		},
		{
			"-6.0",
			-6,
		},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()
		env := object.NewEnvironment(nil)

		result := Evaluate(program, env)

		float, ok := result.(*object.Float)

		if !ok {
			t.Fatalf("expected float, got=%T(%+v)", result, result)
		}

		if float.Value != tt.expected {
			t.Errorf("%f != %f", float.Value, tt.expected)
		}
	}
}

func TestEvaluateStringLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			`"hello"`,
			"hello",
		},
		{
			`"(list 1 2 3)"`,
			"(list 1 2 3)",
		},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()
		env := object.NewEnvironment(nil)

		result := Evaluate(program, env)

		string, ok := result.(*object.String)

		if !ok {
			t.Fatalf("expected string, got=%T(%+v)", result, result)
		}

		if string.Value != tt.expected {
			t.Errorf("%s != %s", string.Value, tt.expected)
		}
	}
}

func TestEvaluateEmptyList(t *testing.T) {
	input := "()"
	expected := "()"

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment(nil)

	result := Evaluate(program, env)

	obj, ok := result.(*object.List)

	if !ok {
		t.Fatalf("expected List, got=%T(%+v)", obj, obj)
	}

	if obj.Values != nil {
		t.Errorf("empty list has args: %+v", obj.Values)
	}

	if obj.Inspect() != expected {
		t.Errorf("inspect wrong: expected=%s got=%s", expected, obj.Inspect())
	}
}

func TestEvaluateFunctionCall(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{
			"(+ 1 2)",
			3,
		},
		{
			"(+ 1 2 3)",
			6,
		},
		{
			"(+)",
			0,
		},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()
		env := object.NewEnvironment(nil)

		result := Evaluate(program, env)

		integer, ok := result.(*object.Integer)

		if !ok {
			t.Fatalf("expected integer, got=%T(%+v)", result, result)
		}

		if integer.Value != tt.expected {
			t.Errorf("%d != %d", integer.Value, tt.expected)
		}
	}
}

func TestEvaluateListCall(t *testing.T) {
	tests := []struct {
		input              string
		expectedValueCount int
		expectedInspect    string
	}{
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
	tests := []struct {
		input    string
		expected bool
	}{
		{
			"true",
			true,
		},
		{
			"false",
			false,
		},
		{
			"(= 1 1)",
			true,
		},
		{
			"(= 1 1)",
			true,
		},
		{
			"(not false)",
			true,
		},
		{
			"(not true)",
			false,
		},
		{
			"(not (= 1 1 1))",
			false,
		},
		{
			"(and true true)",
			true,
		},
		{
			"(and true false)",
			false,
		},
		{
			"(and true (= 1 1 1))",
			true,
		},
		{
			"(and true (= 1 2 1) true)",
			false,
		},
		{
			"(and)",
			true,
		},
		{
			"(and 4)",
			true,
		},
		{
			"(or false true)",
			true,
		},
		{
			"(or false false true)",
			true,
		},
		{
			"(or false (= 1 2))",
			false,
		},
		{
			"(or false 1)",
			true,
		},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()
		env := object.NewEnvironment(nil)

		output := Evaluate(program, env)

		result, ok := output.(*object.BooleanObject)

		if !ok {
			t.Fatalf("expected bool, instead got %T(%+v)", output, output)
		}

		if result.Value != tt.expected {
			t.Errorf("expected %t, got %t", tt.expected, result.Value)
		}
	}
}

func TestIfExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			`(if true
                1)`,
			int64(1),
		},
		{
			`(if false
                1)`,
			nil,
		},
		{
			`(if 1
                1)`,
			int64(1),
		},
		{
			`(if false
                1
                2)`,
			int64(2),
		},
	}

	for _, tt := range tests {
		t.Logf("testing %s", tt.input)
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()
		env := object.NewEnvironment(nil)

		output := Evaluate(program, env)

		if output == NULL && tt.expected == nil {
			continue
		}

		result, ok := output.(*object.Integer)

		if !ok {
			t.Fatalf("expected integer, instead got %T(%+V)", output, output)
		}

		if result.Value != tt.expected {
			t.Errorf("expected %d, got %d", tt.expected, result.Value)
		}
	}
}

func TestDefineExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
		envIdent string
		envValue int64
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

		result, ok := output.(*object.Integer)

		if !ok {
			t.Fatalf("expected integer, instead got %T(%+V)", output, output)
		}

		if result.Value != tt.expected {
			t.Errorf("expected %d, got %d", tt.expected, result.Value)
		}

		entry := env.Get(tt.envIdent)

		val, ok := entry.(*object.Integer)

		if !ok {
			t.Fatalf("expected integer, instead got %T(%+V)", entry, entry)
		}

		if val.Value != tt.envValue {
			t.Errorf("expected %d, got %d", tt.envValue, val.Value)
		}
	}
}
