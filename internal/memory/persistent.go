package memory

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"bfma-opencode-lab/internal/scenario"
)

type Persistent struct {
	records    []MemoryRecord
	storePath  string
	lastEvents []Event
	maxContext int
}

func NewPersistent(storePath string, maxContext int) *Persistent {
	if maxContext <= 0 {
		maxContext = 8
	}
	return &Persistent{storePath: storePath, maxContext: maxContext}
}

func (m *Persistent) BuildContext(turn scenario.Turn) (Context, error) {
	retrieved := topByRelevance(m.records, turn.UserPrompt, m.maxContext)
	lines := []string{"Memorias persistentes recuperadas:"}
	for _, r := range retrieved {
		lines = append(lines, fmt.Sprintf("- [%s] %s", r.ID, r.Content))
	}
	if len(retrieved) == 0 {
		lines = append(lines, "- No hay memorias relevantes recuperadas.")
	}
	text := strings.Join(lines, "\n")
	events := make([]Event, 0, len(retrieved))
	for _, r := range retrieved {
		events = append(events, Event{Type: "retrieve", MemoryID: r.ID, Content: r.Content, Reason: "text_overlap"})
	}
	m.lastEvents = events
	return Context{Text: text, Items: retrieved, Events: events, TokenUsed: estimateTokens(text)}, nil
}

func (m *Persistent) Observe(turn scenario.Turn, response AgentResponse) ([]Event, error) {
	events := []Event{}
	for _, hint := range turn.MemoryHints {
		if strings.TrimSpace(hint.Content) == "" {
			continue
		}
		record := MemoryRecord{
			ID:         fmt.Sprintf("mem_%03d", len(m.records)+1),
			Content:    hint.Content,
			Type:       hint.Type,
			SourceTurn: turn.Turn,
			CreatedAt:  time.Now().Format(time.RFC3339),
			Tags:       append([]string(nil), hint.Tags...),
			Importance: hint.Importance,
			Frequency:  1,
		}
		m.records = append(m.records, record)
		event := Event{Type: "save", MemoryID: record.ID, Content: record.Content}
		events = append(events, event)
		m.lastEvents = append(m.lastEvents, event)
		if err := m.appendRecord(record); err != nil {
			return events, err
		}
	}
	return events, nil
}

func (m *Persistent) Snapshot() Snapshot {
	return Snapshot{
		"memory_count": len(m.records),
		"memories":     append([]MemoryRecord(nil), m.records...),
	}
}

func (m *Persistent) appendRecord(record MemoryRecord) error {
	if m.storePath == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(m.storePath), 0o755); err != nil {
		return fmt.Errorf("create memory store dir: %w", err)
	}
	file, err := os.OpenFile(m.storePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open memory store: %w", err)
	}
	defer file.Close()
	line, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("marshal memory record: %w", err)
	}
	if _, err := file.Write(append(line, '\n')); err != nil {
		return fmt.Errorf("write memory record: %w", err)
	}
	return nil
}
