package vm

import (
	"errors"
	"fmt"
	"os"
)

const dumpPermission = 0644

// runFile читает двоичный файл, исполняет его и сохраняет дамп после успеха.
func runFile(parameters options) error {
	data, err := os.ReadFile(parameters.program)
	if err != nil {
		return err
	}
	machine, err := New(data)
	if err != nil {
		return err
	}
	if err := machine.Execute(); err != nil {
		return err
	}
	return saveDump(machine, parameters)
}

// saveDump записывает CSV потоком, не позволяя затереть исходный бинарник.
func saveDump(machine *Machine, parameters options) error {
	if err := ValidateRange(parameters.start, parameters.end); err != nil {
		return err
	}
	programInfo, err := os.Stat(parameters.program)
	if err != nil {
		return err
	}
	dumpInfo, err := os.Stat(parameters.dump)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if dumpInfo != nil && os.SameFile(programInfo, dumpInfo) {
		return fmt.Errorf("двоичная программа и файл дампа должны различаться")
	}
	output, err := os.OpenFile(parameters.dump, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, dumpPermission)
	if err != nil {
		return err
	}
	return errors.Join(machine.DumpCSV(output, parameters.start, parameters.end), output.Close())
}
