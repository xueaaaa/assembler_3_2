package assembler

import "fmt"

const (
	opcodeMask         = (singleBit << opcodeWidth) - singleBit
	zeroReservedBits   = 0
	unknownCommandSize = 0
)

// InstructionSize определяет полную длину команды по ее первому байту.
func InstructionSize(firstByte byte) (int, error) {
	opcode := uint64(firstByte) & opcodeMask
	specification, found := findOpcodeSpecification(opcode)
	if !found {
		return unknownCommandSize, fmt.Errorf("неизвестный код операции: %d", opcode)
	}
	return specification.size, nil
}

// DecodeInstruction переводит одну полную двоичную команду в поля УВМ.
func DecodeInstruction(data []byte) (Instruction, error) {
	if len(data) < writeSize {
		return Instruction{}, fmt.Errorf("пустая двоичная команда")
	}
	opcode := uint64(data[opcodePosition]) & opcodeMask
	specification, found := findOpcodeSpecification(opcode)
	if !found {
		return Instruction{}, fmt.Errorf("неизвестный код операции: %d", opcode)
	}
	if len(data) != specification.size {
		return Instruction{}, fmt.Errorf("для команды %d требуется %d байт", opcode, specification.size)
	}
	var word uint64
	for index, value := range data {
		word |= uint64(value) << (index * bitsPerByte)
	}
	fields := []Field{{Name: opcodeField, Value: opcode}}
	mask := uint64(opcodeMask)
	for _, operand := range specification.operands {
		value := (word >> operand.shift) & operand.maximum()
		fields = append(fields, Field{Name: operand.field, Value: value})
		mask |= operand.maximum() << operand.shift
	}
	if word & ^mask != zeroReservedBits {
		return Instruction{}, fmt.Errorf("резервные биты команды должны быть нулевыми")
	}
	return Instruction{Fields: fields}, nil
}
