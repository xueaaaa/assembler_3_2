GO ?= go
SOURCE ?= examples/specification.json
OUTPUT ?= build/program.bin
PROGRAM ?= build/copy.bin
DUMP ?= build/memory.csv
START ?= 64
END ?= 82

.PHONY: build run demo interpret copy test check

build:
	mkdir -p build
	$(GO) build -o build/assembler ./src
	$(GO) build -o build/interpreter ./src/interpreter

run:
	mkdir -p build
	$(GO) run ./src "$(SOURCE)" "$(OUTPUT)"

demo:
	mkdir -p build
	$(GO) run ./src --test "$(SOURCE)" "$(OUTPUT)"

interpret:
	$(GO) run ./src/interpreter "$(PROGRAM)" "$(DUMP)" "$(START)" "$(END)"

copy:
	mkdir -p build
	$(GO) run ./src examples/copy.json "$(PROGRAM)"
	$(GO) run ./src/interpreter "$(PROGRAM)" "$(DUMP)" "$(START)" "$(END)"
	cat "$(DUMP)"

test:
	$(GO) test ./... -v

check: test
	$(GO) vet ./...
	test -z "$$(gofmt -l src tests)"
