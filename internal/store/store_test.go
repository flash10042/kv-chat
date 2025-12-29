package store

import (
	"reflect"
	"testing"
	"time"
)

// Test NewStorage
func TestNewStorage(t *testing.T) {
	storage := NewStorage()
	if storage == nil {
		t.Fatal("NewStorage returned nil")
	}
	if storage.data == nil {
		t.Fatal("Storage data map is nil")
	}
}

// Test Set and Get - Positive cases
func TestSetGet(t *testing.T) {
	storage := NewStorage()

	// Test basic set and get
	storage.Set("key", []byte("value"))
	value, err := storage.Get("key")
	if err != nil {
		t.Fatalf("Failed to get value: %v", err)
	}
	if string(value) != "value" {
		t.Fatalf("Expected value to be 'value', got %s", string(value))
	}

	// Test overwriting
	storage.Set("key", []byte("newvalue"))
	value, err = storage.Get("key")
	if err != nil {
		t.Fatalf("Failed to get value: %v", err)
	}
	if string(value) != "newvalue" {
		t.Fatalf("Expected value to be 'newvalue', got %s", string(value))
	}

	// Test empty value
	storage.Set("empty", []byte(""))
	value, err = storage.Get("empty")
	if err != nil {
		t.Fatalf("Failed to get empty value: %v", err)
	}
	if len(value) != 0 {
		t.Fatalf("Expected empty value, got %s", string(value))
	}
}

// Test Get - Negative cases
func TestGetNonExistent(t *testing.T) {
	storage := NewStorage()

	value, err := storage.Get("nonexistent")
	if err != nil {
		t.Fatalf("Get should not return error for non-existent key, got: %v", err)
	}
	if value != nil {
		t.Fatalf("Expected nil for non-existent key, got %v", value)
	}
}

func TestGetWrongType(t *testing.T) {
	storage := NewStorage()

	// Create a list first
	_, err := storage.LPush("key", []byte("value"))
	if err != nil {
		t.Fatalf("Failed to LPush: %v", err)
	}

	// Try to get it as string
	_, err = storage.Get("key")
	if err != ErrWrongType {
		t.Fatalf("Expected ErrWrongType, got %v", err)
	}
}

// Test Del - Positive cases
func TestDel(t *testing.T) {
	storage := NewStorage()

	// Delete existing key
	storage.Set("key", []byte("value"))
	deleted := storage.Del("key")
	if !deleted {
		t.Fatal("Del should return true for existing key")
	}

	// Verify it's deleted
	value, _ := storage.Get("key")
	if value != nil {
		t.Fatal("Key should be deleted")
	}
}

// Test Del - Negative cases
func TestDelNonExistent(t *testing.T) {
	storage := NewStorage()

	deleted := storage.Del("nonexistent")
	if deleted {
		t.Fatal("Del should return false for non-existent key")
	}
}

// Test Exists - Positive cases
func TestExists(t *testing.T) {
	storage := NewStorage()

	storage.Set("key", []byte("value"))
	if !storage.Exists("key") {
		t.Fatal("Exists should return true for existing key")
	}
}

// Test Exists - Negative cases
func TestExistsNonExistent(t *testing.T) {
	storage := NewStorage()

	if storage.Exists("nonexistent") {
		t.Fatal("Exists should return false for non-existent key")
	}
}

// Test LPush - Positive cases
func TestLPush(t *testing.T) {
	storage := NewStorage()

	// Push to new list
	length, err := storage.LPush("list", []byte("first"))
	if err != nil {
		t.Fatalf("LPush failed: %v", err)
	}
	if length != 1 {
		t.Fatalf("Expected length 1, got %d", length)
	}

	// Push more elements
	length, err = storage.LPush("list", []byte("second"))
	if err != nil {
		t.Fatalf("LPush failed: %v", err)
	}
	if length != 2 {
		t.Fatalf("Expected length 2, got %d", length)
	}

	// Verify order (LPush adds to front)
	values, err := storage.LRange("list", 0, -1)
	if err != nil {
		t.Fatalf("LRange failed: %v", err)
	}
	if len(values) != 2 {
		t.Fatalf("Expected 2 values, got %d", len(values))
	}
	if string(values[0]) != "second" {
		t.Fatalf("Expected first element to be 'second', got %s", string(values[0]))
	}
	if string(values[1]) != "first" {
		t.Fatalf("Expected second element to be 'first', got %s", string(values[1]))
	}
}

