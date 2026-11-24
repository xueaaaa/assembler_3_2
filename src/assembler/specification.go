// Package assembler переводит JSON в машинный код УВМ варианта №2.
package assembler

const (
	// LoadOpcode обозначает загрузку константы на стек.
	LoadOpcode uint64 = 24
	// ReadOpcode обозначает чтение памяти на стек.
	ReadOpcode uint64 = 46
	// WriteOpcode обозначает запись значения со стека в память.
	WriteOpcode uint64 = 38
	// PowOpcode обозначает возведение в степень на этапе 4.
	PowOpcode uint64 = 61

	constantWidth = 8
	// AddressWidth задает разрядность адреса памяти УВМ.
	AddressWidth = 31
	offsetWidth  = 11
	opcodeWidth  = 6
	bitsPerByte  = 8
	operandShift = opcodeWidth
	resultShift  = opcodeWidth + offsetWidth
	loadSize     = 2
	readSize     = 5
	writeSize    = 1
	powSize      = 4
	decimalBase  = 10
	singleBit    = 1

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

// MaximumAddress задает наибольший адрес общей памяти команд и данных.
const MaximumAddress uint64 = (singleBit << AddressWidth) - singleBit

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
		return instructionSpec{LoadOpcode, loadSize, []operandSpec{
			{valueKey, fieldB, constantWidth, operandShift},
		}}, true
	case readMnemonic:
		return instructionSpec{ReadOpcode, readSize, []operandSpec{
			{addressKey, fieldB, AddressWidth, operandShift},
		}}, true
	case writeMnemonic:
		return instructionSpec{opcode: WriteOpcode, size: writeSize}, true
	case powMnemonic:
		return instructionSpec{PowOpcode, powSize, []operandSpec{
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
	case LoadOpcode:
		return findSpecification(loadMnemonic)
	case ReadOpcode:
		return findSpecification(readMnemonic)
	case WriteOpcode:
		return findSpecification(writeMnemonic)
	case PowOpcode:
		return findSpecification(powMnemonic)
	default:
		return instructionSpec{}, false
	}
}
