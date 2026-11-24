package assembler_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"assembler/src/assembler"
)

// TestCLITestMode проверяет вывод и оба положения переключателя --test.
func TestCLITestMode(t *testing.T) {
	source := filepath.Join(exampleDirectory, programFilename)
	output := filepath.Join(t.TempDir(), outputFilename)
	for _, arguments := range [][]string{{"--test", source, output}, {source, output, "--test"}} {
		result := runCLI(arguments...)
		if result.status != successStatus || result.stderr != "" {
			t.Fatalf("ошибка запуска: %+v", result)
		}
		if expected := expectedCLIOutput(t); result.stdout != expected {
			t.Fatalf("получено %q, ожидалось %q", result.stdout, expected)
		}
		assertMachineCode(t, output)
	}
}

// TestCLINormalMode проверяет вывод размера и замену существующего результата.
func TestCLINormalMode(t *testing.T) {
	const sentinel = "existing result"
	source := filepath.Join(exampleDirectory, programFilename)
	output := filepath.Join(t.TempDir(), outputFilename)
	writeFixture(t, output, []byte(sentinel))
	result := runCLI(source, output)
	if result.status != successStatus || result.stdout != expectedSizeLine || result.stderr != "" {
		t.Fatalf("неверный результат: %+v", result)
	}
	assertMachineCode(t, output)
}

// TestCLIMissingSource проверяет обработку отсутствующего исходного файла.
func TestCLIMissingSource(t *testing.T) {
	directory := t.TempDir()
	result := runCLI(filepath.Join(directory, programFilename), filepath.Join(directory, outputFilename))
	if result.status != failureStatus || result.stdout != "" || !strings.HasPrefix(result.stderr, "Ошибка:") {
		t.Fatalf("ожидалась ошибка чтения: %+v", result)
	}
}

// TestCLIInvalidPrograms проверяет ошибки до вывода и изменения результата.
func TestCLIInvalidPrograms(t *testing.T) {
	const sentinel = "existing result"
	directory := t.TempDir()
	source, output := filepath.Join(directory, programFilename), filepath.Join(directory, outputFilename)
	writeFixture(t, output, []byte(sentinel))
	for _, program := range []string{`[`, `[{"op":"WRITE"},{}]`, "\xff"} {
		writeFixture(t, source, []byte(program))
		result := runCLI(source, output, "--test")
		if result.status != failureStatus || result.stdout != "" || !strings.HasPrefix(result.stderr, "Ошибка:") {
			t.Fatalf("ожидалась ошибка программы: %+v", result)
		}
		content, err := os.ReadFile(output)
		if err != nil || string(content) != sentinel {
			t.Fatalf("существующий результат изменен: %q, %v", content, err)
		}
	}
}

// TestCLIArguments проверяет обязательность путей и неизвестные флаги.
func TestCLIArguments(t *testing.T) {
	for _, arguments := range [][]string{nil, {"source"}, {"source", "output", "extra"}, {"--unknown"}} {
		result := runCLI(arguments...)
		if result.status != argumentStatus || result.stdout != "" || result.stderr == "" {
			t.Fatalf("неверная обработка аргументов %v: %+v", arguments, result)
		}
	}
}

// TestCLIHelp проверяет доступность справки без исходной программы.
func TestCLIHelp(t *testing.T) {
	for _, argument := range []string{"--help", "-h"} {
		result := runCLI(argument)
		if result.status != successStatus || !strings.Contains(result.stderr, "source.json output.bin") {
			t.Fatalf("справка недоступна: %+v", result)
		}
	}
}

// closedWriter имитирует недоступный поток стандартного вывода.
type closedWriter struct{}

// Write сообщает об ошибке записи без принятых байтов.
func (closedWriter) Write(_ []byte) (int, error) {
	const rejectedByteCount = 0
	return rejectedByteCount, io.ErrClosedPipe
}

// TestCLIOutputFailure проверяет код завершения при ошибке вывода.
func TestCLIOutputFailure(t *testing.T) {
	var stderr bytes.Buffer
	source := filepath.Join(exampleDirectory, programFilename)
	output := filepath.Join(t.TempDir(), outputFilename)
	status := assembler.Run([]string{"--test", source, output}, closedWriter{}, &stderr)
	if status != failureStatus || !strings.HasPrefix(stderr.String(), "Ошибка вывода:") {
		t.Fatalf("ожидалась ошибка вывода: %d, %s", status, stderr.String())
	}
}
