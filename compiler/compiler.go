// The compiler package contains the definition of the Compiler type, which
// compiles an AST node into Bytecode.
package compiler

import (
	"fmt"
	"lisp/ast"
	"lisp/code"
	"lisp/object"
)

// A representation of an instruction.
type EmittedInstruction struct {
	Opcode   code.Opcode // Opcode associated with the instruction
	Position int         // the position in instructions where this instruction begins
}

// A CompilationScope represents the current level in which instructions are
// being compiled. This allows for functions to be compiled in their own scope
// and then returned as instructions for use as its own object.
type CompilationScope struct {
	instructions        code.Instructions //instructions generated from Compile
	lastInstruction     EmittedInstruction
	previousInstruction EmittedInstruction
}

// The Compiler is a struct that holds the result of calls to the Compile
// method.
type Compiler struct {
	constants   []object.Object    // constant expressions found during Compile
	symbolTable *SymbolTable       // a map from a source code symbol to its memory address
	scopes      []CompilationScope // a stack of currently used scopes
	scopeIndex  int                // the currently active scope
}

// Bytecode is a struct containing the instructions produced by a Compiler and
// returned by the Bytecode method.
type Bytecode struct {
	Instructions code.Instructions // a collection of OpCodes stored as a slice of bytes
	Constants    []object.Object   // each of the constant values found in the program
}

// Return the address of a new Compiler instance.
func New() *Compiler {
	mainScope := CompilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}

	symbolTable := NewSymbolTable()

	// Define builtin functions for the global scope.
	for i, v := range object.Builtins {
		symbolTable.DefineBuiltin(i, v.Name)
	}

	return &Compiler{
		constants:   []object.Object{},
		symbolTable: symbolTable,
		scopes:      []CompilationScope{mainScope},
	}
}

// Return the address of a new Compiler instance, which uses the constants and
// symbols that are passed to it.
//
// This is to maintain program state between compiler instances.
func NewWithState(constants []object.Object, symbolTable *SymbolTable) *Compiler {
	mainScope := CompilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}

	return &Compiler{
		constants:   constants,
		symbolTable: symbolTable,
		scopes:      []CompilationScope{mainScope},
	}
}

// Compile an AST Expression into bytecode instructions. Return an error if there is
// a problem during the compilation step.
func (c *Compiler) Compile(expr ast.Expression) error {
	switch expr := expr.(type) {
	case *ast.Program:
		for _, e := range expr.Expressions {
			err := c.Compile(e)

			if err != nil {
				return err
			}

			// Pop the top element of the stack after each top-level expression.
			c.emit(code.OpPop)
		}
	case *ast.SExpression:
		// Conditionally compile an SExpression based on the first element.
		if expr.Fn == nil {
			c.emit(code.OpEmptyList)
		} else {
			var err error

			switch expr.Fn.String() {
			case "if":
				err = c.compileIfExpression(expr)
			case "def":
				err = c.compileDefExpression(expr)
			case "lambda":
				err = c.compileLambdaExpression(expr)
			default:
				err = c.compileCallExpression(expr)
			}

			if err != nil {
				return err
			}
		}
	case *ast.IntegerLiteral:
		integer := &object.Integer{Value: expr.Value}

		c.emit(code.OpConstant, c.addConstant(integer))
	case *ast.FloatLiteral:
		float := &object.Float{Value: expr.Value}

		c.emit(code.OpConstant, c.addConstant(float))
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

			c.getSymbol(sym)
		}
	}

	return nil
}

// Returns the instructions from the currently active scope.
func (c *Compiler) currentInstructions() code.Instructions {
	return c.scopes[c.scopeIndex].instructions
}

// Return a Bytecode instance containing the compiled instructions along with
// a slice of constant values.
func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.currentInstructions(),
		Constants:    c.constants,
	}
}

func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)

	return len(c.constants) - 1
}

// Create a new instruction associated with the Opcode and add it to the
// finished instructions.
func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	ins := code.Make(op, operands...)
	pos := c.addInstruction(ins)

	c.setLastInstruction(op, pos)
	return pos
}

// Set the value of the last instruction emitted in the current scope. Also
// update the previous instruction emitted.
func (c *Compiler) setLastInstruction(op code.Opcode, pos int) {
	c.scopes[c.scopeIndex].previousInstruction = c.scopes[c.scopeIndex].lastInstruction

	c.scopes[c.scopeIndex].lastInstruction = EmittedInstruction{
		Opcode:   op,
		Position: pos,
	}
}

// Append the provided instruction to the instructions of the current scope.
func (c *Compiler) addInstruction(ins []byte) int {
	posNewInstruction := len(c.currentInstructions())

	c.scopes[c.scopeIndex].instructions = append(
		c.scopes[c.scopeIndex].instructions, ins...,
	)

	return posNewInstruction
}

// Overwrite a previous instruction with a new one.
//
// Only works for instructions of the same length.
func (c *Compiler) replaceInstruction(pos int, newInstruction []byte) {
	ins := c.currentInstructions()

	for i := 0; i < len(newInstruction); i++ {
		ins[pos+i] = newInstruction[i]
	}
}

