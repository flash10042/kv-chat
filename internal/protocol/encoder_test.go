package protocol

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

func TestEncodeCommand(t *testing.T) {
	args := [][]byte{
		[]byte("SET"),
		[]byte("key"),
		[]byte("value"),
	}

	result := EncodeCommand(args)

	expected := "*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n"
	if string(result) != expected {
		t.Fatalf("Expected %q, got %q", expected, string(result))
	}
}

func TestEncodeCommand_SingleArg(t *testing.T) {
	args := [][]byte{
		[]byte("PING"),
	}

	result := EncodeCommand(args)

	expected := "*1\r\n$4\r\nPING\r\n"
	if string(result) != expected {
		t.Fatalf("Expected %q, got %q", expected, string(result))
	}
}

func TestEncodeCommand_EmptyArgs(t *testing.T) {
	args := [][]byte{}

	result := EncodeCommand(args)

	expected := "*0\r\n"
	if string(result) != expected {
		t.Fatalf("Expected %q, got %q", expected, string(result))
	}
}

func TestEncodeCommand_EmptyString(t *testing.T) {
	args := [][]byte{
		[]byte("SET"),
		[]byte(""),
		[]byte("value"),
	}

	result := EncodeCommand(args)

	expected := "*3\r\n$3\r\nSET\r\n$0\r\n\r\n$5\r\nvalue\r\n"
	if string(result) != expected {
		t.Fatalf("Expected %q, got %q", expected, string(result))
	}
}

func TestEncodeCommand_MultipleArgs(t *testing.T) {
	args := [][]byte{
		[]byte("LPUSH"),
		[]byte("mylist"),
		[]byte("item1"),
		[]byte("item2"),
		[]byte("item3"),
	}

	result := EncodeCommand(args)

	expected := "*5\r\n$5\r\nLPUSH\r\n$6\r\nmylist\r\n$5\r\nitem1\r\n$5\r\nitem2\r\n$5\r\nitem3\r\n"
	if string(result) != expected {
		t.Fatalf("Expected %q, got %q", expected, string(result))
	}
}

func TestEncodeCommand_LargeArgs(t *testing.T) {
	largeValue := make([]byte, 1000)
	for i := range largeValue {
		largeValue[i] = byte('a' + (i % 26))
	}

	args := [][]byte{
		[]byte("SET"),
		[]byte("key"),
		largeValue,
	}

	result := EncodeCommand(args)

	// Verify it starts correctly
	if !bytes.HasPrefix(result, []byte("*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$1000\r\n")) {
		t.Fatal("Encoding format incorrect for large value")
	}

	// Verify the large value is included
	if !bytes.Contains(result, largeValue) {
		t.Fatal("Large value not included in encoding")
	}
}

func TestEncodeCommand_BinaryData(t *testing.T) {
	binaryData := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE}
	args := [][]byte{
		[]byte("SET"),
		[]byte("key"),
		binaryData,
	}

	result := EncodeCommand(args)

	// Verify binary data is preserved
	if !bytes.Contains(result, binaryData) {
		t.Fatal("Binary data not preserved in encoding")
	}
}

func TestEncodeCommand_WithSpaces(t *testing.T) {
	args := [][]byte{
		[]byte("SET"),
		[]byte("my key"),
		[]byte("my value"),
	}

	result := EncodeCommand(args)

	expected := "*3\r\n$3\r\nSET\r\n$6\r\nmy key\r\n$8\r\nmy value\r\n"
	if string(result) != expected {
		t.Fatalf("Expected %q, got %q", expected, string(result))
	}
}

func TestEncodeCommand_WithNewlines(t *testing.T) {
	args := [][]byte{
		[]byte("SET"),
		[]byte("key"),
		[]byte("value\nwith\nnewlines"),
	}

	result := EncodeCommand(args)

	// Verify newlines are preserved
	if !bytes.Contains(result, []byte("value\nwith\nnewlines")) {
		t.Fatal("Newlines not preserved in encoding")
	}
}

func TestEncodeCommand_RoundTrip(t *testing.T) {
	originalArgs := [][]byte{
		[]byte("SET"),
		[]byte("key"),
		[]byte("value"),
	}

	encoded := EncodeCommand(originalArgs)

	// Decode using ReadCommand
	reader := bufio.NewReader(strings.NewReader(string(encoded)))
	decodedArgs, err := ReadCommand(reader)
	if err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}

	if len(decodedArgs) != len(originalArgs) {
		t.Fatalf("Length mismatch: expected %d, got %d", len(originalArgs), len(decodedArgs))
	}

	for i := range originalArgs {
		if !bytes.Equal(originalArgs[i], decodedArgs[i]) {
			t.Fatalf("Arg %d mismatch: expected %q, got %q", i, originalArgs[i], decodedArgs[i])
		}
	}
}

