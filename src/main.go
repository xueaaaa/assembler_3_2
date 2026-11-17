// Command assembler переводит JSON-программу УВМ в промежуточные команды.
package main

import (
	"os"

	"assembler/src/assembler"
)

const firstUserArgument = 1

// main передает аргументы и стандартные потоки приложению.
func main() {
	os.Exit(assembler.Run(os.Args[firstUserArgument:], os.Stdout, os.Stderr))
}
