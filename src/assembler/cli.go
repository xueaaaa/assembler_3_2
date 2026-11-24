package assembler

import (
	"errors"
	"flag"
	"fmt"
	"io"
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

// options хранит исходный путь, выходной путь и тестовый режим.
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
	parser.BoolVar(&result.test, testFlagName, false, "Вывести машинный код в байтовом формате")
	parser.Usage = func() {
		fmt.Fprintln(stderr, "Использование: assembler [--test] source.json output.bin")
		fmt.Fprintln(stderr, "Этап 2: записать машинный код и вывести размер файла в байтах.")
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

// Run запускает ассемблер и возвращает код завершения CLI.
func Run(arguments []string, stdout, stderr io.Writer) int {
	argumentsOptions, err := parseArguments(arguments, stderr)
	if errors.Is(err, flag.ErrHelp) {
		return exitSuccess
	}
	if err != nil {
		fmt.Fprintln(stderr, "Ошибка:", err)
		return exitArgumentError
	}
	commands, err := assembleFile(argumentsOptions)
	if err != nil {
		fmt.Fprintln(stderr, "Ошибка:", err)
		return exitFailure
	}
	if err := writeResult(stdout, commands, argumentsOptions.test); err != nil {
		fmt.Fprintln(stderr, "Ошибка вывода:", err)
		return exitFailure
	}
	return exitSuccess
}
