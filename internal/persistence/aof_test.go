package persistence

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestNewAOF(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.aof")

	aof := NewAOF(filename)
	if aof == nil {
		t.Fatal("NewAOF returned nil")
	}
	if aof.file == nil {
		t.Fatal("AOF file is nil")
	}

	// Cleanup
	aof.Close()
}

func TestNewAOF_CreatesFile(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.aof")

	aof := NewAOF(filename)
	defer aof.Close()

	// Check file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Fatal("AOF file was not created")
	}
}

func TestAppend(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.aof")

	aof := NewAOF(filename)
	defer aof.Close()

	// Test appending data
	data := []byte("test data")
	err := aof.Append(data)
	if err != nil {
		t.Fatalf("Append failed: %v", err)
	}

	// Verify data was written
	file, err := os.Open(filename)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	buf := make([]byte, len(data))
	n, err := file.Read(buf)
	if err != nil && err.Error() != "EOF" {
		t.Fatalf("Failed to read file: %v", err)
	}
	if n != len(data) {
		t.Fatalf("Expected to read %d bytes, got %d", len(data), n)
	}
	if string(buf) != string(data) {
		t.Fatalf("Expected %s, got %s", string(data), string(buf))
	}
}

func TestAppend_Multiple(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.aof")

	aof := NewAOF(filename)
	defer aof.Close()

	// Append multiple times
	data1 := []byte("data1")
	data2 := []byte("data2")
	data3 := []byte("data3")

	if err := aof.Append(data1); err != nil {
		t.Fatalf("Append failed: %v", err)
	}
	if err := aof.Append(data2); err != nil {
		t.Fatalf("Append failed: %v", err)
	}
	if err := aof.Append(data3); err != nil {
		t.Fatalf("Append failed: %v", err)
	}

	// Verify all data was written
	file, err := os.Open(filename)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	expected := string(data1) + string(data2) + string(data3)
	buf := make([]byte, len(expected))
	n, err := file.Read(buf)
	if err != nil && err.Error() != "EOF" {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(buf[:n]) != expected {
		t.Fatalf("Expected %s, got %s", expected, string(buf[:n]))
	}
}

func TestAppend_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.aof")

	aof := NewAOF(filename)
	defer aof.Close()

	// Append empty data
	err := aof.Append([]byte{})
	if err != nil {
		t.Fatalf("Append failed with empty data: %v", err)
	}
}

func TestAppend_LargeData(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.aof")

	aof := NewAOF(filename)
	defer aof.Close()

	// Append large data
	largeData := make([]byte, 1024*1024) // 1MB
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	err := aof.Append(largeData)
	if err != nil {
		t.Fatalf("Append failed with large data: %v", err)
	}

	// Verify size
	info, err := os.Stat(filename)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}
	if info.Size() != int64(len(largeData)) {
		t.Fatalf("Expected file size %d, got %d", len(largeData), info.Size())
	}
}

func TestClose(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.aof")

	aof := NewAOF(filename)

	// Write some data
	data := []byte("test data")
	if err := aof.Append(data); err != nil {
		t.Fatalf("Append failed: %v", err)
	}

	// Close should sync and close
	err := aof.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Verify file is closed (write should fail)
	err = aof.Append([]byte("more data"))
	if err == nil {
		t.Fatal("Append should fail after Close")
	}
}

func TestClose_MultipleCalls(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.aof")

	aof := NewAOF(filename)

	// First close should succeed
	err := aof.Close()
	if err != nil {
		t.Fatalf("First Close failed: %v", err)
	}

	// Subsequent closes should not panic (though behavior may vary)
	// This tests that Close is idempotent-safe
	err = aof.Close()
	// We don't check error here as behavior may vary by OS
}

func TestConcurrentAppend(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.aof")

	aof := NewAOF(filename)
	defer aof.Close()

	// Concurrent appends
	var wg sync.WaitGroup
	numGoroutines := 10
	iterations := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				data := []byte{byte(id), byte(j)}
				if err := aof.Append(data); err != nil {
					t.Errorf("Append failed in goroutine %d: %v", id, err)
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify file has expected size (each append is 2 bytes)
	expectedSize := int64(numGoroutines * iterations * 2)
	info, err := os.Stat(filename)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}
	if info.Size() != expectedSize {
		t.Fatalf("Expected file size %d, got %d", expectedSize, info.Size())
	}
}

func TestAppend_ThreadSafety(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.aof")

	aof := NewAOF(filename)
	defer aof.Close()

	// Test that mutex prevents race conditions
	var wg sync.WaitGroup
	iterations := 1000

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				data := []byte("test")
				if err := aof.Append(data); err != nil {
					t.Errorf("Append failed: %v", err)
				}
			}
		}()
	}

	wg.Wait()

	// Verify all data was written correctly
	info, err := os.Stat(filename)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}
	expectedSize := int64(10 * iterations * len("test"))
	if info.Size() != expectedSize {
		t.Fatalf("Expected file size %d, got %d", expectedSize, info.Size())
	}
}

