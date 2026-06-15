package memory

import (
	"os"
	"path/filepath"
	"testing"

	"bfma-opencode-lab/internal/scenario"
)

func TestPersistentSavesAndRetrievesWithoutDeleting(t *testing.T) {
	store := filepath.Join(t.TempDir(), "memories.jsonl")
	m := NewPersistent(store, 5)
	turn := scenario.Turn{Turn: 1, UserPrompt: "autenticación correo", MemoryHints: []scenario.MemoryHint{{Content: "El sistema autentica con correo institucional.", Type: "constraint", Importance: 0.9}}}
	events, err := m.Observe(turn, AgentResponse{})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Type != "save" {
		t.Fatalf("expected one save event, got %#v", events)
	}
	ctx, err := m.BuildContext(scenario.Turn{Turn: 2, UserPrompt: "¿Cómo autentica el sistema?"})
	if err != nil {
		t.Fatal(err)
	}
	if len(ctx.Items) == 0 {
		t.Fatal("expected retrieved memory")
	}
	if _, err := os.Stat(store); err != nil {
		t.Fatalf("expected isolated memory store: %v", err)
	}
	if got := m.Snapshot()["memory_count"]; got != 1 {
		t.Fatalf("memory_count = %v, want 1", got)
	}
}
