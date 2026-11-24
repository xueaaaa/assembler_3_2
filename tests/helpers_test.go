package assembler_test

import (
	"bytes"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"assembler/src/assembler"
)

const (
	exampleDirectory  = "../examples"
	programFilename   = "specification.json"
	expectedFilename  = "expected.txt"
	byteFilename      = "expected-bytes.txt"
	expectedFileSize  = 12
	expectedSizeLine  = "Размер двоичного файла: 12 байт\n"
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

// decodeHex читает текстовый эталон байтов, независимый от кодировщика.
func decodeHex(t *testing.T, text string) []byte {
	t.Helper()
	compact := strings.NewReplacer("0x", "", ",", "", " ", "", "\n", "").Replace(text)
	data, err := hex.DecodeString(compact)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// expectedByteCommands читает отдельные эталонные команды из задания.
func expectedByteCommands(t *testing.T) [][]byte {
	t.Helper()
	lines := strings.Split(strings.TrimSpace(string(readFixture(t, byteFilename))), "\n")
	commands := make([][]byte, len(lines))
	for index, line := range lines {
		commands[index] = decodeHex(t, line)
	}
	return commands
}

// expectedMachineCode объединяет независимые эталонные команды.
func expectedMachineCode(t *testing.T) []byte {
	t.Helper()
	return bytes.Join(expectedByteCommands(t), nil)
}

// expectedCLIOutput возвращает размер файла и байтовый вывод спецификации.
func expectedCLIOutput(t *testing.T) string {
	t.Helper()
	return expectedSizeLine + string(readFixture(t, byteFilename))
}

// assertMachineCode проверяет содержимое файла и его фактический размер.
func assertMachineCode(t *testing.T, path string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != expectedFileSize || !bytes.Equal(data, expectedMachineCode(t)) {
		t.Fatalf("неверный машинный код (%d байт): % X", len(data), data)
	}
}
