package vm

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strconv"

	"assembler/src/assembler"
)

const (
	exitSuccess           = 0
	exitFailure           = 1
	exitArguments         = 2
	requiredArgumentCount = 4
	programPosition       = 0
	dumpPosition          = 1
	startPosition         = 2
	endPosition           = 3
	cliName               = "interpreter"
)

// options хранит пути программы и дампа и включительный диапазон адресов.
type options struct {
	program string
	dump    string
	start   uint64
	end     uint64
}

// parseArguments проверяет пути и десятичные границы адресов памяти.
func parseArguments(arguments []string, stderr io.Writer) (options, error) {
	parser := flag.NewFlagSet(cliName, flag.ContinueOnError)
	parser.SetOutput(stderr)
	parser.Usage = func() {
		fmt.Fprintln(stderr, "Использование: interpreter program.bin dump.csv start end")
		fmt.Fprintln(stderr, "Обе границы диапазона включены; адреса задаются десятичными числами.")
	}
	if err := parser.Parse(arguments); err != nil {
		return options{}, err
	}
	if parser.NArg() != requiredArgumentCount {
		parser.Usage()
		return options{}, fmt.Errorf("необходимо указать два пути и две границы диапазона")
	}
	start, err := strconv.ParseUint(parser.Arg(startPosition), decimalBase, assembler.AddressWidth)
	if err != nil {
		return options{}, fmt.Errorf("неверный начальный адрес: %w", err)
	}
	end, err := strconv.ParseUint(parser.Arg(endPosition), decimalBase, assembler.AddressWidth)
	if err != nil {
		return options{}, fmt.Errorf("неверный конечный адрес: %w", err)
	}
	if err := ValidateRange(start, end); err != nil {
		return options{}, err
	}
	return options{parser.Arg(programPosition), parser.Arg(dumpPosition), start, end}, nil
}

// Run исполняет программу и сохраняет CSV-дамп; возвращает код завершения CLI.
func Run(arguments []string, stderr io.Writer) int {
	parameters, err := parseArguments(arguments, stderr)
	if errors.Is(err, flag.ErrHelp) {
		return exitSuccess
	}
	if err != nil {
		fmt.Fprintln(stderr, "Ошибка:", err)
		return exitArguments
	}
	if err := runFile(parameters); err != nil {
		fmt.Fprintln(stderr, "Ошибка:", err)
		return exitFailure
	}
	return exitSuccess
}
