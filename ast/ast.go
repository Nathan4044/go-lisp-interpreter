package ast

import (
	"bytes"
	"lisp/token"
)

type node interface {
	String() string
}

type Expression interface {
	node
	expression()
}

type Program struct {
	Expressions []Expression
}

func (p *Program) String() string {
	var output bytes.Buffer

	for _, e := range p.Expressions {
		output.WriteString(e.String())
	}

	return output.String()
}

func (p *Program) expression() {}

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

type SExpression struct {
	Fn   Expression
	Args []Expression
}

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
