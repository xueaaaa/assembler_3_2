package assembler_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"assembler/src/vm"
)

const dumpFilename = "dump.csv"

// runInterpreter вызывает CLI интерпретатора с перехватом потока ошибок.
func runInterpreter(arguments ...string) cliResult {
	var stderr bytes.Buffer
	status := vm.Run(arguments, &stderr)
	return cliResult{status: status, stderr: stderr.String()}
}

// TestInterpreterCopy проверяет полный путь от бинарника до CSV-файла.
func TestInterpreterCopy(t *testing.T) {
	directory := t.TempDir()
	program, dump := filepath.Join(directory, outputFilename), filepath.Join(directory, dumpFilename)
	writeFixture(t, program, assembleProgram(t, readFixture(t, copyFilename)))
	writeFixture(t, dump, []byte("old dump"))
	result := runInterpreter(program, dump, copyStartArgument, copyEndArgument)
	if result.status != successStatus || result.stderr != "" {
		t.Fatalf("ошибка запуска: %+v", result)
	}
	actual, err := os.ReadFile(dump)
	if err != nil || !bytes.Equal(actual, readFixture(t, copyExpectedFilename)) {
		t.Fatalf("неверный CSV: %s, %v", actual, err)
	}
}

// TestInterpreterArguments проверяет число аргументов и все ошибки диапазона.
func TestInterpreterArguments(t *testing.T) {
	cases := [][]string{
		nil, {"program"}, {"program", "dump", "0"}, {"program", "dump", "0", "1", "extra"},
		{"program", "dump", "-1", "2"}, {"program", "dump", "3", "2"},
		{"program", "dump", "0", "2147483648"}, {"program", "dump", "2147483648", "2147483648"},
		{"program", "dump", "0.5", "1"}, {"program", "dump", "0", "one"}, {"--unknown"},
	}
	for _, arguments := range cases {
		result := runInterpreter(arguments...)
		if result.status != argumentStatus || result.stderr == "" {
			t.Fatalf("неверная обработка аргументов %v: %+v", arguments, result)
		}
	}
}

// TestInterpreterHelp проверяет справку до чтения программы.
func TestInterpreterHelp(t *testing.T) {
	for _, argument := range []string{"--help", "-h"} {
		result := runInterpreter(argument)
		if result.status != successStatus || !strings.Contains(result.stderr, "program.bin dump.csv start end") {
			t.Fatalf("справка недоступна: %+v", result)
		}
	}
}

// TestInterpreterMissingProgram проверяет ошибку чтения и отсутствие дампа.
func TestInterpreterMissingProgram(t *testing.T) {
	directory := t.TempDir()
	dump := filepath.Join(directory, dumpFilename)
	result := runInterpreter(filepath.Join(directory, "missing.bin"), dump, copyStartArgument, copyEndArgument)
	if result.status != failureStatus || !strings.HasPrefix(result.stderr, "Ошибка:") {
		t.Fatalf("ожидалась ошибка чтения: %+v", result)
	}
	if _, err := os.Stat(dump); !os.IsNotExist(err) {
		t.Fatalf("дамп создан при ошибке: %v", err)
	}
}

// TestInterpreterExecutionErrors сохраняет старый дамп при ошибке исполнения.
func TestInterpreterExecutionErrors(t *testing.T) {
	const sentinel = "previous dump"
	directory := t.TempDir()
	program, dump := filepath.Join(directory, outputFilename), filepath.Join(directory, dumpFilename)
	writeFixture(t, dump, []byte(sentinel))
	for _, code := range []string{"00", "18", "26", "3D 00 00 00", "18 00 00"} {
		writeFixture(t, program, decodeHex(t, code))
		result := runInterpreter(program, dump, copyStartArgument, copyEndArgument)
		if result.status != failureStatus || !strings.HasPrefix(result.stderr, "Ошибка:") {
			t.Fatalf("ожидалась ошибка исполнения: %+v", result)
		}
		actual, err := os.ReadFile(dump)
		if err != nil || string(actual) != sentinel {
			t.Fatalf("дамп изменен при ошибке: %q, %v", actual, err)
		}
	}
}

// TestInterpreterDumpWriteErrors проверяет недоступные выходные пути.
func TestInterpreterDumpWriteErrors(t *testing.T) {
	directory := t.TempDir()
	program := filepath.Join(directory, outputFilename)
	writeFixture(t, program, nil)
	for _, dump := range []string{directory, filepath.Join(directory, "missing", dumpFilename)} {
		result := runInterpreter(program, dump, copyStartArgument, copyEndArgument)
		if result.status != failureStatus || !strings.HasPrefix(result.stderr, "Ошибка:") {
			t.Fatalf("ожидалась ошибка записи: %+v", result)
		}
	}
}

// TestInterpreterMaximumRange проверяет последнюю ячейку адресного пространства.
func TestInterpreterMaximumRange(t *testing.T) {
	const maximumArgument = "2147483647"
	const expected = "address,value\n2147483647,0\n"
	directory := t.TempDir()
	program, dump := filepath.Join(directory, outputFilename), filepath.Join(directory, dumpFilename)
	writeFixture(t, program, nil)
	result := runInterpreter(program, dump, maximumArgument, maximumArgument)
	if result.status != successStatus || result.stderr != "" {
		t.Fatalf("неверная обработка границы: %+v", result)
	}
	actual, err := os.ReadFile(dump)
	if err != nil || string(actual) != expected {
		t.Fatalf("неверный CSV на границе памяти: %q, %v", actual, err)
	}
}

// TestInterpreterSourceProtection запрещает замену программы ее дампом.
func TestInterpreterSourceProtection(t *testing.T) {
	program := filepath.Join(t.TempDir(), outputFilename)
	original := assembleProgram(t, []byte(`[{"op":"LOAD","value":38}]`))
	writeFixture(t, program, original)
	result := runInterpreter(program, program, copyStartArgument, copyEndArgument)
	if result.status != failureStatus || !strings.Contains(result.stderr, "должны различаться") {
		t.Fatalf("ожидался отказ перезаписи программы: %+v", result)
	}
	actual, err := os.ReadFile(program)
	if err != nil || !bytes.Equal(actual, original) {
		t.Fatalf("двоичный исходник изменен: %v", err)
	}
}
