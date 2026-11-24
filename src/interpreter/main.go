// Command interpreter исполняет двоичный код УВМ и сохраняет CSV-дамп памяти.
package main

import (
	"os"

	"assembler/src/vm"
)

const firstUserArgument = 1

// main передает аргументы и поток ошибок интерпретатору.
func main() {
	os.Exit(vm.Run(os.Args[firstUserArgument:], os.Stderr))
}
