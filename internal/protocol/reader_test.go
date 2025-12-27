package protocol

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

// Test ReadCommand with array format (RESP)
func TestReadCommand_ArrayFormat(t *testing.T) {
	// *3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n
	input := "*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n"
	reader := bufio.NewReader(strings.NewReader(input))

	args, err := ReadCommand(reader)
	if err != nil {
		t.Fatalf("ReadCommand failed: %v", err)
	}

	if len(args) != 3 {
		t.Fatalf("Expected 3 args, got %d", len(args))
	}
	if string(args[0]) != "SET" {
		t.Fatalf("Expected 'SET', got %s", string(args[0]))
	}
	if string(args[1]) != "key" {
		t.Fatalf("Expected 'key', got %s", string(args[1]))
	}
	if string(args[2]) != "value" {
		t.Fatalf("Expected 'value', got %s", string(args[2]))
	}
}

// Test ReadCommand with inline format
func TestReadCommand_InlineFormat(t *testing.T) {
	input := "GET key\n"
	reader := bufio.NewReader(strings.NewReader(input))

	args, err := ReadCommand(reader)
	if err != nil {
		t.Fatalf("ReadCommand failed: %v", err)
	}

	if len(args) != 2 {
		t.Fatalf("Expected 2 args, got %d", len(args))
	}
	if string(args[0]) != "GET" {
		t.Fatalf("Expected 'GET', got %s", string(args[0]))
	}
	if string(args[1]) != "key" {
		t.Fatalf("Expected 'key', got %s", string(args[1]))
	}
}

func TestReadCommand_InlineFormat_MultipleArgs(t *testing.T) {
	input := "SET key value with spaces\n"
	reader := bufio.NewReader(strings.NewReader(input))

	args, err := ReadCommand(reader)
	if err != nil {
		t.Fatalf("ReadCommand failed: %v", err)
	}

	if len(args) != 5 {
		t.Fatalf("Expected 5 args, got %d", len(args))
	}
}

func TestReadCommand_InlineFormat_Trimmed(t *testing.T) {
	input := "  GET  key  \n"
	reader := bufio.NewReader(strings.NewReader(input))

	args, err := ReadCommand(reader)
	if err != nil {
		t.Fatalf("ReadCommand failed: %v", err)
	}

	if len(args) != 2 {
		t.Fatalf("Expected 2 args, got %d", len(args))
	}
	if string(args[0]) != "GET" {
		t.Fatalf("Expected 'GET', got %s", string(args[0]))
	}
	if string(args[1]) != "key" {
		t.Fatalf("Expected 'key', got %s", string(args[1]))
	}
}

// Test readArray
func TestReadArray(t *testing.T) {
	input := "*2\r\n$3\r\nGET\r\n$3\r\nkey\r\n"
	reader := bufio.NewReader(strings.NewReader(input))

	args, err := readArray(reader)
	if err != nil {
		t.Fatalf("readArray failed: %v", err)
	}

	if len(args) != 2 {
		t.Fatalf("Expected 2 args, got %d", len(args))
	}
	if string(args[0]) != "GET" {
		t.Fatalf("Expected 'GET', got %s", string(args[0]))
	}
	if string(args[1]) != "key" {
		t.Fatalf("Expected 'key', got %s", string(args[1]))
	}
}

func TestReadArray_Empty(t *testing.T) {
	input := "*0\r\n"
	reader := bufio.NewReader(strings.NewReader(input))

	args, err := readArray(reader)
	if err != nil {
		t.Fatalf("readArray failed: %v", err)
	}

	if len(args) != 0 {
		t.Fatalf("Expected 0 args, got %d", len(args))
	}
}

func TestReadArray_Negative(t *testing.T) {
	input := "*-1\r\n"
	reader := bufio.NewReader(strings.NewReader(input))

	args, err := readArray(reader)
	if err != nil {
		t.Fatalf("readArray failed: %v", err)
	}

	if len(args) != 0 {
		t.Fatalf("Expected 0 args for negative count, got %d", len(args))
	}
}

