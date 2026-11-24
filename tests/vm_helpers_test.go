package assembler_test

import (
	"bytes"
	"testing"

	"assembler/src/assembler"
	"assembler/src/vm"
)

const (
	copyFilename         = "copy.json"
	copyExpectedFilename = "expected-copy.csv"
	copyStart            = 64
	copyEnd              = 82
	copyDestination      = 80
	copyFirstValue       = 11
	copySecondValue      = 22
	copyThirdValue       = 33
	memoryZero           = 0
	stackFirstIndex      = 0
	modifiedConstant     = 2
	modifiedByte         = 152
	modifiedAddress      = 5
	nonByteValue         = maximumConstant + outsideRangeStep
	copyStartArgument    = "64"
	copyEndArgument      = "82"
)

// assembleProgram переводит тестовый JSON в непрерывный машинный код.
func assembleProgram(t *testing.T, source []byte) []byte {
	t.Helper()
	instructions, err := assembler.Translate(source)
	if err != nil {
		t.Fatal(err)
	}
	commands, err := assembler.Encode(instructions)
	if err != nil {
		t.Fatal(err)
	}
	return bytes.Join(commands, nil)
}

// newTestMachine создает машину из тестовой программы JSON.
func newTestMachine(t *testing.T, source string) *vm.Machine {
	t.Helper()
	machine, err := vm.New(assembleProgram(t, []byte(source)))
	if err != nil {
		t.Fatal(err)
	}
	return machine
}

// executeMachine запускает машину и завершает тест при ошибке исполнения.
func executeMachine(t *testing.T, machine *vm.Machine) {
	t.Helper()
	if err := machine.Execute(); err != nil {
		t.Fatal(err)
	}
}

// assertCell проверяет значение конкретной ячейки общей памяти.
func assertCell(t *testing.T, machine *vm.Machine, address, expected uint64) {
	t.Helper()
	actual, err := machine.Read(address)
	if err != nil || actual != expected {
		t.Fatalf("память[%d]: %d вместо %d, ошибка %v", address, actual, expected, err)
	}
}