// Test LPush - Negative cases
func TestLPushWrongType(t *testing.T) {
	storage := NewStorage()

	// Create a string value
	storage.Set("key", []byte("value"))

	// Try to LPush to string
	_, err := storage.LPush("key", []byte("item"))
	if err != ErrWrongType {
		t.Fatalf("Expected ErrWrongType, got %v", err)
	}
}

// Test RPush - Positive cases
func TestRPush(t *testing.T) {
	storage := NewStorage()

	// Push to new list
	length, err := storage.RPush("list", []byte("first"))
	if err != nil {
		t.Fatalf("RPush failed: %v", err)
	}
	if length != 1 {
		t.Fatalf("Expected length 1, got %d", length)
	}

	// Push more elements
	length, err = storage.RPush("list", []byte("second"))
	if err != nil {
		t.Fatalf("RPush failed: %v", err)
	}
	if length != 2 {
		t.Fatalf("Expected length 2, got %d", length)
	}

	// Verify order (RPush adds to end)
	values, err := storage.LRange("list", 0, -1)
	if err != nil {
		t.Fatalf("LRange failed: %v", err)
	}
	if len(values) != 2 {
		t.Fatalf("Expected 2 values, got %d", len(values))
	}
	if string(values[0]) != "first" {
		t.Fatalf("Expected first element to be 'first', got %s", string(values[0]))
	}
	if string(values[1]) != "second" {
		t.Fatalf("Expected second element to be 'second', got %s", string(values[1]))
	}
}

// Test RPush - Negative cases
func TestRPushWrongType(t *testing.T) {
	storage := NewStorage()

	// Create a string value
	storage.Set("key", []byte("value"))

	// Try to RPush to string
	_, err := storage.RPush("key", []byte("item"))
	if err != ErrWrongType {
		t.Fatalf("Expected ErrWrongType, got %v", err)
	}
}

// Test LRange - Positive cases
func TestLRange(t *testing.T) {
	storage := NewStorage()

	// Setup list
	storage.RPush("list", []byte("a"))
	storage.RPush("list", []byte("b"))
	storage.RPush("list", []byte("c"))
	storage.RPush("list", []byte("d"))

	// Test full range
	values, err := storage.LRange("list", 0, -1)
	if err != nil {
		t.Fatalf("LRange failed: %v", err)
	}
	if len(values) != 4 {
		t.Fatalf("Expected 4 values, got %d", len(values))
	}

	// Test partial range
	values, err = storage.LRange("list", 1, 2)
	if err != nil {
		t.Fatalf("LRange failed: %v", err)
	}
	if len(values) != 2 {
		t.Fatalf("Expected 2 values, got %d", len(values))
	}
	if string(values[0]) != "b" || string(values[1]) != "c" {
		t.Fatalf("Expected ['b', 'c'], got [%s, %s]", string(values[0]), string(values[1]))
	}

	// Test negative indices
	values, err = storage.LRange("list", -2, -1)
	if err != nil {
		t.Fatalf("LRange failed: %v", err)
	}
	if len(values) != 2 {
		t.Fatalf("Expected 2 values, got %d", len(values))
	}
	if string(values[0]) != "c" || string(values[1]) != "d" {
		t.Fatalf("Expected ['c', 'd'], got [%s, %s]", string(values[0]), string(values[1]))
	}

	// Test single element
	values, err = storage.LRange("list", 0, 0)
	if err != nil {
		t.Fatalf("LRange failed: %v", err)
	}
	if len(values) != 1 {
		t.Fatalf("Expected 1 value, got %d", len(values))
	}
}

