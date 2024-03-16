// Define the lexer object
package lexer

import (
	"bytes"
	"fmt"
	"lisp/token"
)

const EOF byte = 0

// A Lexer is an object that transforms the input text
// into tokens until reaching an EOF.
type Lexer struct {
	Input   string // The source code text.
	pos     int    // The current character position in the text.
	readPos int    // The position of the next character.
	ch      byte   // The currently highlighted character.
}

// Create a new lexer object that will tokenize the given
// input text.
func New(input string) *Lexer {
	l := &Lexer{
		Input: input,
	}

	l.pos = 0
	l.readPos = 1
	l.ch = l.Input[l.pos]

	return l
}

// Read bytes from input until a complete token is formed.
// Return the newly created token.
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
	case l.ch == EOF:
		tok.Type = token.EOF
		tok.Literal = ""
	default:
		tok.Type = token.ILLEGAL
		tok.Literal = string(l.ch)
	}

	return tok
}

// Update the position, read position, and
// the current character fields in the lexer.
//
// If the read position is beyond the end of
// the input, return EOF.
func (l *Lexer) readChar() {
	l.pos++
	l.readPos = l.pos + 1

	if l.readPos >= len(l.Input)+1 {
		l.ch = EOF
	} else {
		l.ch = l.Input[l.pos]
	}
}

// See the next character in the input.
//
// If the read position is beyond the end of
// the input, return EOF.
func (l *Lexer) peekChar() byte {
	if l.readPos >= len(l.Input)+1 {
		return EOF
	}

	return l.Input[l.readPos]
}

// Read characters until either reaching whitespace or
// a reserved character. Return a Token of type number
// with the literal value of a string of the read characters.
func (l *Lexer) readNumber() token.Token {
	start := l.pos

	for !isWhitespace(l.ch) && !isReservedChar(l.ch) {
		l.readChar()
	}

	return token.Token{
		Type:    token.NUM,
		Literal: l.Input[start:l.pos],
	}
}

// Read characters until either reaching whitespace or
// a reserved character. Return a Token of type identifier
// with the literal value of a string of the read characters.
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

// Read characters until reaching a terminating `"`.
// Return a Token of type identifier string with
// the literal value of a string of the read characters.
func (l *Lexer) readString() token.Token {
	l.readChar()

	var output bytes.Buffer

	for l.ch != '"' {
		if l.ch == 0 {
			return token.Token{
				Type:    token.ILLEGAL,
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

// Checks if the provided character is in a set of
// reserved characters that can't be part of another
// token.
func isReservedChar(ch byte) bool {
	reserved := map[byte]bool{
		'(': true,
		')': true,
		'{': true,
		'}': true,
		EOF: true,
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
