package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type JSONL struct {
	mu   sync.Mutex
	path string
	file *os.File
}

func NewJSONL(path string) (*JSONL, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create log dir: %w", err)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}
	return &JSONL{path: path, file: file}, nil
}

func (l *JSONL) Write(v any) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	line, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal jsonl: %w", err)
	}
	if _, err := l.file.Write(append(line, '\n')); err != nil {
		return fmt.Errorf("write jsonl: %w", err)
	}
	return nil
}

func (l *JSONL) Close() error {
	if l.file == nil {
		return nil
	}
	return l.file.Close()
}

func (l *JSONL) Path() string { return l.path }
