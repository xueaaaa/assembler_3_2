package vm

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"

	"assembler/src/assembler"
)

const decimalBase = 10

// ValidateRange проверяет включительные границы диапазона дампа.
func ValidateRange(start, end uint64) error {
	if start > end || end > assembler.MaximumAddress {
		return fmt.Errorf("диапазон должен удовлетворять 0 <= start <= end <= %d", assembler.MaximumAddress)
	}
	return nil
}

// DumpCSV записывает адрес и значение каждой ячейки диапазона, включая нули.
func (machine *Machine) DumpCSV(output io.Writer, start, end uint64) error {
	if err := ValidateRange(start, end); err != nil {
		return err
	}
	writer := csv.NewWriter(output)
	if err := writer.Write([]string{"address", "value"}); err != nil {
		return err
	}
	for address := start; address <= end; address++ {
		record := []string{
			strconv.FormatUint(address, decimalBase),
			strconv.FormatUint(machine.memory[address], decimalBase),
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}
