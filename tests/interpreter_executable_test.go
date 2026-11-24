package assembler_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

const (
	interpreterPackage  = "./src/interpreter"
	interpreterFilename = "interpreter"
)

// buildInterpreter собирает отдельный исполняемый файл интерпретатора.
func buildInterpreter(t *testing.T) string {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(t.TempDir(), interpreterFilename))
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command("go", "build", "-o", path, interpreterPackage)
	command.Dir = projectDirectory
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("ошибка сборки: %v\n%s", err, output)
	}
	return path
}

// TestCopyExecutablePipeline проверяет оба исполняемых файла и итоговый дамп.
func TestCopyExecutablePipeline(t *testing.T) {
	const sourceFilename = "copy with spaces.json"
	directory := t.TempDir()
	source := filepath.Join(directory, sourceFilename)
	program, dump := filepath.Join(directory, outputFilename), filepath.Join(directory, dumpFilename)
	writeFixture(t, source, readFixture(t, copyFilename))
	command := exec.Command(buildExecutable(t), source, program)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("ошибка ассемблера: %v, %s", err, output)
	}
	command = exec.Command(buildInterpreter(t), program, dump, copyStartArgument, copyEndArgument)
	if output, err := command.CombinedOutput(); err != nil || len(output) != emptyOutputLength {
		t.Fatalf("ошибка интерпретатора: %v, %s", err, output)
	}
	actual, err := os.ReadFile(dump)
	if err != nil || !bytes.Equal(actual, readFixture(t, copyExpectedFilename)) {
		t.Fatalf("неверный дамп после полного запуска: %s, %v", actual, err)
	}
}
