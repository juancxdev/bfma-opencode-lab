package memory

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"bfma-opencode-lab/internal/scenario"
)

type BFMA struct {
	records      []MemoryRecord
	lastEvents   []Event
	turnCounter  int
	tokenBudget  int
	keepMinScore float64
}

func NewBFMA(tokenBudget int, keepMinScore float64) *BFMA {
	if tokenBudget <= 0 {
		tokenBudget = 220
	}
	if keepMinScore <= 0 {
		keepMinScore = 0.28
	}
	return &BFMA{tokenBudget: tokenBudget, keepMinScore: keepMinScore}
}

func (m *BFMA) BuildContext(turn scenario.Turn) (Context, error) {
	m.turnCounter = turn.Turn
	scored := make([]scoredMemory, 0, len(m.records))
	for _, r := range m.records {
		components := m.components(r, turn)
		utility := utilityScore(components)
		scored = append(scored, scoredMemory{record: r, components: components, utility: utility})
	}
	sort.SliceStable(scored, func(i, j int) bool { return scored[i].utility > scored[j].utility })

	usedTokens := 0
	selected := []MemoryRecord{}
	events := []Event{}
	for _, item := range scored {
		cost := estimateTokens(item.record.Content)
		decision := "discard"
		reason := "below_min_score"
		if item.utility >= m.keepMinScore && usedTokens+cost <= m.tokenBudget {
			decision = "keep"
			reason = "within_budget"
			usedTokens += cost
			selected = append(selected, item.record)
		} else if item.utility >= m.keepMinScore {
			reason = "token_budget_exceeded"
		}
		events = append(events, Event{
			Type:         "bfma_decision",
			MemoryID:     item.record.ID,
			Content:      item.record.Content,
			Decision:     decision,
			UtilityScore: round3(item.utility),
			Components:   roundComponents(item.components),
			Reason:       reason,
		})
	}
	m.lastEvents = events

	lines := []string{fmt.Sprintf("Memorias seleccionadas por BFMA (presupuesto %d tokens, usados %d):", m.tokenBudget, usedTokens)}
	for _, r := range selected {
		lines = append(lines, fmt.Sprintf("- [%s] %s", r.ID, r.Content))
	}
	if len(selected) == 0 {
		lines = append(lines, "- No hay memorias seleccionadas.")
	}
	return Context{Text: strings.Join(lines, "\n"), Items: selected, Events: events, TokenUsed: usedTokens}, nil
}

func (m *BFMA) Observe(turn scenario.Turn, response AgentResponse) ([]Event, error) {
	events := []Event{}
	for _, hint := range turn.MemoryHints {
		if strings.TrimSpace(hint.Content) == "" {
			continue
		}
		if idx := m.findSimilar(hint.Content); idx >= 0 {
			m.records[idx].Frequency++
			m.records[idx].LastUsedTurn = turn.Turn
			m.records[idx].Importance = max(m.records[idx].Importance, hint.Importance)
			event := Event{Type: "update", MemoryID: m.records[idx].ID, Content: m.records[idx].Content, Reason: "similar_memory_seen"}
			events = append(events, event)
			m.lastEvents = append(m.lastEvents, event)
			continue
		}
		record := MemoryRecord{
			ID:           fmt.Sprintf("mem_%03d", len(m.records)+1),
			Content:      hint.Content,
			Type:         hint.Type,
			SourceTurn:   turn.Turn,
			CreatedAt:    time.Now().Format(time.RFC3339),
			Tags:         append([]string(nil), hint.Tags...),
			Importance:   hint.Importance,
			Frequency:    1,
			LastUsedTurn: turn.Turn,
		}
		m.records = append(m.records, record)
		event := Event{Type: "save", MemoryID: record.ID, Content: record.Content}
		events = append(events, event)
		m.lastEvents = append(m.lastEvents, event)
	}
	return events, nil
}

func (m *BFMA) Snapshot() Snapshot {
	return Snapshot{
		"memory_count":   len(m.records),
		"memories":       append([]MemoryRecord(nil), m.records...),
		"token_budget":   m.tokenBudget,
		"keep_min_score": m.keepMinScore,
	}
}

type scoredMemory struct {
	record     MemoryRecord
	components map[string]float64
	utility    float64
}

func (m *BFMA) components(record MemoryRecord, turn scenario.Turn) map[string]float64 {
	age := turn.Turn - record.SourceTurn
	if age < 0 {
		age = 0
	}
	return map[string]float64{
		"importance":   clamp01(record.Importance),
		"relevance":    overlapScore(turn.UserPrompt, record.Content),
		"recency":      clamp01(1.0 / float64(age+1)),
		"frequency":    clamp01(float64(record.Frequency) / 5.0),
		"token_cost":   clamp01(float64(estimateTokens(record.Content)) / float64(m.tokenBudget)),
		"obsolescence": obsolescenceScore(record, turn),
	}
}

func utilityScore(c map[string]float64) float64 {
	return clamp01(0.30*c["importance"] + 0.25*c["relevance"] + 0.15*c["recency"] + 0.10*c["frequency"] - 0.10*c["token_cost"] - 0.10*c["obsolescence"])
}

func obsolescenceScore(record MemoryRecord, turn scenario.Turn) float64 {
	content := strings.ToLower(record.Content + " " + turn.UserPrompt)
	if strings.Contains(content, "reemplaza") || strings.Contains(content, "ya no") || strings.Contains(content, "cambia") {
		return 0.55
	}
	return 0
}

func (m *BFMA) findSimilar(content string) int {
	for i, r := range m.records {
		if overlapScore(content, r.Content) >= 0.75 {
			return i
		}
	}
	return -1
}

func roundComponents(in map[string]float64) map[string]float64 {
	out := map[string]float64{}
	for k, v := range in {
		out[k] = round3(v)
	}
	return out
}

func round3(v float64) float64 { return float64(int(v*1000+0.5)) / 1000 }
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
