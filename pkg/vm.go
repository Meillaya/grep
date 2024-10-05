package pkg

import (
	"fmt"
	"sync"
)

// VM represents a virtual machine
type VM struct {
	memory  []byte
	pc      int
	stack   []int
	mutex   sync.Mutex
	running bool
}

// NewVM creates a new virtual machine with the given memory size
func NewVM(memorySize int) *VM {
	return &VM{
		memory: make([]byte, memorySize),
		pc:     0,
		stack:  make([]int, 0),
	}
}

// Run starts the execution of the virtual machine
func (vm *VM) Run() error {
	vm.mutex.Lock()
	defer vm.mutex.Unlock()

	if vm.running {
		return fmt.Errorf("VM is already running")
	}

	vm.running = true
	defer func() { vm.running = false }()

	for vm.pc < len(vm.memory) {
		opcode := vm.memory[vm.pc]
		vm.pc++

		switch opcode {
		case 0x00: // NOP
			// Do nothing
		case 0x01: // PUSH
			if vm.pc >= len(vm.memory) {
				return fmt.Errorf("unexpected end of memory")
			}
			value := int(vm.memory[vm.pc])
			vm.pc++
			vm.stack = append(vm.stack, value)
		case 0x02: // POP
			if len(vm.stack) == 0 {
				return fmt.Errorf("stack underflow")
			}
			vm.stack = vm.stack[:len(vm.stack)-1]
		case 0x03: // ADD
			if len(vm.stack) < 2 {
				return fmt.Errorf("not enough operands for ADD")
			}
			a := vm.stack[len(vm.stack)-2]
			b := vm.stack[len(vm.stack)-1]
			vm.stack = vm.stack[:len(vm.stack)-2]
			vm.stack = append(vm.stack, a+b)
		case 0x04: // SUB
			if len(vm.stack) < 2 {
				return fmt.Errorf("not enough operands for SUB")
			}
			a := vm.stack[len(vm.stack)-2]
			b := vm.stack[len(vm.stack)-1]
			vm.stack = vm.stack[:len(vm.stack)-2]
			vm.stack = append(vm.stack, a-b)
		case 0x05: // HALT
			return nil
		default:
			return fmt.Errorf("unknown opcode: 0x%02x", opcode)
		}
	}

	return nil
}

// LoadProgram loads a program into the VM's memory
func (vm *VM) LoadProgram(program []byte) error {
	vm.mutex.Lock()
	defer vm.mutex.Unlock()

	if len(program) > len(vm.memory) {
		return fmt.Errorf("program size exceeds memory size")
	}

	copy(vm.memory, program)
	return nil
}

// GetStackTop returns the top value of the stack without removing it
func (vm *VM) GetStackTop() (int, error) {
	vm.mutex.Lock()
	defer vm.mutex.Unlock()

	if len(vm.stack) == 0 {
		return 0, fmt.Errorf("stack is empty")
	}

	return vm.stack[len(vm.stack)-1], nil
}
