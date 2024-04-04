// Contains the type definitions used to create AST nodes.
package ast

import (
	"bytes"
	"lisp/token"
)

// Base interface for all Expressions.
//
// expression() is an empty method used
// for interface satisfaction only.
type Expression interface {
	expression()
	String() string
}

// Used as an alias to identify the
// whole program.
type Program struct {
	Expressions []Expression
}

// Prints the each expression inside
// of the program.
func (p *Program) String() string {
	var output bytes.Buffer

	for _, e := range p.Expressions {
		output.WriteString(e.String())
	}

	return output.String()
}

func (p *Program) expression() {}

// Identifiers are variable names.
type Identifier struct {
	Token token.Token
}

func (i *Identifier) String() string {
	return i.Token.Literal
}

func (i *Identifier) expression() {}

type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) String() string {
	return il.Token.Literal
}

func (il *IntegerLiteral) expression() {}

type FloatLiteral struct {
	Token token.Token
	Value float64
}

func (fl *FloatLiteral) String() string {
	return fl.Token.Literal
}

func (fl *FloatLiteral) expression() {}

type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) String() string {
	return sl.Token.Literal
}

func (sl *StringLiteral) expression() {}

// SExpressions are the lisp representation of a function call.
//
// Fn represents the function `func` and Args represents the
// arguments arg1, arg2, etc. in the expression:
//
// (func arg1 arg2 arg3)
//
// SExpression fulfills the Expression interface, so both Fn
// and any arg can also be an SExpression.
type SExpression struct {
	Fn   Expression
	Args []Expression
	Name string
}

// Recursively print the values in the SExpression.
func (se *SExpression) String() string {
	var output bytes.Buffer

	output.WriteString("(")
	if se.Fn != nil {
		output.WriteString(se.Fn.String())
	}

	for _, arg := range se.Args {
		output.WriteString(" ")
		output.WriteString(arg.String())
	}

	output.WriteString(")")

	return output.String()
}

func (se *SExpression) expression() {}
