package parser

import (
	"fmt"
	"lisp/ast"
	"lisp/lexer"
	"lisp/token"
	"strconv"
)

type Parser struct {
	lexer     *lexer.Lexer
	curToken  token.Token
	peekToken token.Token
	Errors    []string
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		lexer: l,
	}
	p.readToken()

	return p
}

func (p *Parser) ParseProgram() ast.Program {
	expressions := []ast.Expression{}

	p.readToken()
	for p.curToken.Type != token.EOF {
		expressions = append(expressions, p.parseExpression())
	}

	return ast.Program{
		Expressions: expressions,
	}
}

func (p *Parser) parseExpression() ast.Expression {
	switch p.curToken.Type {
	case token.NUM:
		int, err := strconv.ParseInt(p.curToken.Literal, 10, 64)

		if err == nil {
			tok := p.curToken
			p.readToken()
			return &ast.IntegerLiteral{
				Token: tok,
				Value: int,
			}
		}

		float, err := strconv.ParseFloat(p.curToken.Literal, 64)

		if err == nil {
			tok := p.curToken
			p.readToken()
			return &ast.FloatLiteral{
				Token: tok,
				Value: float,
			}
		}

		errMsg := fmt.Sprintf("%s is invalid number", p.curToken.Literal)
		p.Errors = append(p.Errors, errMsg)
		return nil
	case token.STRING:
		string := &ast.StringLiteral{
			Token: p.curToken,
			Value: p.curToken.Literal,
		}
		p.readToken()
		return string
	case token.IDENT:
		ident := &ast.Identifier{Token: p.curToken}
		p.readToken()
		return ident
	case token.LPAREN:
		return p.parseSExpression()
	case token.LBRACE:
		return p.parseDictLiteral()
	case token.QUOTE:
		return p.parseQuoteExpression()
	case token.EOF:
		return nil
	default:
		errorMessage := fmt.Sprintf("should not reach here:\n\treceived: %+v\n\tpeek: %s", p.curToken, p.peekToken)
		p.Errors = append(p.Errors, errorMessage)
		p.readToken()
		return nil
	}
}

func (p *Parser) parseSExpression() ast.Expression {
	sExpression := &ast.SExpression{}

	p.readToken()

	if p.curToken.Type == token.RPAREN {
		p.readToken()
		return sExpression
	}

	sExpression.Fn = p.parseExpression()

	if p.curToken.Type == token.RPAREN {
		p.readToken()
		return sExpression
	}

	args := []ast.Expression{}

	for p.curToken.Type != token.RPAREN {
		if p.curToken.Type == token.EOF {
			p.Errors = append(
				p.Errors,
				"Reached EOF before ')'",
			)
			return sExpression
		}
		args = append(args, p.parseExpression())
	}

	p.readToken()
	sExpression.Args = args
	return sExpression
}

func (p *Parser) parseDictLiteral() ast.Expression {
	sExpression := &ast.SExpression{}
	sExpression.Fn = &ast.Identifier{
		Token: token.Token{
			Type:    token.IDENT,
			Literal: "dict",
		},
	}

	p.readToken()

	if p.curToken.Type == token.RBRACE {
		p.readToken()
		return sExpression
	}

	args := []ast.Expression{}

	for p.curToken.Type != token.RBRACE {
		if p.curToken.Type == token.EOF {
			p.Errors = append(
				p.Errors,
				"Reached EOF before '}'",
			)
			return sExpression
		}
		args = append(args, p.parseExpression())
	}

	p.readToken()
	sExpression.Args = args
	return sExpression
}

func (p *Parser) parseQuoteExpression() ast.Expression {
	sExpression := &ast.SExpression{}

	p.readToken()

	if p.curToken.Type != token.LPAREN {
		p.Errors = append(
			p.Errors,
			"' not followed by (",
		)
		return sExpression
	}
	p.readToken()

	sExpression.Fn = &ast.Identifier{
		Token: token.Token{
			Type:    token.IDENT,
			Literal: "list",
		},
	}

	if p.curToken.Type == token.RPAREN {
		p.readToken()
		return sExpression
	}

	args := []ast.Expression{}

	for p.curToken.Type != token.RPAREN {
		if p.curToken.Type == token.EOF {
			p.Errors = append(
				p.Errors,
				"Reached EOF before ')'",
			)
			return sExpression
		}
		args = append(args, p.parseExpression())
	}

	p.readToken()
	sExpression.Args = args
	return sExpression
}

func (p *Parser) readToken() token.Token {
	p.curToken = p.peekToken

	if p.curToken.Type != token.EOF {
		p.peekToken = p.lexer.NextToken()
	}

	return p.curToken
}
