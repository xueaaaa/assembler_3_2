package vm

import (
	"fmt"

	"assembler/src/assembler"
)

const minimumValue = 0

// instructionByte читает байт команды из той же памяти, что и данные.
func (machine *Machine) instructionByte(address uint64) (byte, error) {
	value, err := machine.Read(address)
	if err != nil {
		return byte(minimumValue), err
	}
	if uint64(byte(value)) != value {
		return byte(minimumValue), fmt.Errorf("ячейка кода %d не содержит байт: %d", address, value)
	}
	return byte(value), nil
}

// fetch читает и декодирует текущую команду из общей памяти.
func (machine *Machine) fetch() (assembler.Instruction, uint64, error) {
	firstByte, err := machine.instructionByte(machine.counter)
	if err != nil {
		return assembler.Instruction{}, minimumValue, err
	}
	size, err := assembler.InstructionSize(firstByte)
	if err != nil {
		return assembler.Instruction{}, minimumValue, err
	}
	if uint64(size) > machine.programSize-machine.counter {
		return assembler.Instruction{}, minimumValue, fmt.Errorf("оборванная команда длиной %d байт", size)
	}
	data := make([]byte, size)
	for index := range data {
		data[index], err = machine.instructionByte(machine.counter + uint64(index))
		if err != nil {
			return assembler.Instruction{}, minimumValue, err
		}
	}
	instruction, err := assembler.DecodeInstruction(data)
	return instruction, uint64(size), err
}

// Execute выполняет команды до конца исходного двоичного файла.
func (machine *Machine) Execute() error {
	for machine.counter < machine.programSize {
		instruction, size, err := machine.fetch()
		if err != nil {
			return fmt.Errorf("адрес команды %d: %w", machine.counter, err)
		}
		if err := machine.executeInstruction(instruction); err != nil {
			return fmt.Errorf("адрес команды %d: %w", machine.counter, err)
		}
		machine.counter += size
	}
	return nil
}

// executeInstruction исполняет базовую команду или сообщает о неподдержанной.
func (machine *Machine) executeInstruction(instruction assembler.Instruction) error {
	switch instruction.Fields[opcodeIndex].Value {
	case assembler.LoadOpcode:
		machine.stack = append(machine.stack, instruction.Fields[operandIndex].Value)
		return nil
	case assembler.ReadOpcode:
		value, err := machine.Read(instruction.Fields[operandIndex].Value)
		if err != nil {
			return err
		}
		machine.stack = append(machine.stack, value)
		return nil
	case assembler.WriteOpcode:
		return machine.writeFromStack()
	case assembler.PowOpcode:
		return fmt.Errorf("команда POW выполняется начиная с этапа 4")
	default:
		return fmt.Errorf("неподдержанная операция")
	}
}

// writeFromStack снимает значение, затем адрес и записывает значение в память.
func (machine *Machine) writeFromStack() error {
	if len(machine.stack) < writeArity {
		return fmt.Errorf("команде WRITE требуется %d элемента стека", writeArity)
	}
	valueIndex := len(machine.stack) - valueStep
	addressIndex := len(machine.stack) - writeArity
	if err := machine.Write(machine.stack[addressIndex], machine.stack[valueIndex]); err != nil {
		return err
	}
	machine.stack = machine.stack[:addressIndex]
	return nil
}
