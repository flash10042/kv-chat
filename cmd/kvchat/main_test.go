package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flash10042/kv-chat/internal/protocol"
	"github.com/flash10042/kv-chat/internal/store"
)

func TestReplayAOF_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.aof")

	// Create empty file
	file, err := os.Create(filename)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	file.Close()

	storage := store.NewStorage()
	err = replayAOF(storage, filename)
	if err != nil {
		t.Fatalf("replayAOF should not error on empty file: %v", err)
	}

	// Storage should be empty
	if storage.Exists("anykey") {
		t.Fatal("Storage should be empty after replaying empty AOF")
	}
}

func TestReplayAOF_NonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "nonexistent.aof")

	storage := store.NewStorage()
	err := replayAOF(storage, filename)
	if err != nil {
		t.Fatalf("replayAOF should not error on non-existent file: %v", err)
	}
}

func TestReplayAOF_SetCommand(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.aof")

	// Create AOF file with SET command
	file, err := os.Create(filename)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	args := [][]byte{[]byte("SET"), []byte("key1"), []byte("value1")}
	encoded := protocol.EncodeCommand(args)
	_, err = file.Write(encoded)
	if err != nil {
		t.Fatalf("Failed to write to file: %v", err)
	}
	file.Close()

	storage := store.NewStorage()
	err = replayAOF(storage, filename)
	if err != nil {
		t.Fatalf("replayAOF failed: %v", err)
	}

	// Verify value was set
	value, err := storage.Get("key1")
	if err != nil {
		t.Fatalf("Failed to get key1: %v", err)
	}
	if string(value) != "value1" {
		t.Fatalf("Expected 'value1', got %q", string(value))
	}
}

func TestReplayAOF_MultipleSetCommands(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.aof")

	// Create AOF file with multiple SET commands
	file, err := os.Create(filename)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	commands := [][][]byte{
		{[]byte("SET"), []byte("key1"), []byte("value1")},
		{[]byte("SET"), []byte("key2"), []byte("value2")},
		{[]byte("SET"), []byte("key3"), []byte("value3")},
	}

	for _, args := range commands {
		encoded := protocol.EncodeCommand(args)
		_, err = file.Write(encoded)
		if err != nil {
			t.Fatalf("Failed to write to file: %v", err)
		}
	}
	file.Close()

	storage := store.NewStorage()
	err = replayAOF(storage, filename)
	if err != nil {
		t.Fatalf("replayAOF failed: %v", err)
	}

	// Verify all values were set
	testCases := []struct {
		key   string
		value string
	}{
		{"key1", "value1"},
		{"key2", "value2"},
		{"key3", "value3"},
	}

	for _, tc := range testCases {
		value, err := storage.Get(tc.key)
		if err != nil {
			t.Fatalf("Failed to get %s: %v", tc.key, err)
		}
		if string(value) != tc.value {
			t.Fatalf("Expected %q for %s, got %q", tc.value, tc.key, string(value))
		}
	}
}

func TestReplayAOF_OverwriteKey(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.aof")

	// Create AOF file with SET commands that overwrite
	file, err := os.Create(filename)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	commands := [][][]byte{
		{[]byte("SET"), []byte("key1"), []byte("value1")},
		{[]byte("SET"), []byte("key1"), []byte("value2")},
	}

	for _, args := range commands {
		encoded := protocol.EncodeCommand(args)
		_, err = file.Write(encoded)
		if err != nil {
			t.Fatalf("Failed to write to file: %v", err)
		}
	}
	file.Close()

	storage := store.NewStorage()
	err = replayAOF(storage, filename)
	if err != nil {
		t.Fatalf("replayAOF failed: %v", err)
	}

	// Verify final value
	value, err := storage.Get("key1")
	if err != nil {
		t.Fatalf("Failed to get key1: %v", err)
	}
	if string(value) != "value2" {
		t.Fatalf("Expected 'value2' (overwritten), got %q", string(value))
	}
}

func TestReplayAOF_ListCommands(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.aof")

	// Create AOF file with list commands
	file, err := os.Create(filename)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	commands := [][][]byte{
		{[]byte("LPUSH"), []byte("list1"), []byte("item1")},
		{[]byte("LPUSH"), []byte("list1"), []byte("item2")},
		{[]byte("RPUSH"), []byte("list1"), []byte("item3")},
	}

	for _, args := range commands {
		encoded := protocol.EncodeCommand(args)
		_, err = file.Write(encoded)
		if err != nil {
			t.Fatalf("Failed to write to file: %v", err)
		}
	}
	file.Close()

	storage := store.NewStorage()
	err = replayAOF(storage, filename)
	if err != nil {
		t.Fatalf("replayAOF failed: %v", err)
	}

	// Verify list contents
	values, err := storage.LRange("list1", 0, -1)
	if err != nil {
		t.Fatalf("Failed to get list1: %v", err)
	}
	if len(values) != 3 {
		t.Fatalf("Expected 3 items, got %d", len(values))
	}
	// LPUSH adds to front, so order should be: item2, item1, item3
	expected := []string{"item2", "item1", "item3"}
	for i, v := range values {
		if string(v) != expected[i] {
			t.Fatalf("Expected %q at index %d, got %q", expected[i], i, string(v))
		}
	}
}

