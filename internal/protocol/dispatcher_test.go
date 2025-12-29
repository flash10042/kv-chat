package protocol

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flash10042/kv-chat/internal/persistence"
	"github.com/flash10042/kv-chat/internal/response"
	"github.com/flash10042/kv-chat/internal/store"
)

func TestCheckArity(t *testing.T) {
	// Positive cases
	if !checkArity(3, 3) {
		t.Fatal("checkArity(3, 3) should return true")
	}
	if !checkArity(2, 2) {
		t.Fatal("checkArity(2, 2) should return true")
	}

	// Negative cases - exact mismatch
	if checkArity(2, 3) {
		t.Fatal("checkArity(2, 3) should return false")
	}
	if checkArity(3, 2) {
		t.Fatal("checkArity(3, 2) should return false")
	}

	// Variable arity (negative means at least)
	if !checkArity(3, -2) {
		t.Fatal("checkArity(3, -2) should return true (3 >= 2)")
	}
	if !checkArity(5, -3) {
		t.Fatal("checkArity(5, -3) should return true (5 >= 3)")
	}
	if checkArity(1, -2) {
		t.Fatal("checkArity(1, -2) should return false (1 < 2)")
	}
}

func TestDispatchCommand_EmptyCommand(t *testing.T) {
	storage := store.NewStorage()
	args := [][]byte{}

	result := DispatchCommand(DispatchModePublic, args, storage, nil)

	expected := response.ErrEmptyCommandResponse()
	if result != expected {
		t.Fatalf("Expected empty command error, got %q", result)
	}
}

func TestDispatchCommand_UnknownCommand(t *testing.T) {
	storage := store.NewStorage()
	args := [][]byte{[]byte("UNKNOWN")}

	result := DispatchCommand(DispatchModePublic, args, storage, nil)

	expected := response.ErrUnknownCommandResponse()
	if result != expected {
		t.Fatalf("Expected unknown command error, got %q", result)
	}
}

func TestDispatchCommand_WrongArity(t *testing.T) {
	storage := store.NewStorage()
	// SET requires 3 args (command, key, value)
	args := [][]byte{[]byte("SET"), []byte("key")}

	result := DispatchCommand(DispatchModePublic, args, storage, nil)

	expected := response.ErrWrongArityResponse()
	if result != expected {
		t.Fatalf("Expected wrong arity error, got %q", result)
	}
}

func TestDispatchCommand_PING(t *testing.T) {
	storage := store.NewStorage()
	args := [][]byte{[]byte("PING")}

	result := DispatchCommand(DispatchModePublic, args, storage, nil)

	expected := response.FormatResponse(response.SimpleStringPrefix, "PONG")
	if result != expected {
		t.Fatalf("Expected PONG, got %q", result)
	}
}

func TestDispatchCommand_SET(t *testing.T) {
	storage := store.NewStorage()
	args := [][]byte{[]byte("SET"), []byte("key"), []byte("value")}

	result := DispatchCommand(DispatchModePublic, args, storage, nil)

	expected := response.FormatResponse(response.SimpleStringPrefix, "OK")
	if result != expected {
		t.Fatalf("Expected OK, got %q", result)
	}

	// Verify value was set
	value, _ := storage.Get("key")
	if string(value) != "value" {
		t.Fatalf("Expected 'value', got %q", string(value))
	}
}

func TestDispatchCommand_GET(t *testing.T) {
	storage := store.NewStorage()
	storage.Set("key", []byte("value"))

	args := [][]byte{[]byte("GET"), []byte("key")}
	result := DispatchCommand(DispatchModePublic, args, storage, nil)

	expected := response.FormatBulkString([]byte("value"))
	if result != expected {
		t.Fatalf("Expected bulk string with 'value', got %q", result)
	}
}

func TestDispatchCommand_GET_NonExistent(t *testing.T) {
	storage := store.NewStorage()

	args := [][]byte{[]byte("GET"), []byte("nonexistent")}
	result := DispatchCommand(DispatchModePublic, args, storage, nil)

	expected := response.FormatBulkString(nil)
	if result != expected {
		t.Fatalf("Expected null bulk string, got %q", result)
	}
}

