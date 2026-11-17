package assembler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

const (
	arrayStart             = json.Delim('[')
	objectStart            = json.Delim('{')
	firstInstructionNumber = 1
)

// readCommands разбирает массив JSON и запрещает данные после массива.
func readCommands(data []byte) ([]map[string]json.RawMessage, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := expectOpening(decoder, arrayStart); err != nil {
		return nil, err
	}
	var commands []map[string]json.RawMessage
	for decoder.More() {
		command, err := readCommand(decoder)
		if err != nil {
			return nil, fmt.Errorf("команда %d: %w", len(commands)+firstInstructionNumber, err)
		}
		commands = append(commands, command)
	}
	if _, err := decoder.Token(); err != nil {
		return nil, err
	}
	if err := expectEnd(decoder); err != nil {
		return nil, err
	}
	return commands, nil
}

// expectOpening проверяет открывающий разделитель массива или объекта.
func expectOpening(decoder *json.Decoder, expected json.Delim) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	if token != expected {
		return fmt.Errorf("ожидается разделитель %q", expected)
	}
	return nil
}

// readCommand читает объект, обнаруживая повторяющиеся ключи.
func readCommand(decoder *json.Decoder) (map[string]json.RawMessage, error) {
	if err := expectOpening(decoder, objectStart); err != nil {
		return nil, err
	}
	command := make(map[string]json.RawMessage)
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		key, ok := token.(string)
		if !ok {
			return nil, fmt.Errorf("ключ команды должен быть строкой")
		}
		if _, exists := command[key]; exists {
			return nil, fmt.Errorf("повторяющийся ключ JSON: %s", key)
		}
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return nil, err
		}
		command[key] = value
	}
	if _, err := decoder.Token(); err != nil {
		return nil, err
	}
	return command, nil
}

// expectEnd проверяет, что после программы нет другого значения JSON.
func expectEnd(decoder *json.Decoder) error {
	_, err := decoder.Token()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}
	return fmt.Errorf("после массива программы обнаружены лишние данные")
}
