package commands

import (
	"strings"
	"testing"

	"github.com/flash10042/kv-chat/internal/response"
	"github.com/flash10042/kv-chat/internal/store"
)

func TestPingHandler(t *testing.T) {
	storage := store.NewStorage()
	args := [][]byte{[]byte("PING")}

	result := PingHandler(args, storage)

	expected := response.FormatResponse(response.SimpleStringPrefix, "PONG")
	if result != expected {
		t.Fatalf("Expected PONG, got %q", result)
	}
}

func TestSetHandler(t *testing.T) {
	storage := store.NewStorage()
	args := [][]byte{[]byte("SET"), []byte("key"), []byte("value")}

	result := SetHandler(args, storage)

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

func TestSetHandler_Overwrite(t *testing.T) {
	storage := store.NewStorage()
	storage.Set("key", []byte("old"))

	args := [][]byte{[]byte("SET"), []byte("key"), []byte("new")}
	result := SetHandler(args, storage)

	if result != response.FormatResponse(response.SimpleStringPrefix, "OK") {
		t.Fatal("Set should return OK")
	}

	value, _ := storage.Get("key")
	if string(value) != "new" {
		t.Fatalf("Expected 'new', got %q", string(value))
	}
}

func TestGetHandler(t *testing.T) {
	storage := store.NewStorage()
	storage.Set("key", []byte("value"))

	args := [][]byte{[]byte("GET"), []byte("key")}
	result := GetHandler(args, storage)

	expected := response.FormatBulkString([]byte("value"))
	if result != expected {
		t.Fatalf("Expected bulk string with 'value', got %q", result)
	}
}

func TestGetHandler_NonExistent(t *testing.T) {
	storage := store.NewStorage()

	args := [][]byte{[]byte("GET"), []byte("nonexistent")}
	result := GetHandler(args, storage)

	expected := response.FormatBulkString(nil)
	if result != expected {
		t.Fatalf("Expected null bulk string, got %q", result)
	}
}

func TestGetHandler_WrongType(t *testing.T) {
	storage := store.NewStorage()
	storage.LPush("key", []byte("item"))

	args := [][]byte{[]byte("GET"), []byte("key")}
	result := GetHandler(args, storage)

	expected := response.ErrWrongTypeResponse()
	if result != expected {
		t.Fatalf("Expected wrong type error, got %q", result)
	}
}

func TestLPushHandler(t *testing.T) {
	storage := store.NewStorage()

	args := [][]byte{[]byte("LPUSH"), []byte("list"), []byte("item")}
	result := LPushHandler(args, storage)

	expected := response.FormatResponse(response.IntegerPrefix, "1")
	if result != expected {
		t.Fatalf("Expected :1, got %q", result)
	}

	// Verify item was added
	values, _ := storage.LRange("list", 0, -1)
	if len(values) != 1 || string(values[0]) != "item" {
		t.Fatal("Item should be in list")
	}
}

func TestLPushHandler_Multiple(t *testing.T) {
	storage := store.NewStorage()

	args1 := [][]byte{[]byte("LPUSH"), []byte("list"), []byte("first")}
	args2 := [][]byte{[]byte("LPUSH"), []byte("list"), []byte("second")}

	LPushHandler(args1, storage)
	result := LPushHandler(args2, storage)

	expected := response.FormatResponse(response.IntegerPrefix, "2")
	if result != expected {
		t.Fatalf("Expected :2, got %q", result)
	}

	// Verify order (LPush adds to front)
	values, _ := storage.LRange("list", 0, -1)
	if len(values) != 2 || string(values[0]) != "second" {
		t.Fatal("Items should be in correct order")
	}
}

func TestLPushHandler_WrongType(t *testing.T) {
	storage := store.NewStorage()
	storage.Set("key", []byte("value"))

	args := [][]byte{[]byte("LPUSH"), []byte("key"), []byte("item")}
	result := LPushHandler(args, storage)

	expected := response.ErrWrongTypeResponse()
	if result != expected {
		t.Fatalf("Expected wrong type error, got %q", result)
	}
}

func TestRPushHandler(t *testing.T) {
	storage := store.NewStorage()

	args := [][]byte{[]byte("RPUSH"), []byte("list"), []byte("item")}
	result := RPushHandler(args, storage)

	expected := response.FormatResponse(response.IntegerPrefix, "1")
	if result != expected {
		t.Fatalf("Expected :1, got %q", result)
	}
}

func TestRPushHandler_Multiple(t *testing.T) {
	storage := store.NewStorage()

	args1 := [][]byte{[]byte("RPUSH"), []byte("list"), []byte("first")}
	args2 := [][]byte{[]byte("RPUSH"), []byte("list"), []byte("second")}

	RPushHandler(args1, storage)
	result := RPushHandler(args2, storage)

	expected := response.FormatResponse(response.IntegerPrefix, "2")
	if result != expected {
		t.Fatalf("Expected :2, got %q", result)
	}

	// Verify order (RPush adds to end)
	values, _ := storage.LRange("list", 0, -1)
	if len(values) != 2 || string(values[0]) != "first" {
		t.Fatal("Items should be in correct order")
	}
}

func TestRPushHandler_WrongType(t *testing.T) {
	storage := store.NewStorage()
	storage.Set("key", []byte("value"))

	args := [][]byte{[]byte("RPUSH"), []byte("key"), []byte("item")}
	result := RPushHandler(args, storage)

	expected := response.ErrWrongTypeResponse()
	if result != expected {
		t.Fatalf("Expected wrong type error, got %q", result)
	}
}

func TestLRangeHandler(t *testing.T) {
	storage := store.NewStorage()
	storage.RPush("list", []byte("a"))
	storage.RPush("list", []byte("b"))
	storage.RPush("list", []byte("c"))

	args := [][]byte{[]byte("LRANGE"), []byte("list"), []byte("0"), []byte("-1")}
	result := LRangeHandler(args, storage)

	// Should return array
	if result[0] != '*' {
		t.Fatal("Expected array response")
	}
}

func TestLRangeHandler_InvalidStart(t *testing.T) {
	storage := store.NewStorage()
	storage.RPush("list", []byte("a"))

	args := [][]byte{[]byte("LRANGE"), []byte("list"), []byte("invalid"), []byte("0")}
	result := LRangeHandler(args, storage)

	expected := response.ErrInvalidIntegerResponse()
	if result != expected {
		t.Fatalf("Expected invalid integer error, got %q", result)
	}
}

func TestLRangeHandler_InvalidEnd(t *testing.T) {
	storage := store.NewStorage()
	storage.RPush("list", []byte("a"))

	args := [][]byte{[]byte("LRANGE"), []byte("list"), []byte("0"), []byte("invalid")}
	result := LRangeHandler(args, storage)

	expected := response.ErrInvalidIntegerResponse()
	if result != expected {
		t.Fatalf("Expected invalid integer error, got %q", result)
	}
}

func TestLRangeHandler_WrongType(t *testing.T) {
	storage := store.NewStorage()
	storage.Set("key", []byte("value"))

	args := [][]byte{[]byte("LRANGE"), []byte("key"), []byte("0"), []byte("-1")}
	result := LRangeHandler(args, storage)

	expected := response.ErrWrongTypeResponse()
	if result != expected {
		t.Fatalf("Expected wrong type error, got %q", result)
	}
}

func TestLRangeHandler_NonExistent(t *testing.T) {
	storage := store.NewStorage()

	args := [][]byte{[]byte("LRANGE"), []byte("nonexistent"), []byte("0"), []byte("-1")}
	result := LRangeHandler(args, storage)

	// Should return empty array
	if !strings.HasPrefix(result, "*0\r\n") {
		t.Fatalf("Expected empty array, got %q", result)
	}
}

func TestExpireHandler(t *testing.T) {
	storage := store.NewStorage()
	storage.Set("key", []byte("value"))

	args := [][]byte{[]byte("EXPIRE"), []byte("key"), []byte("10")}
	result := ExpireHandler(args, storage)

	expected := response.FormatResponse(response.IntegerPrefix, "1")
	if result != expected {
		t.Fatalf("Expected :1, got %q", result)
	}
}

func TestExpireHandler_NonExistent(t *testing.T) {
	storage := store.NewStorage()

	args := [][]byte{[]byte("EXPIRE"), []byte("nonexistent"), []byte("10")}
	result := ExpireHandler(args, storage)

	expected := response.FormatResponse(response.IntegerPrefix, "0")
	if result != expected {
		t.Fatalf("Expected :0, got %q", result)
	}
}

func TestExpireHandler_InvalidInteger(t *testing.T) {
	storage := store.NewStorage()
	storage.Set("key", []byte("value"))

	args := [][]byte{[]byte("EXPIRE"), []byte("key"), []byte("invalid")}
	result := ExpireHandler(args, storage)

	expected := response.ErrInvalidIntegerResponse()
	if result != expected {
		t.Fatalf("Expected invalid integer error, got %q", result)
	}
}

func TestTTLHandler(t *testing.T) {
	storage := store.NewStorage()
	storage.Set("key", []byte("value"))

	args := [][]byte{[]byte("TTL"), []byte("key")}
	result := TTLHandler(args, storage)

	// Should return integer (TTL -1 for no expiration)
	expected := response.FormatResponse(response.IntegerPrefix, "-1")
	if result != expected {
		t.Fatalf("Expected :-1, got %q", result)
	}
}

func TestTTLHandler_WithExpiration(t *testing.T) {
	storage := store.NewStorage()
	storage.SetEx("key", 10, []byte("value"))

	args := [][]byte{[]byte("TTL"), []byte("key")}
	result := TTLHandler(args, storage)

	// Should return positive integer
	if result[0] != ':' {
		t.Fatal("Expected integer response")
	}
}

func TestTTLHandler_NonExistent(t *testing.T) {
	storage := store.NewStorage()

	args := [][]byte{[]byte("TTL"), []byte("nonexistent")}
	result := TTLHandler(args, storage)

	expected := response.FormatResponse(response.IntegerPrefix, "-2")
	if result != expected {
		t.Fatalf("Expected :-2, got %q", result)
	}
}

func TestDelHandler(t *testing.T) {
	storage := store.NewStorage()
	storage.Set("key", []byte("value"))

	args := [][]byte{[]byte("DEL"), []byte("key")}
	result := DelHandler(args, storage)

	expected := response.FormatResponse(response.IntegerPrefix, "1")
	if result != expected {
		t.Fatalf("Expected :1, got %q", result)
	}

	// Verify key was deleted
	if storage.Exists("key") {
		t.Fatal("Key should be deleted")
	}
}

func TestDelHandler_NonExistent(t *testing.T) {
	storage := store.NewStorage()

	args := [][]byte{[]byte("DEL"), []byte("nonexistent")}
	result := DelHandler(args, storage)

	expected := response.FormatResponse(response.IntegerPrefix, "0")
	if result != expected {
		t.Fatalf("Expected :0, got %q", result)
	}
}

func TestExistsHandler(t *testing.T) {
	storage := store.NewStorage()
	storage.Set("key", []byte("value"))

	args := [][]byte{[]byte("EXISTS"), []byte("key")}
	result := ExistsHandler(args, storage)

	expected := response.FormatResponse(response.IntegerPrefix, "1")
	if result != expected {
		t.Fatalf("Expected :1, got %q", result)
	}
}

func TestExistsHandler_NonExistent(t *testing.T) {
	storage := store.NewStorage()

	args := [][]byte{[]byte("EXISTS"), []byte("nonexistent")}
	result := ExistsHandler(args, storage)

	expected := response.FormatResponse(response.IntegerPrefix, "0")
	if result != expected {
		t.Fatalf("Expected :0, got %q", result)
	}
}

func TestSetExHandler(t *testing.T) {
	storage := store.NewStorage()

	args := [][]byte{[]byte("SETEX"), []byte("key"), []byte("10"), []byte("value")}
	result := SetExHandler(args, storage)

	expected := response.FormatResponse(response.SimpleStringPrefix, "OK")
	if result != expected {
		t.Fatalf("Expected OK, got %q", result)
	}

	// Verify value was set
	value, _ := storage.Get("key")
	if string(value) != "value" {
		t.Fatalf("Expected 'value', got %q", string(value))
	}

	// Verify expiration was set
	ttl := storage.TTL("key")
	if ttl < 1 || ttl > 10 {
		t.Fatalf("Expected TTL between 1 and 10, got %d", ttl)
	}
}

func TestSetExHandler_InvalidInteger(t *testing.T) {
	storage := store.NewStorage()

	args := [][]byte{[]byte("SETEX"), []byte("key"), []byte("invalid"), []byte("value")}
	result := SetExHandler(args, storage)

	expected := response.ErrInvalidIntegerResponse()
	if result != expected {
		t.Fatalf("Expected invalid integer error, got %q", result)
	}
}

func TestSetExHandler_ZeroSeconds(t *testing.T) {
	storage := store.NewStorage()

	args := [][]byte{[]byte("SETEX"), []byte("key"), []byte("0"), []byte("value")}
	SetExHandler(args, storage)

	// Key should not exist
	if storage.Exists("key") {
		t.Fatal("Key should not exist with 0 seconds")
	}
}

