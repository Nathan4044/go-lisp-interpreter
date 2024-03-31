// The compiler package contains the definition of the Compiler type, which
// compiles an AST node into Bytecode
package compiler

import (
	"fmt"
	"lisp/ast"
	"lisp/code"
	"lisp/object"
)

// The Compiler is a struct that holds the result of calls to the Compile
// method
type Compiler struct {
	instructions code.Instructions // instructions generated from Compile
	constants    []object.Object   // constant expressions found during Compile
	symbolTable  *SymbolTable      // a map from a source code symbol to its memory address
}

// Bytecode is a struct containing the instructions produced by a Compiler and
// returned by the Bytecode method
type Bytecode struct {
	Instructions code.Instructions // a collection of OpCodes stored as a slice of bytes
	Constants    []object.Object   // each of the constant values found in the program
}

// Return the address of a new Compiler instance
func New() *Compiler {
	return &Compiler{
		instructions: code.Instructions{},
		constants:    []object.Object{},
		symbolTable:  NewSymbolTable(),
	}
}

// Return the address of a new Compiler instance, which uses the constants and
// symbols that are passed to it.
//
// This is to maintain program state between compiler instances.
func NewWithState(constants []object.Object, symbolTable *SymbolTable) *Compiler {
	return &Compiler{
		instructions: code.Instructions{},
		constants:    constants,
		symbolTable:  symbolTable,
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

			// pop the top element of the stack after each top-level expression
			c.emit(code.OpPop)
		}
	case *ast.SExpression:
		switch expr.Fn.String() {
		case "if":
			err := c.compileIfExpression(expr)

			if err != nil {
				return err
			}
		case "def":
			if len(expr.Args) != 2 {
				return fmt.Errorf("incorrect number of values in def expression")
			}

			err := c.Compile(expr.Args[1])

			if err != nil {
				return err
			}

			name, ok := expr.Args[0].(*ast.Identifier)

			if !ok {
				return fmt.Errorf("first argument to def must be identifier")
			}

			symbol := c.symbolTable.Define(name.Token.Literal)

			c.emit(code.OpSetGlobal, symbol.Index)
			c.emit(code.OpGetGlobal, symbol.Index)
		default:
			for _, e := range expr.Args {
				err := c.Compile(e)

				if err != nil {
					return err
				}
			}
		}
	case *ast.IntegerLiteral:
		integer := &object.Integer{Value: expr.Value}

		c.emit(code.OpConstant, c.addConstant(integer))
	case *ast.StringLiteral:
		string := &object.String{Value: expr.Value}

		c.emit(code.OpConstant, c.addConstant(string))
	case *ast.Identifier:
		switch expr.String() {
		case "true":
			c.emit(code.OpTrue)
		case "false":
			c.emit(code.OpFalse)
		case "null":
			c.emit(code.OpNull)
		default:
			sym, ok := c.symbolTable.Resolve(expr.Token.Literal)

			if !ok {
				return fmt.Errorf("undefined variable %s", expr.Token.Literal)
			}

			c.emit(code.OpGetGlobal, sym.Index)
		}
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

// Create a new instruction associated with the OpCode and add it to the
// finished instructions.
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

// Overwrite a previous instruction with a new one.
//
// Only works for instructions of the same length.
func (c *Compiler) replaceInstruction(pos int, newInstruction []byte) {
	for i := 0; i < len(newInstruction); i++ {
		c.instructions[pos+i] = newInstruction[i]
	}
}

// Change the operand of the instruction at the provided position to the
// provided operand.
func (c *Compiler) changeOperand(opPos int, operand int) {
	op := code.Opcode(c.instructions[opPos])

	newInstruction := code.Make(op, operand)

	c.replaceInstruction(opPos, newInstruction)
}

// Compile an if expression to instructions, adding in a false path if one is
// not provided.
func (c *Compiler) compileIfExpression(expr *ast.SExpression) error {
	// args should consist of condition, consequence, and optional alternative
	if len(expr.Args) < 2 || len(expr.Args) > 3 {
		return fmt.Errorf("incorrect number of values in if expression")
	}

	condition := expr.Args[0]

	err := c.Compile(condition)

	if err != nil {
		return err
	}

	// emit conditional jump with erroneous destination, to be updated later
	// in the function to the start of the alternative
	conditionalJumpPos := c.emit(code.OpJumpWhenFalse, 9999)

	consequence := expr.Args[1]

	err = c.Compile(consequence)

	if err != nil {
		return err
	}

	// emit jump with erroneous destination, to be updated later in the function
	// to the end of the alternative
	jumpPos := c.emit(code.OpJump, 9999)

	positionAfterConsequence := len(c.instructions)
	c.changeOperand(conditionalJumpPos, positionAfterConsequence)

	if len(expr.Args) < 3 {
		// if no alternative is present, add null
		c.emit(code.OpNull)
	} else {
		alternative := expr.Args[2]

		err = c.Compile(alternative)

		if err != nil {
			return err
		}
	}

	positionAfterAlternative := len(c.instructions)
	c.changeOperand(jumpPos, positionAfterAlternative)

	return nil
}
