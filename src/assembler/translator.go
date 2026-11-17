package assembler

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	opcodeFieldCount = 1
	fieldSeparator   = ", "
)

// Field содержит имя поля УВМ и его беззнаковое значение.
type Field struct {
	Name  string
	Value uint64
}

// Instruction хранит поля команды в порядке A, B, C.
type Instruction struct {
	Fields []Field
}

// String форматирует поля по образцу тестов спецификации УВМ.
func (instruction Instruction) String() string {
	parts := make([]string, len(instruction.Fields))
	for index, field := range instruction.Fields {
		parts[index] = fmt.Sprintf("%s=%d", field.Name, field.Value)
	}
	return "(" + strings.Join(parts, fieldSeparator) + ")"
}

// Translate проверяет JSON в UTF-8 и формирует промежуточные команды.
func Translate(data []byte) ([]Instruction, error) {
	if !utf8.Valid(data) {
		return nil, fmt.Errorf("исходный файл должен быть в кодировке UTF-8")
	}
	commands, err := readCommands(data)
	if err != nil {
		return nil, err
	}
	instructions := make([]Instruction, len(commands))
	for index, command := range commands {
		instruction, err := translateCommand(command)
		if err != nil {
			return nil, fmt.Errorf("команда %d: %w", index+firstInstructionNumber, err)
		}
		instructions[index] = instruction
	}
	return instructions, nil
}

// commandSpecification проверяет мнемонику и набор полей объекта.
func commandSpecification(command map[string]json.RawMessage) (instructionSpec, error) {
	var mnemonic string
	if err := json.Unmarshal(command[operationKey], &mnemonic); err != nil {
		return instructionSpec{}, fmt.Errorf("поле %s должно быть строкой", operationKey)
	}
	specification, found := findSpecification(mnemonic)
	if !found {
		return instructionSpec{}, fmt.Errorf("неизвестная операция: %q", mnemonic)
	}
	if len(command) != len(specification.operands)+opcodeFieldCount {
		return instructionSpec{}, fmt.Errorf("неверный набор полей команды %s", mnemonic)
	}
	for _, operand := range specification.operands {
		if _, exists := command[operand.key]; !exists {
			return instructionSpec{}, fmt.Errorf("отсутствует поле %s", operand.key)
		}
	}
	return specification, nil
}

// translateCommand переводит объект JSON в упорядоченные поля УВМ.
func translateCommand(command map[string]json.RawMessage) (Instruction, error) {
	specification, err := commandSpecification(command)
	if err != nil {
		return Instruction{}, err
	}
	fields := []Field{{Name: opcodeField, Value: specification.opcode}}
	for _, operand := range specification.operands {
		value, err := strconv.ParseUint(string(command[operand.key]), decimalBase, operand.width)
		if err != nil {
			return Instruction{}, fmt.Errorf(
				"поле %s должно быть целым числом от 0 до %d", operand.key, operand.maximum(),
			)
		}
		fields = append(fields, Field{Name: operand.field, Value: value})
	}
	return Instruction{Fields: fields}, nil
}
