package memory

import (
	"testing"

	"bfma-opencode-lab/internal/scenario"
)

func TestBFMACalculatesUtilityAndDiscardDecisions(t *testing.T) {
	m := NewBFMA(18, 0.20)
	events, err := m.Observe(scenario.Turn{Turn: 1, MemoryHints: []scenario.MemoryHint{
		{Content: "El sistema debe autenticar con correo institucional.", Type: "constraint", Importance: 0.95},
		{Content: "Dato muy largo poco relacionado con muchas palabras adicionales para consumir presupuesto de tokens sin aportar relevancia al prompt actual.", Type: "fact", Importance: 0.2},
	}}, AgentResponse{})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 {
		t.Fatalf("save events = %d, want 2", len(events))
	}
	ctx, err := m.BuildContext(scenario.Turn{Turn: 2, UserPrompt: "Define autenticación con correo institucional"})
	if err != nil {
		t.Fatal(err)
	}
	if len(ctx.Events) == 0 {
		t.Fatal("expected bfma decision events")
	}
	hasKeep := false
	hasScore := false
	for _, event := range ctx.Events {
		if event.Decision == "keep" {
			hasKeep = true
		}
		if event.UtilityScore > 0 {
			hasScore = true
		}
	}
	if !hasKeep {
		t.Fatal("expected at least one keep decision")
	}
	if !hasScore {
		t.Fatal("expected utility scores")
	}
}
