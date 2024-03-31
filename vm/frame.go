package vm

import (
	"lisp/code"
	"lisp/object"
)

// Frame holds the current VM execution state.
type Frame struct {
	Lambda *object.CompiledLambda
	ip     int // instruction pointer for this frame
}

// Create a new VM with the provided compiled lambda.
func NewFrame(lambda *object.CompiledLambda) *Frame {
	return &Frame{
		Lambda: lambda,
		ip:     -1, // so that ip == 0 after increment
	}
}

func (f *Frame) Instructions() code.Instructions {
	return f.Lambda.Instructions
}
