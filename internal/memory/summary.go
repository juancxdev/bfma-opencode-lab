package memory

import (
	"fmt"
	"strings"
	"time"

	"bfma-opencode-lab/internal/scenario"
)

type IncrementalSummary struct {
	currentSummary string
	versions       []string
	maxItems       int
	lastEvents     []Event
}

func NewIncrementalSummary(maxItems int) *IncrementalSummary {
	if maxItems <= 0 {
		maxItems = 8
	}
	return &IncrementalSummary{maxItems: maxItems}
}

func (m *IncrementalSummary) BuildContext(turn scenario.Turn) (Context, error) {
	text := "No existe resumen acumulado todavía."
	if strings.TrimSpace(m.currentSummary) != "" {
		text = "Resumen acumulado de la conversación:\n" + m.currentSummary
	}
	return Context{Text: text, Events: m.lastEvents, TokenUsed: estimateTokens(text)}, nil
}

func (m *IncrementalSummary) Observe(turn scenario.Turn, response AgentResponse) ([]Event, error) {
	before := m.currentSummary
	items := summaryItems(before)
	for _, hint := range turn.MemoryHints {
		items = append(items, normalizeBullet(hint.Content))
	}
	if strings.TrimSpace(response.Text) != "" {
		items = append(items, normalizeBullet("Respuesta del agente en turno "+fmt.Sprint(turn.Turn)+": "+firstSentence(response.Text)))
	}
	items = dedupeStrings(items)
	if len(items) > m.maxItems {
		items = items[len(items)-m.maxItems:]
	}
	m.currentSummary = strings.Join(items, "\n")
	m.versions = append(m.versions, m.currentSummary)
	events := []Event{{Type: "summary_update", Content: fmt.Sprintf("summary_before_len=%d summary_after_len=%d at=%s", len(before), len(m.currentSummary), time.Now().Format(time.RFC3339))}}
	m.lastEvents = events
	return events, nil
}

func (m *IncrementalSummary) Snapshot() Snapshot {
	versions := append([]string(nil), m.versions...)
	return Snapshot{
		"summary_current":  m.currentSummary,
		"summary_versions": versions,
		"version_count":    len(versions),
	}
}

func summaryItems(summary string) []string {
	lines := strings.Split(summary, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = normalizeBullet(line)
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

func normalizeBullet(s string) string {
	s = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(s), "-"))
	if s == "" {
		return ""
	}
	return "- " + s
}

func firstSentence(s string) string {
	s = strings.TrimSpace(s)
	for _, sep := range []string{".\n", ". ", "\n"} {
		if idx := strings.Index(s, sep); idx > 0 {
			return strings.TrimSpace(s[:idx+1])
		}
	}
	if len([]rune(s)) > 180 {
		return string([]rune(s)[:180]) + "..."
	}
	return s
}

func dedupeStrings(items []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(items))
	for _, item := range items {
		key := strings.ToLower(strings.TrimSpace(item))
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, item)
	}
	return out
}
