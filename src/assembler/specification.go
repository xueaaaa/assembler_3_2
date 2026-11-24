// Package assembler переводит JSON в машинный код УВМ варианта №2.
package assembler

const (
	loadOpcode  uint64 = 24
	readOpcode  uint64 = 46
	writeOpcode uint64 = 38
	powOpcode   uint64 = 61

	constantWidth = 8
	addressWidth  = 31
	offsetWidth   = 11
	opcodeWidth   = 6
	bitsPerByte   = 8
	operandShift  = opcodeWidth
	resultShift   = opcodeWidth + offsetWidth
	loadSize      = 2
	readSize      = 5
	writeSize     = 1
	powSize       = 4
	decimalBase   = 10
	singleBit     = 1

	operationKey = "op"
	opcodeField  = "A"
	fieldB       = "B"
	fieldC       = "C"

	loadMnemonic  = "LOAD"
	readMnemonic  = "READ"
	writeMnemonic = "WRITE"
	powMnemonic   = "POW"
	valueKey      = "value"
	addressKey    = "address"
	operandKey    = "operand_offset"
	resultKey     = "result_offset"
)

// operandSpec задает имя операнда JSON, поле УВМ, разрядность и сдвиг.
type operandSpec struct {
	key   string
	field string
	width int
	shift int
}

// maximum возвращает наибольшее беззнаковое значение поля.
func (operand operandSpec) maximum() uint64 {
	return (uint64(singleBit) << operand.width) - singleBit
}

// instructionSpec хранит код операции, размер команды и описания операндов.
type instructionSpec struct {
	opcode   uint64
	size     int
	operands []operandSpec
}

// findSpecification возвращает спецификацию команды по мнемонике.
func findSpecification(mnemonic string) (instructionSpec, bool) {
	switch mnemonic {
	case loadMnemonic:
		return instructionSpec{loadOpcode, loadSize, []operandSpec{
			{valueKey, fieldB, constantWidth, operandShift},
		}}, true
	case readMnemonic:
		return instructionSpec{readOpcode, readSize, []operandSpec{
			{addressKey, fieldB, addressWidth, operandShift},
		}}, true
	case writeMnemonic:
		return instructionSpec{opcode: writeOpcode, size: writeSize}, true
	case powMnemonic:
		return instructionSpec{powOpcode, powSize, []operandSpec{
			{operandKey, fieldB, offsetWidth, operandShift},
			{resultKey, fieldC, offsetWidth, resultShift},
		}}, true
	default:
		return instructionSpec{}, false
	}
}

// findOpcodeSpecification находит описание команды по коду операции.
func findOpcodeSpecification(opcode uint64) (instructionSpec, bool) {
	switch opcode {
	case loadOpcode:
		return findSpecification(loadMnemonic)
	case readOpcode:
		return findSpecification(readMnemonic)
	case writeOpcode:
		return findSpecification(writeMnemonic)
	case powOpcode:
		return findSpecification(powMnemonic)
	default:
		return instructionSpec{}, false
	}
}
