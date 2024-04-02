package vm

import (
	"fmt"
	"lisp/code"
	"lisp/compiler"
	"lisp/object"
)

const StackSize = 2048
const GlobalSize = 65536

const MaxFrames = 1024

var True = object.TRUE
var False = object.FALSE
var Null = object.NULL

// VM is used to execute the bytecode it contains.
type VM struct {
	constants   []object.Object // slice of constant values that are referenced in the bytecode instructions
	stack       []object.Object // the active stack used during execution
	sp          int             // pointer next open space on the stack
	globals     []object.Object // stack of global objects in the current program
	frames      []*Frame        // stack of frames for function execution
	framesIndex int             // pointer to the next open place on the frames stack
}

// Create a new VM instance from the provided bytecode.
func New(bytecode *compiler.Bytecode) *VM {
	mainFn := &object.CompiledLambda{
		Instructions: bytecode.Instructions,
	}

	mainFrame := NewFrame(mainFn, 0)

	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	return &VM{
		constants:   bytecode.Constants,
		stack:       make([]object.Object, StackSize),
		sp:          0,
		globals:     make([]object.Object, GlobalSize),
		frames:      frames,
		framesIndex: 1,
	}
}

// Create a new VM instance from the provided bytecode, along with predefined
// globals so that state can be maintained between VM instances.
func NewWithState(bytecode *compiler.Bytecode, globals []object.Object) *VM {
	mainFn := &object.CompiledLambda{
		Instructions: bytecode.Instructions,
	}

	mainFrame := NewFrame(mainFn, 0)

	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	return &VM{
		constants:   bytecode.Constants,
		stack:       make([]object.Object, StackSize),
		sp:          0,
		globals:     globals,
		frames:      frames,
		framesIndex: 1,
	}
}

// Execute the bytecode instructions, using a fetch, decode, execute cycle.
//
// Returns an error if something in execution fails.
func (vm *VM) Run() error {
	var ip int
	var ins code.Instructions
	var op code.Opcode

	// fetch
	for vm.currentFrame().ip < len(vm.currentFrame().Instructions())-1 {
		vm.currentFrame().ip++
		ip = vm.currentFrame().ip
		ins = vm.currentFrame().Instructions()
		op = code.Opcode(ins[ip])

		// decode
		switch op {
		case code.OpConstant:
			// execute
			constIndex := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2

			err := vm.push(vm.constants[constIndex])

			if err != nil {
				return err
			}
		case code.OpPop:
			vm.pop()
		case code.OpTrue:
			err := vm.push(True)

			if err != nil {
				return err
			}
		case code.OpFalse:
			err := vm.push(False)

			if err != nil {
				return err
			}
		case code.OpJump:
			pos := int(code.ReadUint16(ins[ip+1:]))

			// decrement jump position so that we arrive at the target position
			// when the cycle increments the instruction pointer
			vm.currentFrame().ip = pos - 1
		case code.OpJumpWhenFalse:
			pos := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			condition := vm.pop()

			if !isTruthy(condition) {
				vm.currentFrame().ip = pos - 1
			}
		case code.OpNull:
			err := vm.push(Null)

			if err != nil {
				return err
			}
		case code.OpSetGlobal:
			index := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2

			vm.globals[index] = vm.stack[vm.sp-1]
		case code.OpGetGlobal:
			index := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2

			err := vm.push(vm.globals[index])

			if err != nil {
				return err
			}
		case code.OpSetLocal:
			index := int(ins[ip+1])
			vm.currentFrame().ip += 1

			basePointer := vm.currentFrame().basePointer

			vm.stack[basePointer+index] = vm.stack[vm.sp-1]
		case code.OpGetLocal:
			index := int(ins[ip+1])
			vm.currentFrame().ip += 1

			basePointer := vm.currentFrame().basePointer

			err := vm.push(vm.stack[basePointer+index])

			if err != nil {
				return err
			}
		case code.OpCall:
			argCount := int(ins[ip+1])
			vm.currentFrame().ip += 1

			// Look for the lambda before the arguments that have been pushed
			// onto the stack above it.
			// Extra -1 is because vm.sp points to the space after the top of
			// the stack.
			lambda, ok := vm.stack[vm.sp-argCount-1].(*object.CompiledLambda)

			if !ok {
				return fmt.Errorf("calling non-function")
			}

			if argCount != lambda.ParameterCount {
				return fmt.Errorf(
					"wrong number of arguments: expected=%d got=%d",
					lambda.ParameterCount, argCount,
				)
			}

			frame := NewFrame(lambda, vm.sp-argCount)

			vm.pushFrame(frame)
			// Reserve space on the stack for local bindings:
			//
			// The space between frame.basePointer (the current stack pointer)
			// and fn.LocalsCount reserves fn.LocalsCount number of spaces for
			// paramaters and local bindings, since parameters are a special
			// case of local bindings. This allows the stack beyond this point
			// to be used as normal in instruction execution.
			vm.sp = frame.basePointer + lambda.LocalsCount
		case code.OpReturn:
			returnValue := vm.pop()

			frame := vm.popFrame()
			vm.sp = frame.basePointer - 1

			err := vm.push(returnValue)

			if err != nil {
				return err
			}
		case code.OpEmptyList:
			err := vm.push(&object.List{})

			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Return the item currently at the top of the stack.
func (vm *VM) StackTop() object.Object {
	if vm.sp == 0 {
		return nil
	}

	return vm.stack[vm.sp-1]
}

// Add an object onto the stack, return an error if the stack is full.
func (vm *VM) push(o object.Object) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}

	vm.stack[vm.sp] = o
	vm.sp++

	return nil
}

// Return the item from the top of the stack and decrement the stack pointer.
func (vm *VM) pop() object.Object {
	o := vm.stack[vm.sp-1]

	vm.sp--

	return o
}

// Return the item that was last popped from the stack.
func (vm *VM) LastPoppedStackElem() object.Object {
	return vm.stack[vm.sp]
}

// Evaluate the Object as true or false.
func isTruthy(o object.Object) bool {
	return o != False && o != Null
}

func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.framesIndex-1]
}

func (vm *VM) pushFrame(f *Frame) {
	vm.frames[vm.framesIndex] = f
	vm.framesIndex++
}

func (vm *VM) popFrame() *Frame {
	vm.framesIndex--
	return vm.frames[vm.framesIndex]
}
