package memory

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"bfma-opencode-lab/internal/scenario"
)

type BFMA struct {
	records     []MemoryRecord
	lastEvents  []Event
	turnCounter int
	config      BFMAConfig
}

func NewBFMA(tokenBudget int, keepMinScore float64) *BFMA {
	return NewBFMAWithConfig(NewBFMAConfig(tokenBudget, keepMinScore))
}

func NewBFMAWithConfig(config BFMAConfig) *BFMA {
	return &BFMA{config: config.Normalize()}
}

func (m *BFMA) BuildContext(turn scenario.Turn) (Context, error) {
	m.turnCounter = turn.Turn
	scored := make([]scoredMemory, 0, len(m.records))
	for _, r := range m.records {
		components, obsolescence := m.components(r, turn)
		breakdown := CalculateScore(components, m.config.Weights)
		scored = append(scored, scoredMemory{record: r, breakdown: breakdown, utility: breakdown.BFMAUtility, obsolescence: obsolescence})
	}
	sort.SliceStable(scored, func(i, j int) bool { return scored[i].utility > scored[j].utility })

	usedTokens := 0
	selected := []MemoryRecord{}
	events := []Event{}
	for _, item := range scored {
		cost := estimateTokens(item.record.Content)
		decision := DecideForget(item.utility, m.config.KeepMinScore, usedTokens, cost, m.config.TokenBudget, item.obsolescence)
		if decision.Decision == "keep" {
			usedTokens += cost
			selected = append(selected, item.record)
		}
		events = append(events, Event{
			Type:                "bfma_decision",
			MemoryID:            item.record.ID,
			Content:             item.record.Content,
			Decision:            decision.Decision,
			UtilityScore:        item.breakdown.BFMAUtility,
			AntecedentScore:     item.breakdown.AntecedentScore,
			BFMAUtility:         item.breakdown.BFMAUtility,
			FormulaVersion:      item.breakdown.FormulaVersion,
			Components:          componentsMap(item.breakdown.Components),
			Weights:             weightsMap(item.breakdown.Weights),
			Reason:              decision.Reason,
			ObsolescenceReason:  item.obsolescence.Reason,
			BudgetUsed:          usedTokens,
			BudgetLimit:         m.config.TokenBudget,
			TokenCost:           cost,
			FrequencyBonus:      item.breakdown.FrequencyBonus,
			TokenCostPenalty:    item.breakdown.TokenCostPenalty,
			ObsolescencePenalty: item.breakdown.ObsolescencePenalty,
		})
	}
	m.lastEvents = events

	lines := []string{fmt.Sprintf("Memorias seleccionadas por BFMA (presupuesto %d tokens, usados %d):", m.config.TokenBudget, usedTokens)}
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
		"memory_count":    len(m.records),
		"memories":        append([]MemoryRecord(nil), m.records...),
		"token_budget":    m.config.TokenBudget,
		"keep_min_score":  m.config.KeepMinScore,
		"formula_version": BFMAFormulaVersion,
		"weights":         weightsMap(m.config.Weights),
	}
}

type scoredMemory struct {
	record       MemoryRecord
	breakdown    ScoreBreakdown
	utility      float64
	obsolescence ObsolescenceAssessment
}

func (m *BFMA) components(record MemoryRecord, turn scenario.Turn) (ScoreComponents, ObsolescenceAssessment) {
	age := turn.Turn - record.SourceTurn
	if age < 0 {
		age = 0
	}
	obsolescence := AssessObsolescence(record, m.records)
	return ScoreComponents{
		Importance:   clamp01(record.Importance),
		Relevance:    overlapScore(turn.UserPrompt, record.Content),
		Recency:      clamp01(1.0 / float64(age+1)),
		Frequency:    clamp01(float64(record.Frequency) / 5.0),
		TokenCost:    clamp01(float64(estimateTokens(record.Content)) / float64(m.config.TokenBudget)),
		Obsolescence: obsolescence.Score,
	}, obsolescence
}

func (m *BFMA) findSimilar(content string) int {
	for i, r := range m.records {
		if overlapScore(content, r.Content) >= 0.75 {
			return i
		}
	}
	return -1
}

func componentsMap(c ScoreComponents) map[string]float64 {
	return map[string]float64{
		"importance":   c.Importance,
		"relevance":    c.Relevance,
		"recency":      c.Recency,
		"frequency":    c.Frequency,
		"token_cost":   c.TokenCost,
		"obsolescence": c.Obsolescence,
	}
}

func weightsMap(w ScoreWeights) map[string]float64 {
	return map[string]float64{
		"importance":   w.Importance,
		"relevance":    w.Relevance,
		"recency":      w.Recency,
		"frequency":    w.Frequency,
		"token_cost":   w.TokenCost,
		"obsolescence": w.Obsolescence,
	}
}

func round3(v float64) float64 { return float64(int(v*1000+0.5)) / 1000 }
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