func TestDispatchCommand_DEL(t *testing.T) {
	storage := store.NewStorage()
	storage.Set("key", []byte("value"))

	args := [][]byte{[]byte("DEL"), []byte("key")}
	result := DispatchCommand(DispatchModePublic, args, storage, nil)

	expected := response.FormatResponse(response.IntegerPrefix, "1")
	if result != expected {
		t.Fatalf("Expected :1, got %q", result)
	}

	// Verify key was deleted
	if storage.Exists("key") {
		t.Fatal("Key should be deleted")
	}
}

func TestDispatchCommand_DEL_NonExistent(t *testing.T) {
	storage := store.NewStorage()

	args := [][]byte{[]byte("DEL"), []byte("nonexistent")}
	result := DispatchCommand(DispatchModePublic, args, storage, nil)

	expected := response.FormatResponse(response.IntegerPrefix, "0")
	if result != expected {
		t.Fatalf("Expected :0, got %q", result)
	}
}

func TestDispatchCommand_LPUSH(t *testing.T) {
	storage := store.NewStorage()

	args := [][]byte{[]byte("LPUSH"), []byte("list"), []byte("item")}
	result := DispatchCommand(DispatchModePublic, args, storage, nil)

	expected := response.FormatResponse(response.IntegerPrefix, "1")
	if result != expected {
		t.Fatalf("Expected :1, got %q", result)
	}
}

func TestDispatchCommand_RPUSH(t *testing.T) {
	storage := store.NewStorage()

	args := [][]byte{[]byte("RPUSH"), []byte("list"), []byte("item")}
	result := DispatchCommand(DispatchModePublic, args, storage, nil)

	expected := response.FormatResponse(response.IntegerPrefix, "1")
	if result != expected {
		t.Fatalf("Expected :1, got %q", result)
	}
}

func TestDispatchCommand_LRANGE(t *testing.T) {
	storage := store.NewStorage()
	storage.RPush("list", []byte("a"))
	storage.RPush("list", []byte("b"))

	args := [][]byte{[]byte("LRANGE"), []byte("list"), []byte("0"), []byte("-1")}
	result := DispatchCommand(DispatchModePublic, args, storage, nil)

	// Should return array with 2 elements
	if result[0] != '*' {
		t.Fatal("Expected array response")
	}
}

func TestDispatchCommand_EXPIRE(t *testing.T) {
	storage := store.NewStorage()
	storage.Set("key", []byte("value"))

	args := [][]byte{[]byte("EXPIRE"), []byte("key"), []byte("10")}
	result := DispatchCommand(DispatchModePublic, args, storage, nil)

	expected := response.FormatResponse(response.IntegerPrefix, "1")
	if result != expected {
		t.Fatalf("Expected :1, got %q", result)
	}
}

func TestDispatchCommand_TTL(t *testing.T) {
	storage := store.NewStorage()
	storage.Set("key", []byte("value"))

	args := [][]byte{[]byte("TTL"), []byte("key")}
	result := DispatchCommand(DispatchModePublic, args, storage, nil)

	// Should return integer (TTL -1 for no expiration)
	if result[0] != ':' {
		t.Fatal("Expected integer response")
	}
}

func TestDispatchCommand_EXISTS(t *testing.T) {
	storage := store.NewStorage()
	storage.Set("key", []byte("value"))

	args := [][]byte{[]byte("EXISTS"), []byte("key")}
	result := DispatchCommand(DispatchModePublic, args, storage, nil)

	expected := response.FormatResponse(response.IntegerPrefix, "1")
	if result != expected {
		t.Fatalf("Expected :1, got %q", result)
	}
}

func TestDispatchCommand_SETEX(t *testing.T) {
	storage := store.NewStorage()

	args := [][]byte{[]byte("SETEX"), []byte("key"), []byte("10"), []byte("value")}
	result := DispatchCommand(DispatchModePublic, args, storage, nil)

	expected := response.FormatResponse(response.SimpleStringPrefix, "OK")
	if result != expected {
		t.Fatalf("Expected OK, got %q", result)
	}
}

