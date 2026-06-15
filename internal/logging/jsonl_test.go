package logging

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestJSONLWritesValidJSONLine(t *testing.T) {
	path := filepath.Join(t.TempDir(), "run.jsonl")
	logger, err := NewJSONL(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := logger.Write(map[string]any{"group": "g1", "turn": 1}); err != nil {
		t.Fatal(err)
	}
	if err := logger.Close(); err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		t.Fatal("expected one jsonl line")
	}
	var decoded map[string]any
	if err := json.Unmarshal(scanner.Bytes(), &decoded); err != nil {
		t.Fatalf("invalid json line: %v", err)
	}
}