// Test LRange - Negative cases
func TestLRangeNonExistent(t *testing.T) {
	storage := NewStorage()

	values, err := storage.LRange("nonexistent", 0, -1)
	if err != nil {
		t.Fatalf("LRange should not return error for non-existent key, got: %v", err)
	}
	if values != nil {
		t.Fatalf("Expected nil for non-existent key, got %v", values)
	}
}

func TestLRangeWrongType(t *testing.T) {
	storage := NewStorage()

	storage.Set("key", []byte("value"))
	_, err := storage.LRange("key", 0, -1)
	if err != ErrWrongType {
		t.Fatalf("Expected ErrWrongType, got %v", err)
	}
}

func TestLRangeInvalidRange(t *testing.T) {
	storage := NewStorage()

	storage.RPush("list", []byte("a"))

	// Test start > end
	values, err := storage.LRange("list", 2, 1)
	if err != nil {
		t.Fatalf("LRange should not return error for invalid range, got: %v", err)
	}
	if len(values) != 0 {
		t.Fatalf("Expected empty slice for invalid range, got %v", values)
	}

	// Test start >= length
	values, err = storage.LRange("list", 10, 20)
	if err != nil {
		t.Fatalf("LRange should not return error, got: %v", err)
	}
	if len(values) != 0 {
		t.Fatalf("Expected empty slice, got %v", values)
	}
}

// Test Expire - Positive cases
func TestExpire(t *testing.T) {
	storage := NewStorage()

	storage.Set("key", []byte("value"))
	ok := storage.Expire("key", 10)
	if !ok {
		t.Fatal("Expire should return true for existing key")
	}

	// Key should still exist
	if !storage.Exists("key") {
		t.Fatal("Key should still exist after Expire")
	}
}

// Test Expire - Negative cases
func TestExpireNonExistent(t *testing.T) {
	storage := NewStorage()

	ok := storage.Expire("nonexistent", 10)
	if ok {
		t.Fatal("Expire should return false for non-existent key")
	}
}

func TestExpireZeroOrNegative(t *testing.T) {
	storage := NewStorage()

	storage.Set("key", []byte("value"))
	ok := storage.Expire("key", 0)
	if !ok {
		t.Fatal("Expire should return true for zero seconds")
	}

	// Key should be deleted
	if storage.Exists("key") {
		t.Fatal("Key should be deleted when expire is 0")
	}

	// Test negative
	storage.Set("key2", []byte("value"))
	ok = storage.Expire("key2", -1)
	if !ok {
		t.Fatal("Expire should return true for negative seconds")
	}
	if storage.Exists("key2") {
		t.Fatal("Key should be deleted when expire is negative")
	}
}

// Test TTL - Positive cases
func TestTTL(t *testing.T) {
	storage := NewStorage()

	storage.Set("key", []byte("value"))
	ttl := storage.TTL("key")
	if ttl != -1 {
		t.Fatalf("Expected TTL -1 (no expiration), got %d", ttl)
	}

	storage.Expire("key", 10)
	ttl = storage.TTL("key")
	if ttl < 1 || ttl > 10 {
		t.Fatalf("Expected TTL between 1 and 10, got %d", ttl)
	}
}

// Test TTL - Negative cases
func TestTTLNonExistent(t *testing.T) {
	storage := NewStorage()

	ttl := storage.TTL("nonexistent")
	if ttl != -2 {
		t.Fatalf("Expected TTL -2 (key doesn't exist), got %d", ttl)
	}
}

// Test SetEx - Positive cases
func TestSetEx(t *testing.T) {
	storage := NewStorage()

	storage.SetEx("key", 10, []byte("value"))
	value, err := storage.Get("key")
	if err != nil {
		t.Fatalf("Failed to get value: %v", err)
	}
	if string(value) != "value" {
		t.Fatalf("Expected 'value', got %s", string(value))
	}

	ttl := storage.TTL("key")
	if ttl < 1 || ttl > 10 {
		t.Fatalf("Expected TTL between 1 and 10, got %d", ttl)
	}
}

