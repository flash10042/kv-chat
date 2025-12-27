package persistence

import (
	"log"
	"os"
	"sync"
)

type AOF struct {
	file *os.File
	mu   sync.Mutex
}

func NewAOF(filename string) *AOF {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open AOF file: %v", err)
	}
	return &AOF{
		file: file,
	}
}

func (a *AOF) Append(command []byte) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	_, err := a.file.Write(command)
	return err
}

func (a *AOF) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if err := a.file.Sync(); err != nil {
		return err
	}
	return a.file.Close()
}
