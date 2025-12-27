package store

import (
	"errors"
	"math"
	"sync"
	"time"
)

var ErrWrongType = errors.New("Wrong type")

type ValueType int

const (
	StringType ValueType = iota
	ListType
)

type Value struct {
	Kind      ValueType
	Str       []byte
	List      [][]byte
	ExpiresAt time.Time
}

type Storage struct {
	mu   sync.Mutex
	data map[string]Value
}

func NewStorage() *Storage {
	return &Storage{
		data: make(map[string]Value),
	}
}

func (s *Storage) Set(key string, value []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	valueCopy := make([]byte, len(value))
	copy(valueCopy, value)
	s.data[key] = Value{
		Kind:      StringType,
		Str:       valueCopy,
		ExpiresAt: time.Time{},
	}
}

func (s *Storage) Get(key string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	value, ok := s.getIfNotExpired(key)
	if !ok {
		return nil, nil
	}
	if value.Kind != StringType {
		return nil, ErrWrongType
	}
	valueCopy := make([]byte, len(value.Str))
	copy(valueCopy, value.Str)
	return valueCopy, nil
}

func (s *Storage) Del(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.getIfNotExpired(key)
	if !ok {
		return false
	}
	delete(s.data, key)
	return true
}

func (s *Storage) Exists(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.getIfNotExpired(key)
	return ok
}

func (s *Storage) SetEx(key string, seconds int64, value []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	valueCopy := make([]byte, len(value))
	if seconds <= 0 {
		delete(s.data, key)
		return
	}
	expiresAt := time.Now().Add(time.Duration(seconds) * time.Second)
	copy(valueCopy, value)
	s.data[key] = Value{
		Kind:      StringType,
		Str:       valueCopy,
		ExpiresAt: expiresAt,
	}
}

func (s *Storage) LPush(key string, value []byte) (int, error) {
	// This is not efficient
	s.mu.Lock()
	defer s.mu.Unlock()
	storageValue, ok := s.getIfNotExpired(key)
	if ok && storageValue.Kind != ListType {
		return 0, ErrWrongType
	}
	valueCopy := make([]byte, len(value))
	copy(valueCopy, value)
	if !ok {
		storageValue = Value{
			Kind:      ListType,
			List:      [][]byte{valueCopy},
			ExpiresAt: time.Time{},
		}
	} else {
		newList := make([][]byte, 0, len(storageValue.List)+1)
		newList = append(newList, valueCopy)
		newList = append(newList, storageValue.List...)
		storageValue.List = newList
	}
	s.data[key] = storageValue
	return len(storageValue.List), nil
}

func (s *Storage) RPush(key string, value []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	storageValue, ok := s.getIfNotExpired(key)
	if ok && storageValue.Kind != ListType {
		return 0, ErrWrongType
	}
	valueCopy := make([]byte, len(value))
	copy(valueCopy, value)
	if !ok {
		storageValue = Value{
			Kind:      ListType,
			List:      [][]byte{valueCopy},
			ExpiresAt: time.Time{},
		}
	} else {
		storageValue.List = append(storageValue.List, valueCopy)
	}
	s.data[key] = storageValue
	return len(storageValue.List), nil
}

func (s *Storage) LRange(key string, start, end int) ([][]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	value, ok := s.getIfNotExpired(key)

	if !ok {
		return nil, nil
	} else if value.Kind != ListType {
		return nil, ErrWrongType
	}

	if start < 0 {
		start = len(value.List) + start
		start = max(start, 0)
	}
	if end < 0 {
		end = len(value.List) + end
	}
	if end >= len(value.List) {
		end = len(value.List) - 1
	}
	if start > end || start >= len(value.List) || end < 0 {
		return [][]byte{}, nil
	}

	slice := value.List[start : end+1]
	sliceCopy := make([][]byte, len(slice))
	for i, v := range slice {
		sliceCopy[i] = make([]byte, len(v))
		copy(sliceCopy[i], v)
	}
	return sliceCopy, nil
}

func (s *Storage) Expire(key string, seconds int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	value, ok := s.getIfNotExpired(key)
	if !ok {
		return false
	}
	if seconds <= 0 {
		delete(s.data, key)
		return false
	}
	expiresAt := time.Now().Add(time.Duration(seconds) * time.Second)
	value.ExpiresAt = expiresAt
	s.data[key] = value
	return true
}

func (s *Storage) TTL(key string) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	value, ok := s.getIfNotExpired(key)
	if !ok {
		return -2
	}
	if value.ExpiresAt.IsZero() {
		return -1
	}
	// Edge case: if key expired after getIfNotExpired, we need to delete it
	remaining := time.Until(value.ExpiresAt)
	if remaining <= 0 {
		delete(s.data, key)
		return -2
	}
	return int64(math.Ceil(remaining.Seconds()))
}

func (value Value) IsExpired() bool {
	if value.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().After(value.ExpiresAt)
}

func (s *Storage) getIfNotExpired(key string) (Value, bool) {
	// Method should be called with the lock held
	v, ok := s.data[key]
	if !ok {
		return Value{}, false
	}
	if v.IsExpired() {
		delete(s.data, key)
		return Value{}, false
	}
	return v, true
}
