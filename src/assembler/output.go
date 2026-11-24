package assembler

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
)

const outputPermission = 0644

// assembleFile переводит исходный файл и сохраняет объединенные команды.
func assembleFile(arguments options) ([][]byte, error) {
	data, err := os.ReadFile(arguments.source)
	if err != nil {
		return nil, err
	}
	instructions, err := Translate(data)
	if err != nil {
		return nil, err
	}
	commands, err := Encode(instructions)
	if err != nil {
		return nil, err
	}
	if err := saveProgram(arguments.source, arguments.output, bytes.Join(commands, nil)); err != nil {
		return nil, err
	}
	return commands, nil
}

// saveProgram записывает машинный код, запрещая перезапись исходного файла.
func saveProgram(source, output string, data []byte) error {
	sourceInfo, err := os.Stat(source)
	if err != nil {
		return err
	}
	outputInfo, err := os.Stat(output)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if outputInfo != nil && os.SameFile(sourceInfo, outputInfo) {
		return fmt.Errorf("исходный и выходной файлы должны различаться")
	}
	return os.WriteFile(output, data, outputPermission)
}

// writeResult выводит размер файла и при необходимости байты каждой команды.
func writeResult(stdout io.Writer, commands [][]byte, test bool) error {
	var size int
	for _, command := range commands {
		size += len(command)
	}
	if _, err := fmt.Fprintf(stdout, "Размер двоичного файла: %d байт\n", size); err != nil {
		return err
	}
	if !test {
		return nil
	}
	for _, command := range commands {
		if _, err := fmt.Fprintln(stdout, FormatBytes(command)); err != nil {
			return err
		}
	}
	return nil
}
