// Package vm исполняет базовые команды стековой УВМ варианта №2.
package vm

import (
	"fmt"

	"assembler/src/assembler"
)

const (
	addressStep  = 1
	valueStep    = 1
	writeArity   = 2
	opcodeIndex  = 0
	operandIndex = 1
)

// Machine хранит общую память команд и данных, стек и счетчик команд.
type Machine struct {
	memory      map[uint64]uint64
	stack       []uint64
	counter     uint64
	programSize uint64
}

// New загружает байты программы с адреса ноль в общую разреженную память.
func New(program []byte) (*Machine, error) {
	if uint64(len(program)) > assembler.MaximumAddress+addressStep {
		return nil, fmt.Errorf("программа не помещается в адресное пространство УВМ")
	}
	machine := &Machine{memory: make(map[uint64]uint64), programSize: uint64(len(program))}
	for address, value := range program {
		machine.memory[uint64(address)] = uint64(value)
	}
	return machine, nil
}

// Read читает ячейку памяти; незаписанная ячейка содержит ноль.
func (machine *Machine) Read(address uint64) (uint64, error) {
	if address > assembler.MaximumAddress {
		return minimumValue, fmt.Errorf("адрес %d вне диапазона 0..%d", address, assembler.MaximumAddress)
	}
	return machine.memory[address], nil
}

// Write записывает ячейку общей памяти, включая область машинного кода.
func (machine *Machine) Write(address, value uint64) error {
	if address > assembler.MaximumAddress {
		return fmt.Errorf("адрес %d вне диапазона 0..%d", address, assembler.MaximumAddress)
	}
	if machine.memory == nil {
		machine.memory = make(map[uint64]uint64)
	}
	machine.memory[address] = value
	return nil
}

// Stack возвращает независимую копию текущего содержимого стека.
func (machine *Machine) Stack() []uint64 {
	return append([]uint64(nil), machine.stack...)
}
