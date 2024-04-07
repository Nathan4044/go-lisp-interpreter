// The vm package contains the definition of the VM type, which executes
// bytecode instructions created by a Compiler.
package vm

import (
	"fmt"
	"lisp/code"
	"lisp/compiler"
	"lisp/object"
)

const (
	StackSize  = 2048
	GlobalSize = 65536
	MaxFrames  = 1024
)

// Global references to true, false, and null resolve to a single object for
// for each value.
var True = object.TRUE
var False = object.FALSE
var Null = object.NULL

// VM is used to execute the bytecode it contains.
type VM struct {
	// Slice of constant values that are referenced in the bytecode instructions
	constants []object.Object
	// The active stack used during execution
	stack []object.Object
	// Pointer next open space on the stack
	sp int
	// Stack of global objects in the current program
	globals []object.Object
	// Stack of frames for function execution
	frames []*Frame
	// Pointer to the next open place on the frames stack
	framesIndex int
}

// Create a new VM instance from the provided bytecode.
func New(bytecode *compiler.Bytecode) *VM {
	// Represent the entire program as a Closure so that each level of
	// execution operate the same.
	mainLambda := &object.CompiledLambda{
		Instructions: bytecode.Instructions,
	}
	mainClosure := &object.Closure{Lambda: mainLambda}

	mainFrame := NewFrame(mainClosure, 0)

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
	vm := New(bytecode)
	vm.globals = globals

	return vm
}

