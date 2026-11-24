package assembler_test

import (
	"fmt"
	"reflect"
	"testing"

	"assembler/src/assembler"
)

const opcodeByteIndex = 0

// TestDecodeSpecification сверяет декодирование со всеми полями задания.
func TestDecodeSpecification(t *testing.T) {
	for index, command := range expectedByteCommands(t) {
		instruction, err := assembler.DecodeInstruction(command)
		if err != nil || !reflect.DeepEqual(instruction, specificationFields[index]) {
			t.Fatalf("неверное декодирование: %v, %v", instruction, err)
		}
		size, err := assembler.InstructionSize(command[opcodeByteIndex])
		if err != nil || size != len(command) {
			t.Fatalf("неверная длина команды: %d, %v", size, err)
		}
	}
}

// TestDecodeOperandLimits проверяет восстановление предельных значений полей.
func TestDecodeOperandLimits(t *testing.T) {
	for _, limit := range operandLimits {
		for _, value := range []int64{minimumOperand, limit.maximum} {
			instructions, err := assembler.Translate([]byte(fmt.Sprintf(limit.template, value)))
			if err != nil {
				t.Fatal(err)
			}
			for _, expected := range instructions {
				data, err := assembler.EncodeInstruction(expected)
				if err != nil {
					t.Fatal(err)
				}
				actual, err := assembler.DecodeInstruction(data)
				if err != nil || !reflect.DeepEqual(actual, expected) {
					t.Fatalf("неверно восстановлено поле %s: %v, %v", limit.name, actual, err)
				}
			}
		}
	}
}

// TestDecodeInvalidCommands отклоняет длины, коды и ненулевые резервные биты.
func TestDecodeInvalidCommands(t *testing.T) {
	commands := []string{"", "00", "98", "98 09 00", "98 89", "6E 4D 00 00 80", "A6", "3D 2D 02 80"}
	for _, command := range commands {
		t.Run(command, func(t *testing.T) {
			if _, err := assembler.DecodeInstruction(decodeHex(t, command)); err == nil {
				t.Fatal("принята некорректная двоичная команда")
			}
		})
	}
	if _, err := assembler.InstructionSize(unknownOpcode); err == nil {
		t.Fatal("принят неизвестный код операции")
	}
}
