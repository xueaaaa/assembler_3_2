package assembler

import (
	"fmt"
	"strings"
)

const opcodePosition = 0

// encodingSpecification проверяет заголовок и число промежуточных полей.
func encodingSpecification(instruction Instruction) (instructionSpec, error) {
	if len(instruction.Fields) < opcodeFieldCount {
		return instructionSpec{}, fmt.Errorf("отсутствует поле %s", opcodeField)
	}
	opcode := instruction.Fields[opcodePosition]
	if opcode.Name != opcodeField {
		return instructionSpec{}, fmt.Errorf("первым должно быть поле %s", opcodeField)
	}
	specification, found := findOpcodeSpecification(opcode.Value)
	if !found {
		return instructionSpec{}, fmt.Errorf("неизвестный код операции: %d", opcode.Value)
	}
	if len(instruction.Fields) != len(specification.operands)+opcodeFieldCount {
		return instructionSpec{}, fmt.Errorf("неверное число промежуточных полей")
	}
	return specification, nil
}

// EncodeInstruction упаковывает поля одной команды в байты little-endian.
func EncodeInstruction(instruction Instruction) ([]byte, error) {
	specification, err := encodingSpecification(instruction)
	if err != nil {
		return nil, err
	}
	word := specification.opcode
	for index, operand := range specification.operands {
		field := instruction.Fields[index+opcodeFieldCount]
		if field.Name != operand.field || field.Value > operand.maximum() {
			return nil, fmt.Errorf("недопустимое поле %s: %s=%d", operand.field, field.Name, field.Value)
		}
		word |= field.Value << operand.shift
	}
	data := make([]byte, specification.size)
	for index := range data {
		data[index] = byte(word >> (index * bitsPerByte))
	}
	return data, nil
}

// Encode переводит программу в массив двоичных команд, сохраняя их границы.
func Encode(instructions []Instruction) ([][]byte, error) {
	commands := make([][]byte, len(instructions))
	for index, instruction := range instructions {
		command, err := EncodeInstruction(instruction)
		if err != nil {
			return nil, fmt.Errorf("команда %d: %w", index+firstInstructionNumber, err)
		}
		commands[index] = command
	}
	return commands, nil
}

// FormatBytes представляет команду в байтовом формате примеров задания.
func FormatBytes(command []byte) string {
	parts := make([]string, len(command))
	for index, value := range command {
		parts[index] = fmt.Sprintf("0x%02X", value)
	}
	return strings.Join(parts, fieldSeparator)
}
