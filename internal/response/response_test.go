package response

import (
	"strings"
	"testing"
)

func TestFormatResponse(t *testing.T) {
	result := FormatResponse(SimpleStringPrefix, "OK")
	expected := "+OK\r\n"
	if result != expected {
		t.Fatalf("Expected %q, got %q", expected, result)
	}
}

func TestFormatResponse_Error(t *testing.T) {
	result := FormatResponse(ErrorPrefix, "Error message")
	expected := "-ERR Error message\r\n"
	if result != expected {
		t.Fatalf("Expected %q, got %q", expected, result)
	}
}

func TestFormatResponse_Integer(t *testing.T) {
	result := FormatResponse(IntegerPrefix, "42")
	expected := ":42\r\n"
	if result != expected {
		t.Fatalf("Expected %q, got %q", expected, result)
	}
}

func TestFormatResponse_Empty(t *testing.T) {
	result := FormatResponse(SimpleStringPrefix, "")
	expected := "+\r\n"
	if result != expected {
		t.Fatalf("Expected %q, got %q", expected, result)
	}
}

func TestFormatBulkString(t *testing.T) {
	result := FormatBulkString([]byte("hello"))
	expected := "$5\r\nhello\r\n"
	if result != expected {
		t.Fatalf("Expected %q, got %q", expected, result)
	}
}

func TestFormatBulkString_Empty(t *testing.T) {
	result := FormatBulkString([]byte(""))
	expected := "$0\r\n\r\n"
	if result != expected {
		t.Fatalf("Expected %q, got %q", expected, result)
	}
}

func TestFormatBulkString_Nil(t *testing.T) {
	result := FormatBulkString(nil)
	expected := "$-1\r\n"
	if result != expected {
		t.Fatalf("Expected %q, got %q", expected, result)
	}
}

func TestFormatBulkString_Large(t *testing.T) {
	largeStr := strings.Repeat("a", 1000)
	result := FormatBulkString([]byte(largeStr))
	
	expected := "$1000\r\n" + largeStr + "\r\n"
	if result != expected {
		t.Fatal("Large bulk string formatting incorrect")
	}
}

func TestFormatBulkString_WithNewlines(t *testing.T) {
	result := FormatBulkString([]byte("hello\nworld"))
	expected := "$11\r\nhello\nworld\r\n"
	if result != expected {
		t.Fatalf("Expected %q, got %q", expected, result)
	}
}

func TestFormatArray(t *testing.T) {
	values := [][]byte{
		[]byte("hello"),
		[]byte("world"),
	}
	result := FormatArray(values)
	
	// Should start with array prefix and length
	if !strings.HasPrefix(result, "*2\r\n") {
		t.Fatal("Array should start with *2")
	}
	
	// Should contain both bulk strings
	if !strings.Contains(result, "$5\r\nhello\r\n") {
		t.Fatal("Array should contain 'hello'")
	}
	if !strings.Contains(result, "$5\r\nworld\r\n") {
		t.Fatal("Array should contain 'world'")
	}
}

func TestFormatArray_Empty(t *testing.T) {
	result := FormatArray([][]byte{})
	expected := "*0\r\n"
	if result != expected {
		t.Fatalf("Expected %q, got %q", expected, result)
	}
}

func TestFormatArray_SingleElement(t *testing.T) {
	values := [][]byte{[]byte("single")}
	result := FormatArray(values)
	
	if !strings.HasPrefix(result, "*1\r\n") {
		t.Fatal("Array should start with *1")
	}
	if !strings.Contains(result, "$6\r\nsingle\r\n") {
		t.Fatal("Array should contain 'single'")
	}
}

func TestFormatArray_MultipleElements(t *testing.T) {
	values := [][]byte{
		[]byte("a"),
		[]byte("b"),
		[]byte("c"),
	}
	result := FormatArray(values)
	
	if !strings.HasPrefix(result, "*3\r\n") {
		t.Fatal("Array should start with *3")
	}
	
	// Verify all elements are present
	if !strings.Contains(result, "$1\r\na\r\n") {
		t.Fatal("Array should contain 'a'")
	}
	if !strings.Contains(result, "$1\r\nb\r\n") {
		t.Fatal("Array should contain 'b'")
	}
	if !strings.Contains(result, "$1\r\nc\r\n") {
		t.Fatal("Array should contain 'c'")
	}
}

func TestFormatArray_WithNil(t *testing.T) {
	values := [][]byte{
		[]byte("hello"),
		nil,
		[]byte("world"),
	}
	result := FormatArray(values)
	
	if !strings.HasPrefix(result, "*3\r\n") {
		t.Fatal("Array should start with *3")
	}
	
	// Should contain nil bulk string
	if !strings.Contains(result, "$-1\r\n") {
		t.Fatal("Array should contain nil bulk string")
	}
}

func TestErrWrongTypeResponse(t *testing.T) {
	result := ErrWrongTypeResponse()
	expected := "-ERR Wrong type\r\n"
	if result != expected {
		t.Fatalf("Expected %q, got %q", expected, result)
	}
}

func TestErrInternalResponse(t *testing.T) {
	result := ErrInternalResponse()
	expected := "-ERR Internal error\r\n"
	if result != expected {
		t.Fatalf("Expected %q, got %q", expected, result)
	}
}

func TestErrEmptyCommandResponse(t *testing.T) {
	result := ErrEmptyCommandResponse()
	expected := "-ERR Empty command\r\n"
	if result != expected {
		t.Fatalf("Expected %q, got %q", expected, result)
	}
}

func TestErrInvalidIntegerResponse(t *testing.T) {
	result := ErrInvalidIntegerResponse()
	expected := "-ERR Invalid integer\r\n"
	if result != expected {
		t.Fatalf("Expected %q, got %q", expected, result)
	}
}

func TestErrWrongArityResponse(t *testing.T) {
	result := ErrWrongArityResponse()
	expected := "-ERR Wrong number of arguments\r\n"
	if result != expected {
		t.Fatalf("Expected %q, got %q", expected, result)
	}
}

func TestErrUnknownCommandResponse(t *testing.T) {
	result := ErrUnknownCommandResponse()
	expected := "-ERR Unknown command\r\n"
	if result != expected {
		t.Fatalf("Expected %q, got %q", expected, result)
	}
}

func TestResponseConstants(t *testing.T) {
	if SimpleStringPrefix != "+" {
		t.Fatalf("Expected SimpleStringPrefix to be '+', got %q", SimpleStringPrefix)
	}
	if ErrorPrefix != "-ERR " {
		t.Fatalf("Expected ErrorPrefix to be '-ERR ', got %q", ErrorPrefix)
	}
	if BulkStringPrefix != "$" {
		t.Fatalf("Expected BulkStringPrefix to be '$', got %q", BulkStringPrefix)
	}
	if IntegerPrefix != ":" {
		t.Fatalf("Expected IntegerPrefix to be ':', got %q", IntegerPrefix)
	}
	if ArrayPrefix != "*" {
		t.Fatalf("Expected ArrayPrefix to be '*', got %q", ArrayPrefix)
	}
}

func TestFormatBulkString_BinaryData(t *testing.T) {
	binaryData := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE}
	result := FormatBulkString(binaryData)
	
	// Should contain the binary data
	if !strings.Contains(result, string(binaryData)) {
		t.Fatal("Binary data not preserved in bulk string")
	}
	
	// Should have correct length prefix
	if !strings.HasPrefix(result, "$5\r\n") {
		t.Fatal("Bulk string should have correct length prefix")
	}
}

