package vm

import (
	"fmt"
	"lisp/code"
	"lisp/compiler"
	"lisp/object"
)

const StackSize = 2048
const GlobalSize = 65536

var True = &object.BooleanObject{Value: true}
var False = &object.BooleanObject{Value: false}
var Null = &object.Null{}

// VM is used to execute the bytecode it contains.
type VM struct {
	constants    []object.Object   // slice of constant values that are referenced in the bytecode instructions
	instructions code.Instructions // bytecode instructions to be executed
	stack        []object.Object   // the active stack used during execution
	sp           int               // pointer next open space on the stack
	globals      []object.Object
}

// Create a new VM instance from the provided bytecode.
func New(bytecode *compiler.Bytecode) *VM {
	return &VM{
		instructions: bytecode.Instructions,
		constants:    bytecode.Constants,
		stack:        make([]object.Object, StackSize),
		sp:           0,
		globals:      make([]object.Object, GlobalSize),
	}
}

// Create a new VM instance from the provided bytecode, along with predefined
// globals so that state can be maintained between VM instances.
func NewWithState(bytecode *compiler.Bytecode, globals []object.Object) *VM {
	return &VM{
		instructions: bytecode.Instructions,
		constants:    bytecode.Constants,
		stack:        make([]object.Object, StackSize),
		sp:           0,
		globals:      globals,
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
			pos := int(code.ReadUint16(vm.instructions[ip+1:]))

			// decrement jump position so that we arrive at the target position
			// when the cycle increments the instruction pointer
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
		case code.OpSetGlobal:
			index := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2

			vm.globals[index] = vm.pop()
		case code.OpGetGlobal:
			index := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2

			err := vm.push(vm.globals[index])

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
