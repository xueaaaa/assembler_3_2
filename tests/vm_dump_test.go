package assembler_test

import (
	"bytes"
	"testing"

	"assembler/src/assembler"
)

// TestCSVCopyDump сравнивает весь CSV-дамп с независимым эталоном копирования.
func TestCSVCopyDump(t *testing.T) {
	machine := newTestMachine(t, string(readFixture(t, copyFilename)))
	executeMachine(t, machine)
	var output bytes.Buffer
	if err := machine.DumpCSV(&output, copyStart, copyEnd); err != nil {
		t.Fatal(err)
	}
	if expected := readFixture(t, copyExpectedFilename); !bytes.Equal(output.Bytes(), expected) {
		t.Fatalf("неверный дамп:\n%s", output.String())
	}
}

// TestCSVInclusiveBounds проверяет включение обеих границ и нулевых ячеек.
func TestCSVInclusiveBounds(t *testing.T) {
	const start = 67
	const end = 68
	const expected = "address,value\n67,0\n68,0\n"
	machine := newTestMachine(t, `[]`)
	var output bytes.Buffer
	if err := machine.DumpCSV(&output, start, end); err != nil || output.String() != expected {
		t.Fatalf("неверный диапазон дампа: %q, %v", output.String(), err)
	}
}

// TestCSVInvalidRanges запрещает неверный диапазон до начала вывода.
func TestCSVInvalidRanges(t *testing.T) {
	machine := newTestMachine(t, `[]`)
	ranges := [][writeRangeBounds]uint64{
		{copyEnd, copyStart},
		{memoryZero, assembler.MaximumAddress + outsideRangeStep},
	}
	for _, bounds := range ranges {
		var output bytes.Buffer
		err := machine.DumpCSV(&output, bounds[rangeStartIndex], bounds[rangeEndIndex])
		if err == nil || output.Len() != emptyOutputLength {
			t.Fatalf("неверный диапазон принят: %v", bounds)
		}
	}
}

const (
	writeRangeBounds = 2
	rangeStartIndex  = 0
	rangeEndIndex    = 1
)

// TestCSVWriteError проверяет ошибку записи, обнаруживаемую при Flush.
func TestCSVWriteError(t *testing.T) {
	machine := newTestMachine(t, `[]`)
	if err := machine.DumpCSV(closedWriter{}, copyStart, copyEnd); err == nil {
		t.Fatal("ошибка вывода CSV не обнаружена")
	}
}
