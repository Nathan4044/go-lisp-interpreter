package vm

import (
	"fmt"
	"lisp/code"
	"lisp/compiler"
	"lisp/object"
)

const StackSize = 2048

var True = &object.BooleanObject{Value: true}
var False = &object.BooleanObject{Value: false}
var Null = &object.Null{}

// VM is used to execute the bytecode it contains.
type VM struct {
	constants    []object.Object   // slice of constant values that are referenced in the bytecode instructions
	instructions code.Instructions // bytecode instructions to be executed
	stack        []object.Object   // the active stack used during execution
	sp           int               // pointer next open space on the stack
}

// Create a new VM instance from the provided bytecode.
func New(bytecode *compiler.Bytecode) *VM {
	return &VM{
		instructions: bytecode.Instructions,
		constants:    bytecode.Constants,
		stack:        make([]object.Object, StackSize),
		sp:           0,
	}
}

// Execute the bytecode instructions, using a fetch, decode, execute cycle.
//
// Returns an error if something in execution fails.
func (vm *VM) Run() error {
	// fetch
	for ip := 0; ip < len(vm.instructions); ip++ {
		op := code.Opcode(vm.instructions[ip])

		// decode
		switch op {
		case code.OpConstant:
			// execute
			constIndex := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2

			err := vm.push(vm.constants[constIndex])

			if err != nil {
				return err
			}
		case code.OpPop:
			vm.pop()
		case code.OpTrue:
			vm.push(True)
		case code.OpFalse:
			vm.push(False)
		case code.OpJump:
			pos := int(code.ReadUint16(vm.instructions[ip+1:]))

			// decrement position so that it is correct when loop
			// increments it
			ip = pos - 1
		case code.OpJumpWhenFalse:
			pos := int(code.ReadUint16(vm.instructions[ip+1:]))
			ip += 2

			condition := vm.pop()

			if !isTruthy(condition) {
				ip = pos - 1
			}
		case code.OpNull:
			err := vm.push(Null)

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

func (vm *VM) push(o object.Object) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}

	vm.stack[vm.sp] = o
	vm.sp++

	return nil
}

func (vm *VM) pop() object.Object {
	o := vm.stack[vm.sp-1]

	vm.sp--

	return o
}

// Return the item that was last popped from the stack.
func (vm *VM) LastPoppedStackElem() object.Object {
	return vm.stack[vm.sp]
}

func isTruthy(o object.Object) bool {
	return o != False && o != Null
}
