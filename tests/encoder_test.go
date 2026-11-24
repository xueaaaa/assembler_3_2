package assembler_test

import (
	"bytes"
	"strings"
	"testing"

	"assembler/src/assembler"
)

const unknownOpcode = 0

// TestEncodeSpecification проверяет байты каждой команды и всей программы.
func TestEncodeSpecification(t *testing.T) {
	commands, err := assembler.Encode(specificationFields)
	if err != nil {
		t.Fatal(err)
	}
	expected := expectedByteCommands(t)
	if len(commands) != len(expected) {
		t.Fatalf("неверное количество команд: %d", len(commands))
	}
	for index, command := range commands {
		if !bytes.Equal(command, expected[index]) {
			t.Errorf("команда %d: % X вместо % X", index, command, expected[index])
		}
	}
	if data := bytes.Join(commands, nil); len(data) != expectedFileSize {
		t.Fatalf("неверный размер программы: %d", len(data))
	}
}

// TestEncodeBoundaries проверяет переносы битов, длины и нулевые резервные биты.
func TestEncodeBoundaries(t *testing.T) {
	cases := []struct{ source, expected string }{
		{`[{"op":"LOAD","value":0}]`, "18 00"},
		{`[{"op":"LOAD","value":255}]`, "D8 3F"},
		{`[{"op":"READ","address":0}]`, "2E 00 00 00 00"},
		{`[{"op":"READ","address":2147483647}]`, "EE FF FF FF 1F"},
		{`[{"op":"POW","operand_offset":0,"result_offset":0}]`, "3D 00 00 00"},
		{`[{"op":"POW","operand_offset":2047,"result_offset":0}]`, "FD FF 01 00"},
		{`[{"op":"POW","operand_offset":0,"result_offset":2047}]`, "3D 00 FE 0F"},
		{`[{"op":"POW","operand_offset":2047,"result_offset":2047}]`, "FD FF FF 0F"},
	}
	for _, item := range cases {
		t.Run(item.source, func(t *testing.T) {
			instructions, err := assembler.Translate([]byte(item.source))
			if err != nil {
				t.Fatal(err)
			}
			commands, err := assembler.Encode(instructions)
			if err != nil {
				t.Fatal(err)
			}
			if data := bytes.Join(commands, nil); !bytes.Equal(data, decodeHex(t, item.expected)) {
				t.Fatalf("неверные байты: % X", data)
			}
		})
	}
}

// TestEncodeInvalidFields отклоняет некорректное промежуточное представление.
func TestEncodeInvalidFields(t *testing.T) {
	cases := [][]assembler.Field{
		nil,
		{{Name: "B", Value: expectedWriteOpcode}},
		{{Name: "A", Value: unknownOpcode}},
		{{Name: "A", Value: expectedLoadOpcode}},
		{{Name: "A", Value: expectedWriteOpcode}, {Name: "B", Value: exampleValue}},
		{{Name: "A", Value: expectedLoadOpcode}, {Name: "C", Value: exampleValue}},
		{{Name: "A", Value: expectedLoadOpcode}, {Name: "B", Value: maximumConstant + outsideRangeStep}},
		{{Name: "A", Value: expectedReadOpcode}, {Name: "B", Value: maximumAddress + outsideRangeStep}},
		{{Name: "A", Value: expectedPowOpcode}, {Name: "C"}, {Name: "B"}},
		{{Name: "A", Value: expectedPowOpcode}, {Name: "B", Value: maximumOffset + outsideRangeStep}, {Name: "C"}},
		{{Name: "A", Value: expectedPowOpcode}, {Name: "B"}, {Name: "C", Value: maximumOffset + outsideRangeStep}},
	}
	for _, fields := range cases {
		if data, err := assembler.EncodeInstruction(assembler.Instruction{Fields: fields}); err == nil || data != nil {
			t.Errorf("неверное представление принято: %v", fields)
		}
	}
}

// TestEncodeErrorPosition запрещает частичный код и указывает номер команды.
func TestEncodeErrorPosition(t *testing.T) {
	const expected = "команда 2:"
	valid := assembler.Instruction{Fields: []assembler.Field{{Name: "A", Value: expectedWriteOpcode}}}
	data, err := assembler.Encode([]assembler.Instruction{valid, {}})
	if err == nil || !strings.HasPrefix(err.Error(), expected) || data != nil {
		t.Fatalf("неверная обработка ошибки: %v, %v", data, err)
	}
}

// TestEncodeEmptyProgram проверяет отсутствие машинных команд для пустого ввода.
func TestEncodeEmptyProgram(t *testing.T) {
	commands, err := assembler.Encode(nil)
	if err != nil || len(commands) != noInstructions {
		t.Fatalf("ожидалась пустая программа: %v, %v", commands, err)
	}
}

// TestFormatBytes сверяет печать байтов с форматом спецификации.
func TestFormatBytes(t *testing.T) {
	var output strings.Builder
	for _, command := range expectedByteCommands(t) {
		output.WriteString(assembler.FormatBytes(command))
		output.WriteString("\n")
	}
	if expected := string(readFixture(t, byteFilename)); output.String() != expected {
		t.Fatalf("получено %q, ожидалось %q", output.String(), expected)
	}
}
