package token

type TokenType string

const (
	EOF     = "eof"
	ILLEGAL = "illegal"

	NUM    = "number"
	STRING = "string"
	IDENT  = "identifier"

	LPAREN = "lparen"
	RPAREN = "rparen"
    LBRACE = "lbrace"
    RBRACE = "rbrace"

	QUOTE = "quote"
)

type Token struct {
	Type    TokenType
	Literal string
}
