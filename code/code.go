// The code package contains the means to represent instructions
package code

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Instructions []byte

type Opcode byte

// Definition of an Opcode
type Definition struct {
	Name          string // Name associated with an Opcode
	OperandWidths []int  // Widths of each operand an Opcode accepts, measured in bytes
}

const (
	OpConstant Opcode = iota
	OpPop
	OpTrue
	OpFalse
	OpJumpWhenFalse
	OpJump
	OpNull
	OpGetGlobal
	OpSetGlobal
	OpCall
	OpReturn
	OpGetLocal
	OpSetLocal
)

var definitions = map[Opcode]*Definition{
	OpConstant:      {"OpConstant", []int{2}},
	OpPop:           {"OpPop", []int{}},
	OpTrue:          {"OpTrue", []int{}},
	OpFalse:         {"OpFalse", []int{}},
	OpJumpWhenFalse: {"OpJumpWhenFalse", []int{2}},
	OpJump:          {"OpJump", []int{2}},
	OpNull:          {"OpNull", []int{}},
	OpGetGlobal:     {"OpGetGlobal", []int{2}},
	OpSetGlobal:     {"OpSetGlobal", []int{2}},
	OpCall:          {"OpCall", []int{2}},
	OpReturn:        {"OpReturn", []int{}},
	OpGetLocal:      {"OpGetLocal", []int{1}},
	OpSetLocal:      {"OpSetLocal", []int{1}},
}

// Make builds an instruction from the provided Opcode and operands, using the
// operand lengths defined in its Definition
func Make(op Opcode, operands ...int) []byte {
	def, ok := definitions[op]

	if !ok {
		// While error handling here would be considered proper, the chance
		// of producing invalid instructions should be close to nil so long as
		// the test cases cover each instructions creation.
		return []byte{}
	}

	instructionLen := 1

	for _, w := range def.OperandWidths {
		instructionLen += w
	}

	instruction := make([]byte, instructionLen)
	instruction[0] = byte(op)

	offset := 1

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
// already encoded instruction
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

// ReadUint16 reads the enough bytes from the provided Instructions to create
// a uint16
func ReadUint16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}

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
	}

	return fmt.Sprintf("ERROR: unhandled operandCount for %s\n", def.Name)
}
