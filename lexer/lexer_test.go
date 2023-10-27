package lexer

import (
	"lisp/token"
	"testing"
)

func TestLexer(t *testing.T) {
	input := `
    (add 
        (+ 1 2) 
        (- 18 -1 2))
    '(list)
    "hello (string)"
    {12.4}`

	expected := []token.Token{
		{
			Type:    token.LPAREN,
			Literal: "(",
		},
		{
			Type:    token.IDENT,
			Literal: "add",
		},
		{
			Type:    token.LPAREN,
			Literal: "(",
		},
		{
			Type:    token.IDENT,
			Literal: "+",
		},
		{
			Type:    token.NUM,
			Literal: "1",
		},
		{
			Type:    token.NUM,
			Literal: "2",
		},
		{
			Type:    token.RPAREN,
			Literal: ")",
		},
		{
			Type:    token.LPAREN,
			Literal: "(",
		},
		{
			Type:    token.IDENT,
			Literal: "-",
		},
		{
			Type:    token.NUM,
			Literal: "18",
		},
		{
			Type:    token.NUM,
			Literal: "-1",
		},
		{
			Type:    token.NUM,
			Literal: "2",
		},
		{
			Type:    token.RPAREN,
			Literal: ")",
		},
		{
			Type:    token.RPAREN,
			Literal: ")",
		},
		{
			Type:    token.QUOTE,
			Literal: "'",
		},
		{
			Type:    token.LPAREN,
			Literal: "(",
		},
		{
			Type:    token.IDENT,
			Literal: "list",
		},
		{
			Type:    token.RPAREN,
			Literal: ")",
		},
		{
			Type:    token.STRING,
			Literal: "hello (string)",
		},
		{
			Type:    token.LBRACE,
			Literal: "{",
		},
		{
			Type:    token.NUM,
			Literal: "12.4",
		},
		{
			Type:    token.RBRACE,
			Literal: "}",
		},
		{
			Type:    token.EOF,
			Literal: "",
		},
	}

	l := New(input)

	for _, expectedToken := range expected {
		tok := l.NextToken()

		if tok != expectedToken {
			t.Errorf("expected %q, got %q", expectedToken, tok)
		}
	}
}
