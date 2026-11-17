GO ?= go
SOURCE ?= examples/specification.json
OUTPUT ?= build/program.bin

.PHONY: build run demo test check

build:
	mkdir -p build
	$(GO) build -o build/assembler ./src

run:
	$(GO) run ./src $(SOURCE) $(OUTPUT)

demo:
	$(GO) run ./src --test $(SOURCE) $(OUTPUT)

test:
	$(GO) test ./... -v

check: test
	$(GO) vet ./...
	test -z "$$(gofmt -l src tests)"
