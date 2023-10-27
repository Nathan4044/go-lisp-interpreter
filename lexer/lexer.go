package lexer

import (
	"bytes"
	"fmt"
	"lisp/token"
)

type Lexer struct {
	Input   string
	pos     int
	readPos int
	ch      byte
}

func New(input string) *Lexer {
	l := &Lexer{
		Input: input,
	}

	l.pos = 0
	l.readPos = 1
	l.ch = l.Input[l.pos]

	return l
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()

	switch {
	case l.ch == '(':
		tok.Type = token.LPAREN
		tok.Literal = string(l.ch)
		l.readChar()
	case l.ch == ')':
		tok.Type = token.RPAREN
		tok.Literal = string(l.ch)
		l.readChar()
	case l.ch == '{':
		tok.Type = token.LBRACE
		tok.Literal = string(l.ch)
		l.readChar()
	case l.ch == '}':
		tok.Type = token.RBRACE
		tok.Literal = string(l.ch)
		l.readChar()
	case l.ch == '\'':
		tok.Type = token.QUOTE
		tok.Literal = string(l.ch)
		l.readChar()
	case l.ch == '-':
		if isNumber(l.peekChar()) {
			l.readChar()
			tok = l.readNumber()
			tok.Literal = "-" + tok.Literal
		} else {
			tok = l.readIdent()
		}
	case l.ch == '"':
		tok = l.readString()
	case isNumber(l.ch):
		tok = l.readNumber()
	case isValidIdentChar(l.ch):
		tok = l.readIdent()
	case l.ch == 0:
		tok.Type = token.EOF
		tok.Literal = ""
	default:
		tok.Type = token.ILLEGAL
		tok.Literal = string(l.ch)
	}

	return tok
}

func (l *Lexer) readChar() {
	l.pos++
	l.readPos = l.pos + 1

	if l.readPos >= len(l.Input)+1 {
		l.ch = 0
	} else {
		l.ch = l.Input[l.pos]
	}
}

func (l *Lexer) peekChar() byte {
	if l.readPos >= len(l.Input)+1 {
		return 0
	}

	return l.Input[l.readPos]
}

func (l *Lexer) readNumber() token.Token {
	start := l.pos

	for isNumber(l.ch) || l.ch == '.' {
		l.readChar()
	}

	return token.Token{
		Type:    token.NUM,
		Literal: l.Input[start:l.pos],
	}
}

func (l *Lexer) readIdent() token.Token {
	start := l.pos

	for isValidIdentChar(l.ch) {
		l.readChar()
	}

	return token.Token{
		Type:    token.IDENT,
		Literal: l.Input[start:l.pos],
	}
}

func (l *Lexer) readString() token.Token {
	l.readChar()

	var output bytes.Buffer

	for l.ch != '"' {
        if l.ch == 0 {
            return token.Token{
                Type: token.ILLEGAL,
                Literal: fmt.Sprintf("unterminated string: \"%s", output.String()),
            }
        }
		output.WriteByte(l.ch)
		l.readChar()
	}
	l.readChar()

	return token.Token{
		Type:    token.STRING,
		Literal: output.String(),
	}
}

func (l *Lexer) skipWhitespace() {
	for isWhitespace(l.ch) {
		l.readChar()
	}
}

func isValidIdentChar(ch byte) bool {
	return !isReservedChar(ch) && !isWhitespace(ch)
}

func isReservedChar(ch byte) bool {
	reserved := map[byte]bool{
		'(': true,
		')': true,
		'{': true,
		'}': true,
		0:   true,
	}

	_, ok := reserved[ch]

	return ok
}

func isNumber(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}
