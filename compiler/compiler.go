// The compiler package contains the definition of the Compiler type, which
// compiles an AST node into Bytecode
package compiler

import (
	"lisp/ast"
	"lisp/code"
	"lisp/object"
)

// The Compiler is a struct that holds the result of calls to the Compile
// method
type Compiler struct {
	instructions code.Instructions // instructions generated from Compile
	constants    []object.Object   // constant expressions found during Compile
}

// Bytecode is a struct containing the instructions produced by a Compiler and
// returned by the Bytecode method
type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
}

// Return the address of a new Compiler instance
func New() *Compiler {
	return &Compiler{
		instructions: code.Instructions{},
		constants:    []object.Object{},
	}
}

// Compile an AST Expression into bytecode instructions. Return an error if there is
// a problem during the compilation step
func (c *Compiler) Compile(expr ast.Expression) error {
	switch expr := expr.(type) {
	case *ast.Program:
		for _, e := range expr.Expressions {
			err := c.Compile(e)

			if err != nil {
				return err
			}

			c.emit(code.OpPop)
		}
	case *ast.SExpression:
		for _, e := range expr.Args {
			err := c.Compile(e)

			if err != nil {
				return err
			}
		}
	case *ast.IntegerLiteral:
		integer := &object.Integer{Value: expr.Value}

		c.emit(code.OpConstant, c.addConstant(integer))
	}

	return nil
}

// Return a Bytecode instance containing the compiled instructions along with
// a slice of constant values.
func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.instructions,
		Constants:    c.constants,
	}
}

func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)

	return len(c.constants) - 1
}

func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	ins := code.Make(op, operands...)
	pos := c.addInstruction(ins)

	return pos
}

func (c *Compiler) addInstruction(ins []byte) int {
	posNewInstruction := len(c.instructions)

	c.instructions = append(c.instructions, ins...)

	return posNewInstruction
}
