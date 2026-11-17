package assembler_test

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"testing"
)

const (
	projectDirectory = ".."
	mainPackage      = "./src"
	binaryFilename   = "assembler"
)

// buildExecutable собирает приложение во временный каталог теста.
func buildExecutable(t *testing.T) string {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(t.TempDir(), binaryFilename))
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command("go", "build", "-o", path, mainPackage)
	command.Dir = projectDirectory
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("ошибка сборки: %v\n%s", err, output)
	}
	return path
}

// TestExecutable проверяет настоящий исполняемый файл с путями с пробелами.
func TestExecutable(t *testing.T) {
	const sourceFilename = "source with spaces.json"
	directory := t.TempDir()
	source := filepath.Join(directory, sourceFilename)
	output := filepath.Join(directory, outputFilename)
	writeFixture(t, source, readFixture(t, programFilename))
	command := exec.Command(buildExecutable(t), source, output, "--test")
	var stderr bytes.Buffer
	command.Stderr = &stderr
	actual, err := command.Output()
	if err != nil || stderr.Len() != emptyOutputLength {
		t.Fatalf("ошибка запуска: %v, %s", err, stderr.String())
	}
	if expected := readFixture(t, expectedFilename); !bytes.Equal(actual, expected) {
		t.Fatalf("получено %q, ожидалось %q", actual, expected)
	}
}
