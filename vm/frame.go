package vm

import (
	"lisp/code"
	"lisp/object"
)

// Frame holds the current VM execution state.
type Frame struct {
	Lambda *object.CompiledLambda
	// Instruction pointer for this frame.
	ip int
	// Pointer to the base of the stack before lambda execution.
	// Used for interacting with local bindings, which are stored
	// on top of the stack.
	basePointer int
}

// Create a new VM with the provided compiled lambda.
func NewFrame(lambda *object.CompiledLambda, basePointer int) *Frame {
	return &Frame{
		Lambda:      lambda,
		ip:          -1, // so that ip == 0 after increment
		basePointer: basePointer,
	}
}

func (f *Frame) Instructions() code.Instructions {
	return f.Lambda.Instructions
}
