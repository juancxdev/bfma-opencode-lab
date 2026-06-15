package memory

import (
	"testing"

	"bfma-opencode-lab/internal/scenario"
)

func TestIncrementalSummaryKeepsVersions(t *testing.T) {
	m := NewIncrementalSummary(4)
	turn := scenario.Turn{Turn: 1, MemoryHints: []scenario.MemoryHint{{Content: "La latencia máxima es 2 segundos.", Type: "constraint", Importance: 0.9}}}
	events, err := m.Observe(turn, AgentResponse{Text: "Se propone una arquitectura modular."})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) == 0 {
		t.Fatal("expected summary update event")
	}
	snap := m.Snapshot()
	if got := snap["version_count"]; got != 1 {
		t.Fatalf("version_count = %v, want 1", got)
	}
	if snap["summary_current"] == "" {
		t.Fatal("summary_current is empty")
	}
}
