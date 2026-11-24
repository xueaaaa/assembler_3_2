package assembler_test

import (
	"reflect"
	"strings"
	"testing"

	"assembler/src/assembler"
	"assembler/src/vm"
)

// TestVMCopyArray проверяет сохранность исходного массива и полученную копию.
func TestVMCopyArray(t *testing.T) {
	machine := newTestMachine(t, string(readFixture(t, copyFilename)))
	assertCell(t, machine, copyStart, memoryZero)
	executeMachine(t, machine)
	for offset, expected := range []uint64{copyFirstValue, copySecondValue, copyThirdValue} {
		assertCell(t, machine, copyStart+uint64(offset), expected)
		assertCell(t, machine, copyDestination+uint64(offset), expected)
	}
	if len(machine.Stack()) != noInstructions {
		t.Fatalf("копирование оставило элементы на стеке: %v", machine.Stack())
	}
}

// TestVMLoadAndRead проверяет стек, чтение кода и незаписанной памяти.
func TestVMLoadAndRead(t *testing.T) {
	const source = `[{"op":"LOAD","value":38},{"op":"READ","address":0},{"op":"READ","address":2147483647}]`
	machine := newTestMachine(t, source)
	executeMachine(t, machine)
	expected := []uint64{exampleValue, modifiedByte, memoryZero}
	if !reflect.DeepEqual(machine.Stack(), expected) {
		t.Fatalf("неверный стек: %v", machine.Stack())
	}
	assertCell(t, machine, memoryZero, modifiedByte)
}

// TestVMWriteOrder проверяет снятие значения перед адресом.
func TestVMWriteOrder(t *testing.T) {
	const source = `[{"op":"LOAD","value":64},{"op":"LOAD","value":38},{"op":"WRITE"}]`
	machine := newTestMachine(t, source)
	executeMachine(t, machine)
	assertCell(t, machine, copyStart, exampleValue)
	if len(machine.Stack()) != noInstructions {
		t.Fatalf("WRITE не сняла оба элемента: %v", machine.Stack())
	}
}

// TestVMStackUnderflow проверяет нехватку обоих и одного элементов для WRITE.
func TestVMStackUnderflow(t *testing.T) {
	cases := []struct {
		source   string
		expected []uint64
	}{
		{`[{"op":"WRITE"}]`, nil},
		{`[{"op":"LOAD","value":64},{"op":"WRITE"}]`, []uint64{copyStart}},
	}
	for _, item := range cases {
		machine := newTestMachine(t, item.source)
		if err := machine.Execute(); err == nil || !strings.Contains(err.Error(), "WRITE") {
			t.Fatalf("ожидалась ошибка стека: %v", err)
		}
		if !reflect.DeepEqual(machine.Stack(), item.expected) {
			t.Fatalf("стек изменился при ошибке: %v", machine.Stack())
		}
	}
}

// TestVMMemoryBounds проверяет полный диапазон памяти и хранение значений шире байта.
func TestVMMemoryBounds(t *testing.T) {
	machine := newTestMachine(t, `[]`)
	assertCell(t, machine, assembler.MaximumAddress, memoryZero)
	if err := machine.Write(assembler.MaximumAddress, nonByteValue); err != nil {
		t.Fatal(err)
	}
	assertCell(t, machine, assembler.MaximumAddress, nonByteValue)
	invalid := assembler.MaximumAddress + outsideRangeStep
	if _, err := machine.Read(invalid); err == nil {
		t.Fatal("принят адрес чтения вне памяти")
	}
	if err := machine.Write(invalid, exampleValue); err == nil {
		t.Fatal("принят адрес записи вне памяти")
	}
}

// TestVMSelfModification подтверждает выборку измененного кода из общей памяти.
func TestVMSelfModification(t *testing.T) {
	const source = `[{"op":"LOAD","value":5},{"op":"LOAD","value":152},{"op":"WRITE"},{"op":"LOAD","value":1}]`
	machine := newTestMachine(t, source)
	executeMachine(t, machine)
	if !reflect.DeepEqual(machine.Stack(), []uint64{modifiedConstant}) {
		t.Fatalf("исполнена устаревшая копия кода: %v", machine.Stack())
	}
	assertCell(t, machine, modifiedAddress, modifiedByte)
}

// TestVMInstructionErrors проверяет испорченный код, обрыв, резервные биты и POW.
func TestVMInstructionErrors(t *testing.T) {
	for _, code := range []string{"00", "18", "26", "A6", "3D 00 00 00"} {
		machine, err := vm.New(decodeHex(t, code))
		if err != nil {
			t.Fatal(err)
		}
		if err := machine.Execute(); err == nil || !strings.HasPrefix(err.Error(), "адрес команды 0:") {
			t.Fatalf("неверная обработка команды %s: %v", code, err)
		}
	}
}

// TestVMNonByteCode запрещает исполнение ячеек кода, не помещающихся в байт.
func TestVMNonByteCode(t *testing.T) {
	const firstCodeByte = 0
	const secondCodeByte = 1
	for _, address := range []uint64{firstCodeByte, secondCodeByte} {
		machine := newTestMachine(t, `[{"op":"LOAD","value":0}]`)
		if err := machine.Write(address, nonByteValue); err != nil {
			t.Fatal(err)
		}
		if err := machine.Execute(); err == nil || !strings.Contains(err.Error(), "не содержит байт") {
			t.Fatalf("исполнена ячейка с потерей старших битов: %v", err)
		}
	}
}

// TestVMWriteInvalidAddress сохраняет стек при недопустимом адресе назначения.
func TestVMWriteInvalidAddress(t *testing.T) {
	const source = `[{"op":"READ","address":64},{"op":"LOAD","value":38},{"op":"WRITE"}]`
	machine := newTestMachine(t, source)
	invalid := assembler.MaximumAddress + outsideRangeStep
	if err := machine.Write(copyStart, invalid); err != nil {
		t.Fatal(err)
	}
	if err := machine.Execute(); err == nil || !strings.Contains(err.Error(), "вне диапазона") {
		t.Fatalf("ожидалась ошибка адреса: %v", err)
	}
	if !reflect.DeepEqual(machine.Stack(), []uint64{invalid, exampleValue}) {
		t.Fatalf("ошибочная запись изменила стек: %v", machine.Stack())
	}
}

// TestVMStackSnapshot запрещает изменение стека через результат Stack.
func TestVMStackSnapshot(t *testing.T) {
	machine := newTestMachine(t, `[{"op":"LOAD","value":38}]`)
	executeMachine(t, machine)
	snapshot := machine.Stack()
	snapshot[stackFirstIndex] = memoryZero
	if !reflect.DeepEqual(machine.Stack(), []uint64{exampleValue}) {
		t.Fatal("возвращена ссылка на внутренний стек")
	}
}

// TestVMEmptyProgram завершает пустую программу без чтения несуществующего кода.
func TestVMEmptyProgram(t *testing.T) {
	machine := newTestMachine(t, `[]`)
	executeMachine(t, machine)
	assertCell(t, machine, memoryZero, memoryZero)
}
