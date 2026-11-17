package assembler_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"assembler/src/assembler"
)

const (
	expectedLoadOpcode  = 24
	expectedReadOpcode  = 46
	expectedWriteOpcode = 38
	expectedPowOpcode   = 61
	exampleValue        = 38
	exampleAddress      = 309
	exampleOperandShift = 180
	exampleResultShift  = 1
	maximumConstant     = 255
	maximumAddress      = 2147483647
	maximumOffset       = 2047
)

// specificationFields повторяет независимые эталоны из задания.
var specificationFields = []assembler.Instruction{
	{Fields: []assembler.Field{
		{Name: "A", Value: expectedLoadOpcode}, {Name: "B", Value: exampleValue},
	}},
	{Fields: []assembler.Field{
		{Name: "A", Value: expectedReadOpcode}, {Name: "B", Value: exampleAddress},
	}},
	{Fields: []assembler.Field{{Name: "A", Value: expectedWriteOpcode}}},
	{Fields: []assembler.Field{
		{Name: "A", Value: expectedPowOpcode},
		{Name: "B", Value: exampleOperandShift},
		{Name: "C", Value: exampleResultShift},
	}},
}

// TestSpecificationFields сравнивает все поля с примерами спецификации.
func TestSpecificationFields(t *testing.T) {
	instructions, err := assembler.Translate(readFixture(t, programFilename))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(instructions, specificationFields) {
		t.Fatalf("получено %v, ожидалось %v", instructions, specificationFields)
	}
}

// TestSpecificationText сверяет форматирование с эталонным файлом.
func TestSpecificationText(t *testing.T) {
	var output strings.Builder
	for _, instruction := range specificationFields {
		fmt.Fprintln(&output, instruction)
	}
	if expected := string(readFixture(t, expectedFilename)); output.String() != expected {
		t.Fatalf("получено %q, ожидалось %q", output.String(), expected)
	}
}

// operandLimit задает независимый предел одного операнда спецификации.
type operandLimit struct {
	name     string
	template string
	maximum  int64
}

var operandLimits = []operandLimit{
	{"value", `[{"op":"LOAD","value":%d}]`, maximumConstant},
	{"address", `[{"op":"READ","address":%d}]`, maximumAddress},
	{"operand_offset", `[{"op":"POW","operand_offset":%d,"result_offset":0}]`, maximumOffset},
	{"result_offset", `[{"op":"POW","operand_offset":0,"result_offset":%d}]`, maximumOffset},
}

// TestOperandBoundaries проверяет обе границы каждого поля и выход за них.
func TestOperandBoundaries(t *testing.T) {
	for _, limit := range operandLimits {
		t.Run(limit.name, func(t *testing.T) {
			for _, value := range []int64{minimumOperand, limit.maximum} {
				data := []byte(fmt.Sprintf(limit.template, value))
				if _, err := assembler.Translate(data); err != nil {
					t.Errorf("допустимое значение %d: %v", value, err)
				}
			}
			for _, value := range []int64{negativeOperand, limit.maximum + outsideRangeStep} {
				data := []byte(fmt.Sprintf(limit.template, value))
				if _, err := assembler.Translate(data); err == nil {
					t.Errorf("принято недопустимое значение %d", value)
				}
			}
		})
	}
}

// TestInvalidOperandTypes запрещает нецелые числа и прочие типы JSON.
func TestInvalidOperandTypes(t *testing.T) {
	values := []string{`true`, `false`, `null`, `"38"`, `38.0`, `1e2`, `[]`, `{}`}
	for _, limit := range operandLimits {
		for _, value := range values {
			t.Run(limit.name+"/"+value, func(t *testing.T) {
				template := strings.ReplaceAll(limit.template, "%d", "%s")
				if _, err := assembler.Translate([]byte(fmt.Sprintf(template, value))); err == nil {
					t.Fatal("принят операнд неверного типа")
				}
			})
		}
	}
}

// TestInvalidPrograms запрещает ошибки структуры и синтаксиса программы.
func TestInvalidPrograms(t *testing.T) {
	programs := []string{
		``, `[`, `{}`, `null`, `true`, `"WRITE"`, `[null]`, `[[]]`, `["WRITE"]`,
		`[{}]`, `[{"op":null}]`, `[{"op":[]}]`, `[{"op":24}]`,
		`[{"op":"UNKNOWN"}]`, `[{"op":"load","value":38}]`,
		`[{"op":"LOAD"}]`, `[{"op":"READ"}]`,
		`[{"op":"POW","operand_offset":180}]`,
		`[{"op":"WRITE","value":38}]`,
		`[{"op":"LOAD","address":38}]`,
		`[{"op":"LOAD","value":38,"extra":null}]`,
		`[{"op":"WRITE","op":"WRITE"}]`,
		`[{"op":"LOAD","value":38,"value":39}]`,
		`[{"op":"WRITE"},]`, `[] []`, `[] garbage`,
	}
	for _, program := range programs {
		t.Run(program, func(t *testing.T) {
			if _, err := assembler.Translate([]byte(program)); err == nil {
				t.Fatal("принята некорректная программа")
			}
		})
	}
}

// TestInvalidEncoding запрещает недопустимые байты UTF-8 во входном файле.
func TestInvalidEncoding(t *testing.T) {
	const invalidUTF8 = "\xff"
	if _, err := assembler.Translate([]byte(invalidUTF8)); err == nil {
		t.Fatal("принята неверная кодировка")
	}
}

// TestEmptyProgram допускает пустой массив команд.
func TestEmptyProgram(t *testing.T) {
	instructions, err := assembler.Translate([]byte(`[]`))
	if err != nil {
		t.Fatal(err)
	}
	if len(instructions) != noInstructions {
		t.Fatalf("ожидалась пустая программа, получено %v", instructions)
	}
}

// TestCommandOrder сохраняет порядок команд независимо от порядка ключей.
func TestCommandOrder(t *testing.T) {
	const source = `[{"value":38,"op":"LOAD"},{"op":"WRITE"},{"value":38,"op":"LOAD"}]`
	const expected = "(A=24, B=38)\n(A=38)\n(A=24, B=38)\n"
	instructions, err := assembler.Translate([]byte(source))
	if err != nil {
		t.Fatal(err)
	}
	var output strings.Builder
	for _, instruction := range instructions {
		fmt.Fprintln(&output, instruction)
	}
	if output.String() != expected {
		t.Fatalf("неверный порядок команд: %s", output.String())
	}
}

// TestErrorPosition включает номер неверной команды в сообщение об ошибке.
func TestErrorPosition(t *testing.T) {
	const source = `[{"op":"WRITE"},{"op":"LOAD"}]`
	const expected = "команда 2:"
	if _, err := assembler.Translate([]byte(source)); err == nil || !strings.Contains(err.Error(), expected) {
		t.Fatalf("ожидался номер команды в ошибке, получено %v", err)
	}
}