func TestReadArray_Large(t *testing.T) {
	// Create array with 100 elements
	var builder strings.Builder
	builder.WriteString("*100\r\n")
	for i := 0; i < 100; i++ {
		builder.WriteString("$1\r\n")
		builder.WriteString("a\r\n")
	}

	reader := bufio.NewReader(strings.NewReader(builder.String()))

	args, err := readArray(reader)
	if err != nil {
		t.Fatalf("readArray failed: %v", err)
	}

	if len(args) != 100 {
		t.Fatalf("Expected 100 args, got %d", len(args))
	}
}

func TestReadArray_InvalidPrefix(t *testing.T) {
	input := "X2\r\n"
	reader := bufio.NewReader(strings.NewReader(input))

	_, err := readArray(reader)
	if err == nil {
		t.Fatal("readArray should fail with invalid prefix")
	}
}

func TestReadArray_InvalidCount(t *testing.T) {
	input := "*abc\r\n"
	reader := bufio.NewReader(strings.NewReader(input))

	_, err := readArray(reader)
	if err == nil {
		t.Fatal("readArray should fail with invalid count")
	}
}

// Test readBulkString
func TestReadBulkString(t *testing.T) {
	input := "$5\r\nhello\r\n"
	reader := bufio.NewReader(strings.NewReader(input))

	result, err := readBulkString(reader)
	if err != nil {
		t.Fatalf("readBulkString failed: %v", err)
	}

	if string(result) != "hello" {
		t.Fatalf("Expected 'hello', got %s", string(result))
	}
}

func TestReadBulkString_Empty(t *testing.T) {
	input := "$0\r\n\r\n"
	reader := bufio.NewReader(strings.NewReader(input))

	result, err := readBulkString(reader)
	if err != nil {
		t.Fatalf("readBulkString failed: %v", err)
	}

	if len(result) != 0 {
		t.Fatalf("Expected empty string, got %s", string(result))
	}
}

func TestReadBulkString_Negative(t *testing.T) {
	input := "$-1\r\n"
	reader := bufio.NewReader(strings.NewReader(input))

	result, err := readBulkString(reader)
	if err != nil {
		t.Fatalf("readBulkString failed: %v", err)
	}

	if len(result) != 0 {
		t.Fatalf("Expected empty string for negative length, got %s", string(result))
	}
}

func TestReadBulkString_Large(t *testing.T) {
	// Create large string
	largeStr := strings.Repeat("a", 10000)
	input := "$10000\r\n" + largeStr + "\r\n"
	reader := bufio.NewReader(strings.NewReader(input))

	result, err := readBulkString(reader)
	if err != nil {
		t.Fatalf("readBulkString failed: %v", err)
	}

	if string(result) != largeStr {
		t.Fatal("Large string read incorrectly")
	}
}

func TestReadBulkString_WithNewlines(t *testing.T) {
	input := "$12\r\nhello\r\nworld\r\n"
	reader := bufio.NewReader(strings.NewReader(input))

	result, err := readBulkString(reader)
	if err != nil {
		t.Fatalf("readBulkString failed: %v", err)
	}

	if string(result) != "hello\r\nworld" {
		t.Fatalf("Expected 'hello\\r\\nworld', got %s", string(result))
	}
}

func TestReadBulkString_InvalidPrefix(t *testing.T) {
	input := "X5\r\nhello\r\n"
	reader := bufio.NewReader(strings.NewReader(input))

	_, err := readBulkString(reader)
	if err == nil {
		t.Fatal("readBulkString should fail with invalid prefix")
	}
}

func TestReadBulkString_InvalidLength(t *testing.T) {
	input := "$abc\r\n"
	reader := bufio.NewReader(strings.NewReader(input))

	_, err := readBulkString(reader)
	if err == nil {
		t.Fatal("readBulkString should fail with invalid length")
	}
}