// Test SetEx - Negative cases
func TestSetExZeroOrNegative(t *testing.T) {
	storage := NewStorage()

	storage.SetEx("key", 0, []byte("value"))
	if storage.Exists("key") {
		t.Fatal("Key should not exist when SetEx with 0 seconds")
	}

	storage.SetEx("key2", -1, []byte("value"))
	if storage.Exists("key2") {
		t.Fatal("Key should not exist when SetEx with negative seconds")
	}
}

// Test expiration behavior
func TestExpiration(t *testing.T) {
	storage := NewStorage()

	// Set with very short expiration
	storage.SetEx("key", 1, []byte("value"))
	if !storage.Exists("key") {
		t.Fatal("Key should exist immediately after SetEx")
	}

	// Wait for expiration
	time.Sleep(1100 * time.Millisecond)

	// Key should be expired
	if storage.Exists("key") {
		t.Fatal("Key should be expired")
	}

	value, _ := storage.Get("key")
	if value != nil {
		t.Fatal("Get should return nil for expired key")
	}
}

// Test concurrent access
func TestConcurrentAccess(t *testing.T) {
	storage := NewStorage()
	done := make(chan bool)

	// Concurrent writes
	go func() {
		for i := 0; i < 100; i++ {
			storage.Set("key1", []byte("value1"))
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			storage.Set("key2", []byte("value2"))
		}
		done <- true
	}()

	// Concurrent reads
	go func() {
		for i := 0; i < 100; i++ {
			storage.Get("key1")
		}
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}

	// Verify final state
	value1, _ := storage.Get("key1")
	value2, _ := storage.Get("key2")
	if string(value1) != "value1" || string(value2) != "value2" {
		t.Fatal("Concurrent access corrupted data")
	}
}

// Test value copying (ensure mutations don't affect stored values)
func TestValueCopying(t *testing.T) {
	storage := NewStorage()

	original := []byte("original")
	storage.Set("key", original)

	// Modify original
	original[0] = 'X'

	// Stored value should not be affected
	value, _ := storage.Get("key")
	if string(value) != "original" {
		t.Fatalf("Stored value was modified, expected 'original', got %s", string(value))
	}
}

// Test list value copying
func TestListValueCopying(t *testing.T) {
	storage := NewStorage()

	storage.RPush("list", []byte("a"))
	storage.RPush("list", []byte("b"))

	values, _ := storage.LRange("list", 0, -1)
	if len(values) != 2 {
		t.Fatalf("Expected 2 values, got %d", len(values))
	}

	// Modify returned values
	values[0][0] = 'X'

	// Original should not be affected
	values2, _ := storage.LRange("list", 0, -1)
	if string(values2[0]) != "a" {
		t.Fatalf("Stored value was modified, expected 'a', got %s", string(values2[0]))
	}
}

// Test IsExpired
func TestIsExpired(t *testing.T) {
	value := Value{
		Kind:      StringType,
		Str:       []byte("test"),
		ExpiresAt: time.Now().Add(-1 * time.Second), // Expired
	}
	if !value.IsExpired() {
		t.Fatal("Value should be expired")
	}

	value2 := Value{
		Kind:      StringType,
		Str:       []byte("test"),
		ExpiresAt: time.Now().Add(1 * time.Hour), // Not expired
	}
	if value2.IsExpired() {
		t.Fatal("Value should not be expired")
	}

	value3 := Value{
		Kind:      StringType,
		Str:       []byte("test"),
		ExpiresAt: time.Time{}, // No expiration
	}
	if value3.IsExpired() {
		t.Fatal("Value with zero expiration time should not be expired")
	}
}

// Test type constants
func TestValueTypes(t *testing.T) {
	if StringType != 0 {
		t.Fatalf("Expected StringType to be 0, got %d", StringType)
	}
	if ListType != 1 {
		t.Fatalf("Expected ListType to be 1, got %d", ListType)
	}
}

// Test Value struct
func TestValueStruct(t *testing.T) {
	value := Value{
		Kind:      StringType,
		Str:       []byte("test"),
		ExpiresAt: time.Now(),
	}

	if value.Kind != StringType {
		t.Fatal("Kind should be StringType")
	}
	if !reflect.DeepEqual(value.Str, []byte("test")) {
		t.Fatal("Str should match")
	}
}
