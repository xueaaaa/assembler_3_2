package assembler

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

const (
	exitSuccess       = 0
	exitFailure       = 1
	exitArgumentError = 2
	requiredPathCount = 2
	sourcePosition    = 0
	outputPosition    = 1
	cliName           = "assembler"
	testFlagName      = "test"
)

// options хранит параметры запуска; output зарезервирован для этапа 2.
type options struct {
	source string
	output string
	test   bool
}

// normalizeArguments разрешает переключатель тестирования после путей.
func normalizeArguments(arguments []string) []string {
	var flags, paths []string
	for _, argument := range arguments {
		switch argument {
		case "--test", "-test", "--help", "-h":
			flags = append(flags, argument)
		default:
			paths = append(paths, argument)
		}
	}
	return append(flags, paths...)
}

// parseArguments разбирает обязательные пути и переключатель --test.
func parseArguments(arguments []string, stderr io.Writer) (options, error) {
	parser := flag.NewFlagSet(cliName, flag.ContinueOnError)
	parser.SetOutput(stderr)
	var result options
	parser.BoolVar(&result.test, testFlagName, false, "Вывести промежуточные поля команд УВМ")
	parser.Usage = func() {
		fmt.Fprintln(stderr, "Использование: assembler [--test] source.json output.bin")
		fmt.Fprintln(stderr, "Этап 1: двоичный файл не создается; путь принят для этапа 2.")
		parser.PrintDefaults()
	}
	if err := parser.Parse(normalizeArguments(arguments)); err != nil {
		return options{}, err
	}
	if parser.NArg() != requiredPathCount {
		parser.Usage()
		return options{}, fmt.Errorf("необходимо указать исходный и выходной пути")
	}
	result.source = parser.Arg(sourcePosition)
	result.output = parser.Arg(outputPosition)
	return result, nil
}

// Run запускает первый этап ассемблера и возвращает код завершения CLI.
func Run(arguments []string, stdout, stderr io.Writer) int {
	argumentsOptions, err := parseArguments(arguments, stderr)
	if errors.Is(err, flag.ErrHelp) {
		return exitSuccess
	}
	if err != nil {
		fmt.Fprintln(stderr, "Ошибка:", err)
		return exitArgumentError
	}
	data, err := os.ReadFile(argumentsOptions.source)
	if err != nil {
		fmt.Fprintln(stderr, "Ошибка:", err)
		return exitFailure
	}
	instructions, err := Translate(data)
	if err != nil {
		fmt.Fprintln(stderr, "Ошибка:", err)
		return exitFailure
	}
	if argumentsOptions.test {
		if err := writeInstructions(stdout, instructions); err != nil {
			fmt.Fprintln(stderr, "Ошибка вывода:", err)
			return exitFailure
		}
	}
	return exitSuccess
}

// writeInstructions выводит поля команд, проверяя ошибки записи.
func writeInstructions(stdout io.Writer, instructions []Instruction) error {
	for _, instruction := range instructions {
		if _, err := fmt.Fprintln(stdout, instruction); err != nil {
			return err
		}
	}
	return nil
}