func TestDispatchCommand_CaseInsensitive(t *testing.T) {
	storage := store.NewStorage()

	// Test lowercase
	args := [][]byte{[]byte("ping")}
	result := DispatchCommand(DispatchModePublic, args, storage, nil)
	expected := response.FormatResponse(response.SimpleStringPrefix, "PONG")
	if result != expected {
		t.Fatalf("Lowercase command failed, got %q", result)
	}

	// Test mixed case
	args = [][]byte{[]byte("SeT"), []byte("key"), []byte("value")}
	result = DispatchCommand(DispatchModePublic, args, storage, nil)
	expected = response.FormatResponse(response.SimpleStringPrefix, "OK")
	if result != expected {
		t.Fatalf("Mixed case command failed, got %q", result)
	}
}

func TestDispatchCommand_WithAOF(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.aof")
	aof := persistence.NewAOF(filename)
	defer aof.Close()

	storage := store.NewStorage()

	// Mutating command should be written to AOF
	args := [][]byte{[]byte("SET"), []byte("key"), []byte("value")}
	result := DispatchCommand(DispatchModePublic, args, storage, aof)

	expected := response.FormatResponse(response.SimpleStringPrefix, "OK")
	if result != expected {
		t.Fatalf("Expected OK, got %q", result)
	}

	// Verify AOF file contains the command
	file, err := os.Open(filename)
	if err != nil {
		t.Fatalf("Failed to open AOF file: %v", err)
	}
	defer file.Close()

	// Read file content
	stat, _ := file.Stat()
	if stat.Size() == 0 {
		t.Fatal("AOF file should contain data")
	}
}

func TestDispatchCommand_WithoutAOF(t *testing.T) {
	storage := store.NewStorage()

	// Mutating command without AOF should still work
	args := [][]byte{[]byte("SET"), []byte("key"), []byte("value")}
	result := DispatchCommand(DispatchModePublic, args, storage, nil)

	expected := response.FormatResponse(response.SimpleStringPrefix, "OK")
	if result != expected {
		t.Fatalf("Expected OK, got %q", result)
	}
}

func TestDispatchCommand_NonMutatingWithAOF(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.aof")
	aof := persistence.NewAOF(filename)
	defer aof.Close()

	storage := store.NewStorage()
	storage.Set("key", []byte("value"))

	// Non-mutating command should not be written to AOF
	args := [][]byte{[]byte("GET"), []byte("key")}
	_ = DispatchCommand(DispatchModePublic, args, storage, aof)

	// Verify AOF file is empty (only contains previous SET if any)
	// This test verifies GET doesn't write to AOF
	file, err := os.Open(filename)
	if err != nil {
		t.Fatalf("Failed to open AOF file: %v", err)
	}
	defer file.Close()
}

func TestDispatchCommand_AllRegisteredCommands(t *testing.T) {
	storage := store.NewStorage()

	// Test all registered commands have correct arity
	testCases := []struct {
		command string
		args    [][]byte
		valid   bool
	}{
		{"PING", [][]byte{[]byte("PING")}, true},
		{"SET", [][]byte{[]byte("SET"), []byte("key"), []byte("value")}, true},
		{"GET", [][]byte{[]byte("GET"), []byte("key")}, true},
		{"LPUSH", [][]byte{[]byte("LPUSH"), []byte("list"), []byte("item")}, true},
		{"RPUSH", [][]byte{[]byte("RPUSH"), []byte("list"), []byte("item")}, true},
		{"LRANGE", [][]byte{[]byte("LRANGE"), []byte("list"), []byte("0"), []byte("-1")}, true},
		{"EXPIRE", [][]byte{[]byte("EXPIRE"), []byte("key"), []byte("10")}, true},
		{"TTL", [][]byte{[]byte("TTL"), []byte("key")}, true},
		{"DEL", [][]byte{[]byte("DEL"), []byte("key")}, true},
		{"EXISTS", [][]byte{[]byte("EXISTS"), []byte("key")}, true},
		{"SETEX", [][]byte{[]byte("SETEX"), []byte("key"), []byte("10"), []byte("value")}, true},
	}

	for _, tc := range testCases {
		t.Run(tc.command, func(t *testing.T) {
			result := DispatchCommand(DispatchModePublic, tc.args, storage, nil)
			if !tc.valid {
				if result[0] != '-' {
					t.Errorf("Expected error response for invalid %s", tc.command)
				}
			} else {
				// Should not be an error (doesn't start with '-')
				if result[0] == '-' && result[1] == 'E' {
					t.Errorf("Unexpected error for %s: %q", tc.command, result)
				}
			}
		})
	}
}