// Change the operand of the instruction at the provided position to the
// provided operand.
func (c *Compiler) changeOperand(opPos int, operand int) {
	op := code.Opcode(c.currentInstructions()[opPos])

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

	// Emit conditional jump with erroneous destination, to be updated later
	// in the function to the start of the alternative.
	conditionalJumpPos := c.emit(code.OpJumpWhenFalse, 9999)

	consequence := expr.Args[1]

	err = c.Compile(consequence)

	if err != nil {
		return err
	}

	// Emit jump with erroneous destination, to be updated later in the function
	// to the end of the alternative.
	jumpPos := c.emit(code.OpJump, 9999)

	// Update the conditional jump instruction's destination to be directly
	// after the consequence of the if expression.
	positionAfterConsequence := len(c.currentInstructions())
	c.changeOperand(conditionalJumpPos, positionAfterConsequence)

	if len(expr.Args) < 3 {
		// Add null as the alternative result of if expressions where no
		// alternative is provided.
		c.emit(code.OpNull)
	} else {
		alternative := expr.Args[2]

		err = c.Compile(alternative)

		if err != nil {
			return err
		}
	}

	// Update the jump instruction's destination to be directly after the
	// alternative of the if expression.
	positionAfterAlternative := len(c.currentInstructions())
	c.changeOperand(jumpPos, positionAfterAlternative)

	return nil
}

// Compile the provided SExpression as a def expression, defining a variable
// in the current scope as the result of the internal Expression provided as the
// second argument.
func (c *Compiler) compileDefExpression(expr *ast.SExpression) error {
	if len(expr.Args) != 2 {
		return fmt.Errorf("incorrect number of values in def expression")
	}

	name, ok := expr.Args[0].(*ast.Identifier)

	if !ok {
		return fmt.Errorf("first argument to def must be identifier")
	}

	symbol := c.symbolTable.Define(name.Token.Literal)

	if sExpr, ok := expr.Args[1].(*ast.SExpression); ok {
		sExpr.Name = name.Token.Literal
	}

	err := c.Compile(expr.Args[1])

	if err != nil {
		return err
	}

	if c.symbolTable.outer == nil {
		c.emit(code.OpSetGlobal, symbol.Index)
	} else {
		c.emit(code.OpSetLocal, symbol.Index)
	}

	return nil
}

// Compile the provided SExpression as a Lambda Expression, resulting in a
// Closure object (all lambdas are treated as closures).
func (c *Compiler) compileLambdaExpression(expr *ast.SExpression) error {
	if len(expr.Args) < 1 {
		return fmt.Errorf("not enough arguments for lambda definition")
	}

	c.enterScope()

	if expr.Name != "" {
		c.symbolTable.DefineFunctionName(expr.Name)
	}

	paramList, ok := expr.Args[0].(*ast.SExpression)

	if !ok {
		return fmt.Errorf("provided args must be a list")
	}

	params := []ast.Expression{}

	if paramList.Fn != nil {
		params = append([]ast.Expression{paramList.Fn}, paramList.Args...)
	}

	for _, p := range params {
		param, ok := p.(*ast.Identifier)

		if !ok {
			return fmt.Errorf("function parameters must be identifiers, got=%T(%+v)", p, params)
		}

		c.symbolTable.Define(param.String())
	}

	expressions := expr.Args[1:]

	if len(expressions) == 0 {
		// Result in null when lambda contians no expressions.
		c.emit(code.OpNull)
		c.emit(code.OpReturn)
	} else {
		for _, arg := range expressions {
			err := c.Compile(arg)

			if err != nil {
				return err
			}

			c.emit(code.OpPop)
		}

		// Change the final pop instruction into a return instruction.
		c.replaceInstruction(
			len(c.currentInstructions())-1,
			[]byte{byte(code.OpReturn)},
		)
	}

	// Take free symbols found during compilation before leaving the inner scope
	// so the values can be added to the produced Closure.
	freeSymbols := c.symbolTable.FreeSymbols
	localsCount := c.symbolTable.count
	ins := c.leaveScope()

	compiledLambda := &object.CompiledLambda{
		Instructions:   ins,
		LocalsCount:    localsCount,
		ParameterCount: len(params),
	}

	// Put values associated with free symbols on the stack in front of the
	// Closure.
	for _, sym := range freeSymbols {
		c.getSymbol(sym)
	}

	c.emit(code.OpClosure, c.addConstant(compiledLambda), len(freeSymbols))

	return nil
}

// Compile the provided SExpression as a call to a function, resulting in a call
// instruction with an operand representing the number of arguments passed in,
// which sit on the stack above the function to be called.
func (c *Compiler) compileCallExpression(expr *ast.SExpression) error {
	err := c.Compile(expr.Fn)

	if err != nil {
		return err
	}

	for _, a := range expr.Args {
		err := c.Compile(a)

		if err != nil {
			return err
		}
	}

	c.emit(code.OpCall, len(expr.Args))

	return nil
}

// Push a new scope into the Compiler's scope stack and use it as the active
// scope.
func (c *Compiler) enterScope() {
	scope := CompilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}

	c.scopes = append(c.scopes, scope)
	c.scopeIndex++

	c.symbolTable = NewEnclosedSymbolTable(c.symbolTable)
}

// Pop the currently active scope of the Compiler's scope stack, and return
// the popped scope's instructions,
func (c *Compiler) leaveScope() code.Instructions {
	ins := c.currentInstructions()

	c.scopes = c.scopes[:len(c.scopes)-1]
	c.scopeIndex--

	c.symbolTable = c.symbolTable.outer

	return ins
}

// Emit the correct get Opcode to retrieve the value associated with the
// provided Symbol.
func (c *Compiler) getSymbol(sym Symbol) {
	switch sym.Scope {
	case GlobalScope:
		c.emit(code.OpGetGlobal, sym.Index)
	case LocalScope:
		c.emit(code.OpGetLocal, sym.Index)
	case BuiltinScope:
		c.emit(code.OpGetBuiltin, sym.Index)
	case FreeScope:
		c.emit(code.OpGetFree, sym.Index)
	case FunctionScope:
		// Emit an instruction specifically for adding the currently executing
		// Closure on to the stack again, for recursive Closure calls.
		c.emit(code.OpCurrentClosure)
	}
}
