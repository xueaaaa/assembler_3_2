package assembler_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCLIEmptyProgram создает пустой файл и выводит нулевой размер.
func TestCLIEmptyProgram(t *testing.T) {
	const expected = "Размер двоичного файла: 0 байт\n"
	directory := t.TempDir()
	source, output := filepath.Join(directory, programFilename), filepath.Join(directory, outputFilename)
	writeFixture(t, source, []byte(`[]`))
	writeFixture(t, output, []byte("old result"))
	result := runCLI(source, output, "--test")
	if result.status != successStatus || result.stdout != expected || result.stderr != "" {
		t.Fatalf("неверный результат: %+v", result)
	}
	data, err := os.ReadFile(output)
	if err != nil || len(data) != emptyOutputLength {
		t.Fatalf("ожидался пустой файл: % X, %v", data, err)
	}
}

// TestCLIFileWriteErrors не сообщает об успехе при недоступном выходном пути.
func TestCLIFileWriteErrors(t *testing.T) {
	source := filepath.Join(exampleDirectory, programFilename)
	directory := t.TempDir()
	for _, output := range []string{directory, filepath.Join(directory, "missing", outputFilename)} {
		result := runCLI(source, output, "--test")
		if result.status != failureStatus || result.stdout != "" || !strings.HasPrefix(result.stderr, "Ошибка:") {
			t.Fatalf("ожидалась ошибка записи: %+v", result)
		}
	}
}

// TestCLISameSourceOutput запрещает перезапись исходника машинным кодом.
func TestCLISameSourceOutput(t *testing.T) {
	source := filepath.Join(t.TempDir(), programFilename)
	original := readFixture(t, programFilename)
	writeFixture(t, source, original)
	assertSourceProtected(t, source, source, original)
}

// TestCLISourceAliases защищает исходник при символьных и жестких ссылках.
func TestCLISourceAliases(t *testing.T) {
	creators := map[string]func(string, string) error{"symlink": os.Symlink, "hardlink": os.Link}
	for name, create := range creators {
		t.Run(name, func(t *testing.T) {
			directory := t.TempDir()
			source, output := filepath.Join(directory, programFilename), filepath.Join(directory, outputFilename)
			original := readFixture(t, programFilename)
			writeFixture(t, source, original)
			if err := create(source, output); err != nil {
				t.Fatal(err)
			}
			assertSourceProtected(t, source, output, original)
		})
	}
}

// assertSourceProtected проверяет отказ и сохранность исходного файла.
func assertSourceProtected(t *testing.T, source, output string, original []byte) {
	t.Helper()
	result := runCLI(source, output)
	if result.status != failureStatus || result.stdout != "" || !strings.Contains(result.stderr, "должны различаться") {
		t.Fatalf("ожидался отказ перезаписи исходника: %+v", result)
	}
	actual, err := os.ReadFile(source)
	if err != nil || !bytes.Equal(actual, original) {
		t.Fatalf("исходный файл поврежден: %v", err)
	}
}
