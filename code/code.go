// The code package contains the means to represent instructions
package code

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// An alias for a byte slice containing Opcode instructions and their operands.
type Instructions []byte

// An alias for the byte that represents an instruction.
type Opcode byte

// Definition of an Opcode
type Definition struct {
	Name          string // Name associated with an Opcode
	OperandWidths []int  // Widths of each operand an Opcode accepts, measured in bytes
}

const (
	// Retrieve the constant value at the given position and place it on top
	// of the stack.
	OpConstant Opcode = iota
	// Remove the top value from the stack.
	OpPop
	// Push the value 'true' on to the top of the stack.
	OpTrue
	// Push the value 'false' on to the top of the stack.
	OpFalse
	// Conditionally move the instruction pointer to the specified instruction
	// index based on the truthiness of the value on top of the stack, removing
	// the value from the stack in the process. Move on to the next instruction
	// without jumping if the top value evaluates as true.
	OpJumpWhenFalse
	// Move the instruction pointer to the specified instruction index.
	OpJump
	// Push the value 'null' on to the top of the stack.
	OpNull
	// Retrieve the value at the given index in the globals slice and place it
	// on top of the stack.
	OpGetGlobal
	// Copy the value on top of the stack into the specified index of the
	// globals slice.
	OpSetGlobal
	// Execute the top function on the stack, using n values above it as the
	// parameters for the function call where n is the provided value.
	OpCall
	// Return the value on top of the stack as the result of the function
	// call.
	OpReturn
	// Retrieve the local value at the specified index and place it on top of
	// the stack.
	OpGetLocal
	// Copy the value on the top of the stack into the specified index of the
	// locals slice. The locals slice is implemented as a space in the stack
	// that is reserved for the local values of the currently executing
	// function.
	OpSetLocal
	// Place an empty list object on top of the stack.
	OpEmptyList
	// Retrieve the builtin function object that is associated with the index
	// and place it on top of the stack.
	OpGetBuiltin
	// Lambda equivalent of OpConstant: retrieve the constant at the specified
	// index. The aditional argument is the number of free variables to capture
	// from the enclosing scope, which are placed on the stack above the
	// closure object.
	OpClosure
	// Retrieve the free variable at the provided index from the free
	// variables slice associated with the closure object.
	OpGetFree
	// Push another instance of the currently executing closure on to the
	// stack.
	OpCurrentClosure
)

// definitions contains a map from an Opcode to its Definition. The Definition
// contains the name of the Opcode, as well as a slice of integers. Each value
// in the slice represents an argument for the Opcode, with the value itself
// being the number of bytes long the argument will be.
var definitions = map[Opcode]*Definition{
	OpConstant:       {"OpConstant", []int{2}},
	OpPop:            {"OpPop", []int{}},
	OpTrue:           {"OpTrue", []int{}},
	OpFalse:          {"OpFalse", []int{}},
	OpJumpWhenFalse:  {"OpJumpWhenFalse", []int{2}},
	OpJump:           {"OpJump", []int{2}},
	OpNull:           {"OpNull", []int{}},
	OpGetGlobal:      {"OpGetGlobal", []int{2}},
	OpSetGlobal:      {"OpSetGlobal", []int{2}},
	OpCall:           {"OpCall", []int{1}},
	OpReturn:         {"OpReturn", []int{}},
	OpGetLocal:       {"OpGetLocal", []int{1}},
	OpSetLocal:       {"OpSetLocal", []int{1}},
	OpEmptyList:      {"OpEmptyList", []int{}},
	OpGetBuiltin:     {"OpGetBuiltin", []int{1}},
	OpClosure:        {"OpClosure", []int{2, 1}},
	OpGetFree:        {"OpGetFree", []int{1}},
	OpCurrentClosure: {"OpCurrentClosure", []int{}},
}

// Make builds an instruction from the provided Opcode and operands, using the
// operand lengths defined in its Definition
func Make(op Opcode, operands ...int) []byte {
	def, ok := definitions[op]

	if !ok {
		// While error handling here would be considered proper, the chance
		// of producing invalid instructions should be close to zero so long as
		// the overall test cases cover each instructions creation.
		return []byte{}
	}

	instructionLen := 1 // Will hold the full length of the instruction.

	// Calculate the full instruction length from its operand sizes.
	for _, w := range def.OperandWidths {
		instructionLen += w
	}

	instruction := make([]byte, instructionLen)

	//convert Opcode to byte and make first byte in instruction.
	instruction[0] = byte(op)

	offset := 1

	// Convert each operand into bytes, with big endian ordering, then append
	// to instruction.
	for i, o := range operands {
		width := def.OperandWidths[i]

		switch width {
		case 2:
			binary.BigEndian.PutUint16(instruction[offset:], uint16(o))
		case 1:
			instruction[offset] = byte(o)
		}

		offset += width
	}

	return instruction
}

// Lookup returns the Definition of the provided Opcode
func Lookup(op byte) (*Definition, error) {
	def, ok := definitions[Opcode(op)]

	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}

	return def, nil
}

// A simple bytecode decoder: convert bytecode instructions from bytes into
// a human readable format.
func (ins Instructions) String() string {
	var out bytes.Buffer

	i := 0

	for i < len(ins) {
		def, err := Lookup(ins[i])

		if err != nil {
			fmt.Fprintf(&out, "ERROR: %s\n", err)
			continue
		}

		operands, read := ReadOperands(def, ins[i+1:])

		fmt.Fprintf(&out, "%04d %s\n", i, ins.fmtInstruction(def, operands))
		i += 1 + read
	}

	return out.String()
}

// ReadOperands uses an Opcode Definition to extract the operands from an
// already encoded instruction and converts them to a human readable format.
// Returns the decoded operands and the byte width they occupied.
func ReadOperands(def *Definition, ins Instructions) ([]int, int) {
	operands := make([]int, len(def.OperandWidths))

	offset := 0

	for i, width := range def.OperandWidths {
		switch width {
		case 2:
			operands[i] = int(ReadUint16(ins[offset:]))
		case 1:
			operands[i] = int(byte(ins[offset]))
		}

		offset += width
	}

	return operands, offset
}

// ReadUint16 reads enough bytes from the provided Instructions to create
// a uint16.
func ReadUint16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}

// Convert an Opcode definition and its operands into a human readable string.
func (ins Instructions) fmtInstruction(def *Definition, operands []int) string {
	operandCount := len(def.OperandWidths)

	if len(operands) != operandCount {
		return fmt.Sprintf("ERROR: operand len %d does not match defined %d\n", len(operands), operandCount)
	}

	switch operandCount {
	case 0:
		return def.Name
	case 1:
		return fmt.Sprintf("%s %d", def.Name, operands[0])
	case 2:
		return fmt.Sprintf("%s %d %d", def.Name, operands[0], operands[1])
	}

	return fmt.Sprintf("ERROR: unhandled operandCount for %s\n", def.Name)
}