func TestReplayAOF_DelCommand(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.aof")

	// Create AOF file with SET and DEL commands
	file, err := os.Create(filename)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	commands := [][][]byte{
		{[]byte("SET"), []byte("key1"), []byte("value1")},
		{[]byte("SET"), []byte("key2"), []byte("value2")},
		{[]byte("DEL"), []byte("key1")},
	}

	for _, args := range commands {
		encoded := protocol.EncodeCommand(args)
		_, err = file.Write(encoded)
		if err != nil {
			t.Fatalf("Failed to write to file: %v", err)
		}
	}
	file.Close()

	storage := store.NewStorage()
	err = replayAOF(storage, filename)
	if err != nil {
		t.Fatalf("replayAOF failed: %v", err)
	}

	// Verify key1 was deleted
	if storage.Exists("key1") {
		t.Fatal("key1 should not exist after DEL command")
	}

	// Verify key2 still exists
	if !storage.Exists("key2") {
		t.Fatal("key2 should still exist")
	}
}

func TestReplayAOF_MixedCommands(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.aof")

	// Create AOF file with mixed commands
	file, err := os.Create(filename)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	commands := [][][]byte{
		{[]byte("SET"), []byte("str1"), []byte("value1")},
		{[]byte("LPUSH"), []byte("list1"), []byte("item1")},
		{[]byte("SET"), []byte("str2"), []byte("value2")},
		{[]byte("RPUSH"), []byte("list1"), []byte("item2")},
		{[]byte("DEL"), []byte("str1")},
	}

	for _, args := range commands {
		encoded := protocol.EncodeCommand(args)
		_, err = file.Write(encoded)
		if err != nil {
			t.Fatalf("Failed to write to file: %v", err)
		}
	}
	file.Close()

	storage := store.NewStorage()
	err = replayAOF(storage, filename)
	if err != nil {
		t.Fatalf("replayAOF failed: %v", err)
	}

	// Verify final state
	if storage.Exists("str1") {
		t.Fatal("str1 should not exist after DEL")
	}

	value, err := storage.Get("str2")
	if err != nil {
		t.Fatalf("Failed to get str2: %v", err)
	}
	if string(value) != "value2" {
		t.Fatalf("Expected 'value2', got %q", string(value))
	}

	values, err := storage.LRange("list1", 0, -1)
	if err != nil {
		t.Fatalf("Failed to get list1: %v", err)
	}
	if len(values) != 2 {
		t.Fatalf("Expected 2 items in list1, got %d", len(values))
	}
}

func TestReplayAOF_PrivateCommands(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.aof")

	// Create AOF file with private commands (EXPIREAT, SETEXAT)
	// These should be replayable since replayAOF uses DispatchModePrivate
	file, err := os.Create(filename)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	commands := [][][]byte{
		{[]byte("SET"), []byte("key1"), []byte("value1")},
		{[]byte("EXPIREAT"), []byte("key1"), []byte("9999999999")}, // Far future timestamp
	}

	for _, args := range commands {
		encoded := protocol.EncodeCommand(args)
		_, err = file.Write(encoded)
		if err != nil {
			t.Fatalf("Failed to write to file: %v", err)
		}
	}
	file.Close()

	storage := store.NewStorage()
	err = replayAOF(storage, filename)
	if err != nil {
		t.Fatalf("replayAOF failed: %v", err)
	}

	// Verify key exists and has expiration set
	if !storage.Exists("key1") {
		t.Fatal("key1 should exist")
	}

	ttl := storage.TTL("key1")
	if ttl <= 0 {
		t.Fatalf("Expected positive TTL, got %d", ttl)
	}
}

func TestReplayAOF_InvalidCommand(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.aof")

	// Create AOF file with invalid command (unknown command name)
	// This will be parsed as inline command but will fail dispatch
	file, err := os.Create(filename)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// Write invalid command as inline format
	// ReadCommand will parse this successfully as inline, but DispatchCommand will reject it
	_, err = file.WriteString("INVALIDCOMMAND key value\n")
	if err != nil {
		t.Fatalf("Failed to write to file: %v", err)
	}
	file.Close()

	storage := store.NewStorage()
	// replayAOF should not error - it will parse the command and dispatch it
	// DispatchCommand returns an error response string but doesn't cause replayAOF to fail
	err = replayAOF(storage, filename)
	if err != nil {
		t.Fatalf("replayAOF should not error on unknown command (it just dispatches it): %v", err)
	}

	// Verify no data was stored (since command was invalid)
	if storage.Exists("key") {
		t.Fatal("key should not exist after invalid command")
	}
}

func TestReplayAOF_PartialCommand(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.aof")

	// Create AOF file with partial/incomplete command
	file, err := os.Create(filename)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// Write a complete command first, then a partial one
	completeCmd := protocol.EncodeCommand([][]byte{[]byte("SET"), []byte("key1"), []byte("value1")})
	_, err = file.Write(completeCmd)
	if err != nil {
		t.Fatalf("Failed to write to file: %v", err)
	}

	// Write partial array command (missing second argument)
	_, err = file.WriteString("*2\r\n$3\r\nSET\r\n")
	if err != nil {
		t.Fatalf("Failed to write to file: %v", err)
	}
	file.Close()

	storage := store.NewStorage()
	err = replayAOF(storage, filename)
	// replayAOF should not error - it will stop at EOF when reading the partial command
	if err != nil {
		t.Fatalf("replayAOF should not error on partial command (stops at EOF): %v", err)
	}

	// Verify the complete command was replayed
	value, err := storage.Get("key1")
	if err != nil {
		t.Fatalf("Failed to get key1: %v", err)
	}
	if string(value) != "value1" {
		t.Fatalf("Expected 'value1', got %q", string(value))
	}

	// Verify the partial command was not executed (no key2)
	if storage.Exists("key2") {
		t.Fatal("key2 should not exist - partial command should not be executed")
	}
}