// Execute the bytecode instructions, using a fetch, decode, execute cycle.
//
// The top Frame of the VM represents the currently executing state of the VM,
// including executing instructions in the form of a Closure, the instruction
// pointer, and the pointer to where the current Frame execution began.
//
// Returns an error if something in execution fails.
func (vm *VM) Run() error {
	var ip int
	var ins code.Instructions
	var op code.Opcode

	// Fetch
	for vm.currentFrame().ip < len(vm.currentFrame().Instructions())-1 {
		vm.currentFrame().ip++
		ip = vm.currentFrame().ip
		ins = vm.currentFrame().Instructions()
		op = code.Opcode(ins[ip])

		// Decode
		switch op {
		case code.OpConstant:
			// Execute

			// Place the references constant on top of the stack.
			constIndex := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2

			err := vm.push(vm.constants[constIndex])

			if err != nil {
				return err
			}
		case code.OpPop:
			// Remove the top item from the stack.
			vm.pop()
		case code.OpTrue:
			// Place the value of 'true' of top of the stack.
			err := vm.push(True)

			if err != nil {
				return err
			}
		case code.OpFalse:
			// Place the value of 'false' of top of the stack.
			err := vm.push(False)

			if err != nil {
				return err
			}
		case code.OpJump:
			// Move the instruction pointer to the position provided.
			pos := int(code.ReadUint16(ins[ip+1:]))

			// Decrement the new position so that we arrive at the target
			// position when the cycle increments the instruction pointer.
			vm.currentFrame().ip = pos - 1
		case code.OpJumpWhenFalse:
			// Take the object on top of the stack and evaluate its truthiness,
			// jump to the provided instruction position if the evaluation is
			// false.
			pos := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			condition := vm.pop()

			if !isTruthy(condition) {
				// Decrement the new position so that we arrive at the target
				// position when the cycle increments the instruction pointer.
				vm.currentFrame().ip = pos - 1
			}
		case code.OpNull:
			// Place the value of 'null' of top of the stack.
			err := vm.push(Null)

			if err != nil {
				return err
			}
		case code.OpSetGlobal:
			// Set the value of the global at the provided index to the object
			// on top of the stack without removing the object from the stack.
			index := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2

			vm.globals[index] = vm.stack[vm.sp-1]
		case code.OpGetGlobal:
			// Place the requested global value onto the top of the stack.
			index := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2

			err := vm.push(vm.globals[index])

			if err != nil {
				return err
			}
		case code.OpSetLocal:
			// Set the value of the local at the provided index to the object on
			// top of the stack without removing the object from the stack.
			index := int(ins[ip+1])
			vm.currentFrame().ip += 1

			basePointer := vm.currentFrame().basePointer

			vm.stack[basePointer+index] = vm.stack[vm.sp-1]
		case code.OpGetLocal:
			// Place the requested local value onto the top of the stack.
			index := int(ins[ip+1])
			vm.currentFrame().ip += 1

			basePointer := vm.currentFrame().basePointer

			// Local values are retrieved from the 'hole' in the stack
			// that's reserved for locals, which sits just above the
			// currently executing Closure.
			err := vm.push(vm.stack[basePointer+index])

			if err != nil {
				return err
			}
		case code.OpGetBuiltin:
			// Retrieve the builtin function at the provided index and place it
			// on top of the stack.
			index := int(ins[ip+1])
			vm.currentFrame().ip += 1

			err := vm.push(object.Builtins[index])

			if err != nil {
				return err
			}
		case code.OpCall:
			// Execute the function at the top of the stack, using the arguments
			// placed on top of it.
			argCount := int(ins[ip+1])
			vm.currentFrame().ip += 1

			// Look for the fn before the arguments that have been pushed
			// onto the stack above it.
			// Extra -1 is because vm.sp points to the space after the top of
			// the stack.
			switch fn := vm.stack[vm.sp-argCount-1].(type) {
			case *object.Closure:
				// When executing a Closure, a new frame is created and pushed
				// onto the frame stack, the next loop through Run will use the
				// instructions and values of the new Frame, which will be
				// popped off the frame stack when execution completes.
				if argCount != fn.Lambda.ParameterCount {
					return fmt.Errorf(
						"wrong number of arguments: expected=%d got=%d",
						fn.Lambda.ParameterCount, argCount,
					)
				}

				frame := NewFrame(fn, vm.sp-argCount)

				vm.pushFrame(frame)
				// Reserve space on the stack for local bindings:
				//
				// The space between frame.basePointer (the current stack pointer)
				// and fn.LocalsCount reserves fn.LocalsCount number of spaces for
				// paramaters and local bindings, since parameters are a special
				// case of local bindings. This allows the stack beyond this point
				// to be used as normal in instruction execution.
				vm.sp = frame.basePointer + fn.Lambda.LocalsCount
			case *object.FunctionObject:
				// When executing a builtin function, call the inner function
				// written in go and push the resulting value onto the stack.
				args := vm.stack[vm.sp-argCount : vm.sp]

				result := fn.Fn(args...)

				if result.Type() == object.ERROR_OBJ {
					errObj, _ := result.(*object.ErrorObject)

					return fmt.Errorf("%s", errObj.Error)
				}

				vm.sp = vm.sp - argCount - 1

				err := vm.push(result)

				if err != nil {
					return err
				}
			default:
				return fmt.Errorf("calling non-function")
			}
		case code.OpReturn:
			// Return the value from a function. Pop the current Frame from the
			// frame stack and remove its execution state from the stack, then
			// push the resulting value on to the top of the stack
			returnValue := vm.pop()

			frame := vm.popFrame()
			vm.sp = frame.basePointer - 1

			err := vm.push(returnValue)

			if err != nil {
				return err
			}
		case code.OpEmptyList:
			// Place an empty list object on top of the stack.
			err := vm.push(&object.List{})

			if err != nil {
				return err
			}
		case code.OpClosure:
			// Create a Closure object from the CompiledLambda at the provided
			// index and the free variables from the top of the stack, then
			// place the new Closure on top of the stack.
			index := code.ReadUint16(ins[ip+1:])
			freeCount := int(ins[ip+3])
			vm.currentFrame().ip += 3

			constant := vm.constants[index]
			lambda, ok := constant.(*object.CompiledLambda)

			if !ok {
				return fmt.Errorf("object not lambda: %+v", constant)
			}

			freeVariables := make([]object.Object, freeCount)

			for i := 0; i < freeCount; i++ {
				freeVariables[i] = vm.stack[vm.sp-freeCount+i]
			}

			vm.sp -= freeCount

			err := vm.push(&object.Closure{Lambda: lambda, Free: freeVariables})

			if err != nil {
				return err
			}
		case code.OpGetFree:
			// Retrieve the free variable at the provided index from the
			// free variables associated with the Closure of the current Frame.
			index := int(ins[ip+1])
			vm.currentFrame().ip += 1

			err := vm.push(vm.currentFrame().Closure.Free[index])

			if err != nil {
				return err
			}
		case code.OpCurrentClosure:
			// Place the Closure of the currently executing Frame and place it
			// on top of the stack
			currentClosure := vm.currentFrame().Closure

			err := vm.push(currentClosure)

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
