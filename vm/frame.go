package vm

import (
	"lisp/code"
	"lisp/object"
)

// Frame holds the current VM execution state.
type Frame struct {
	// The currently executing Closure.
	Closure *object.Closure
	// Instruction pointer for this frame.
	ip int
	// Pointer to the base of the stack before lambda execution.
	// Used for interacting with local bindings, which are stored
	// on top of the stack.
	basePointer int
}

// Create a new VM with the provided compiled lambda.
func NewFrame(closure *object.Closure, basePointer int) *Frame {
	return &Frame{
		Closure:     closure,
		ip:          -1, // so that ip == 0 after increment
		basePointer: basePointer,
	}
}

// Return the instructions of the Closure associated with the current Frame.
func (f *Frame) Instructions() code.Instructions {
	return f.Closure.Lambda.Instructions
}