func TestReadBulkString_MissingCRLF(t *testing.T) {
	input := "$5\r\nhello"
	reader := bufio.NewReader(strings.NewReader(input))

	_, err := readBulkString(reader)
	if err == nil {
		t.Fatal("readBulkString should fail with missing CRLF")
	}
}

func TestReadBulkString_WrongCRLF(t *testing.T) {
	input := "$5\r\nhello\n\n"
	reader := bufio.NewReader(strings.NewReader(input))

	_, err := readBulkString(reader)
	if err == nil {
		t.Fatal("readBulkString should fail with wrong CRLF")
	}
}

func TestReadBulkString_ShortRead(t *testing.T) {
	input := "$10\r\nshort\r\n"
	reader := bufio.NewReader(strings.NewReader(input))

	_, err := readBulkString(reader)
	if err == nil {
		t.Fatal("readBulkString should fail when data is shorter than length")
	}
}

// Test ReadCommand error cases
func TestReadCommand_EmptyInput(t *testing.T) {
	input := ""
	reader := bufio.NewReader(strings.NewReader(input))

	_, err := ReadCommand(reader)
	if err == nil {
		t.Fatal("ReadCommand should fail with empty input")
	}
}

func TestReadCommand_EmptyLine(t *testing.T) {
	input := "\n"
	reader := bufio.NewReader(strings.NewReader(input))

	_, err := ReadCommand(reader)
	if err == nil {
		t.Fatal("ReadCommand should fail with empty line")
	}
}

func TestReadCommand_WhitespaceOnly(t *testing.T) {
	input := "   \n"
	reader := bufio.NewReader(strings.NewReader(input))

	_, err := ReadCommand(reader)
	if err == nil {
		t.Fatal("ReadCommand should fail with whitespace only")
	}
}

// Test complex scenarios
func TestReadCommand_ComplexArray(t *testing.T) {
	// LPUSH mylist "hello world"
	input := "*3\r\n$5\r\nLPUSH\r\n$6\r\nmylist\r\n$11\r\nhello world\r\n"
	reader := bufio.NewReader(strings.NewReader(input))

	args, err := ReadCommand(reader)
	if err != nil {
		t.Fatalf("ReadCommand failed: %v", err)
	}

	if len(args) != 3 {
		t.Fatalf("Expected 3 args, got %d", len(args))
	}
	if string(args[0]) != "LPUSH" {
		t.Fatalf("Expected 'LPUSH', got %s", string(args[0]))
	}
	if string(args[1]) != "mylist" {
		t.Fatalf("Expected 'mylist', got %s", string(args[1]))
	}
	if string(args[2]) != "hello world" {
		t.Fatalf("Expected 'hello world', got %s", string(args[2]))
	}
}

func TestReadCommand_MultipleCommands(t *testing.T) {
	// Multiple commands in sequence
	input := "*2\r\n$3\r\nGET\r\n$3\r\nkey\r\n*2\r\n$3\r\nSET\r\n$3\r\nkey\r\n"
	reader := bufio.NewReader(strings.NewReader(input))

	// Read first command
	args1, err := ReadCommand(reader)
	if err != nil {
		t.Fatalf("ReadCommand failed: %v", err)
	}
	if len(args1) != 2 || string(args1[0]) != "GET" {
		t.Fatal("First command read incorrectly")
	}

	// Read second command
	args2, err := ReadCommand(reader)
	if err != nil {
		t.Fatalf("ReadCommand failed: %v", err)
	}
	if len(args2) != 2 || string(args2[0]) != "SET" {
		t.Fatal("Second command read incorrectly")
	}
}

// Test binary data handling
func TestReadBulkString_BinaryData(t *testing.T) {
	binaryData := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE}
	var buf bytes.Buffer
	buf.WriteString("$5\r\n")
	buf.Write(binaryData)
	buf.WriteString("\r\n")

	reader := bufio.NewReader(&buf)
	result, err := readBulkString(reader)
	if err != nil {
		t.Fatalf("readBulkString failed: %v", err)
	}

	if !bytes.Equal(result, binaryData) {
		t.Fatal("Binary data read incorrectly")
	}
}
