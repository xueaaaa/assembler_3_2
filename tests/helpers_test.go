package assembler_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"assembler/src/assembler"
)

const (
	exampleDirectory  = "../examples"
	programFilename   = "specification.json"
	expectedFilename  = "expected.txt"
	outputFilename    = "program.bin"
	fixturePermission = 0600
	successStatus     = 0
	failureStatus     = 1
	argumentStatus    = 2
	noInstructions    = 0
	emptyOutputLength = 0
	minimumOperand    = 0
	negativeOperand   = -1
	outsideRangeStep  = 1
)

// readFixture читает пример или эталон из каталога examples.
func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(exampleDirectory, name))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// writeFixture записывает временный исходник или существующий результат.
func writeFixture(t *testing.T, path string, content []byte) {
	t.Helper()
	if err := os.WriteFile(path, content, fixturePermission); err != nil {
		t.Fatal(err)
	}
}

// cliResult хранит код завершения и оба потока приложения.
type cliResult struct {
	status int
	stdout string
	stderr string
}

// runCLI вызывает интерфейс приложения с перехватом вывода.
func runCLI(arguments ...string) cliResult {
	var stdout, stderr bytes.Buffer
	status := assembler.Run(arguments, &stdout, &stderr)
	return cliResult{status, stdout.String(), stderr.String()}
}
