// Package assembler реализует первый этап ассемблера УВМ варианта №2.
package assembler

const (
	loadOpcode  uint64 = 24
	readOpcode  uint64 = 46
	writeOpcode uint64 = 38
	powOpcode   uint64 = 61

	constantWidth = 8
	addressWidth  = 31
	offsetWidth   = 11
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

// operandSpec задает имя операнда JSON, поле УВМ и разрядность.
type operandSpec struct {
	key   string
	field string
	width int
}

// maximum возвращает наибольшее беззнаковое значение поля.
func (operand operandSpec) maximum() uint64 {
	return (uint64(singleBit) << operand.width) - singleBit
}

// instructionSpec хранит код операции и описания ее операндов.
type instructionSpec struct {
	opcode   uint64
	operands []operandSpec
}

// findSpecification возвращает спецификацию команды по мнемонике.
func findSpecification(mnemonic string) (instructionSpec, bool) {
	switch mnemonic {
	case loadMnemonic:
		return instructionSpec{loadOpcode, []operandSpec{
			{valueKey, fieldB, constantWidth},
		}}, true
	case readMnemonic:
		return instructionSpec{readOpcode, []operandSpec{
			{addressKey, fieldB, addressWidth},
		}}, true
	case writeMnemonic:
		return instructionSpec{opcode: writeOpcode}, true
	case powMnemonic:
		return instructionSpec{powOpcode, []operandSpec{
			{operandKey, fieldB, offsetWidth},
			{resultKey, fieldC, offsetWidth},
		}}, true
	default:
		return instructionSpec{}, false
	}
}
