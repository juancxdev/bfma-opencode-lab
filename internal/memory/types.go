package memory

import "bfma-opencode-lab/internal/scenario"

type Group string

const (
	GroupG1 Group = "g1"
	GroupG2 Group = "g2"
	GroupG3 Group = "g3"
)

type Context struct {
	Text      string         `json:"text"`
	Items     []MemoryRecord `json:"items,omitempty"`
	Events    []Event        `json:"events,omitempty"`
	TokenUsed int            `json:"token_used,omitempty"`
}

type Snapshot map[string]any

type AgentResponse struct {
	Text string `json:"text"`
}

type Manager interface {
	BuildContext(turn scenario.Turn) (Context, error)
	Observe(turn scenario.Turn, response AgentResponse) ([]Event, error)
	Snapshot() Snapshot
}

type MemoryRecord struct {
	ID           string   `json:"id"`
	Content      string   `json:"content"`
	Type         string   `json:"type"`
	SourceTurn   int      `json:"source_turn"`
	CreatedAt    string   `json:"created_at"`
	Tags         []string `json:"tags,omitempty"`
	Importance   float64  `json:"importance,omitempty"`
	Frequency    int      `json:"frequency,omitempty"`
	LastUsedTurn int      `json:"last_used_turn,omitempty"`
}

type Event struct {
	MemoryID     string             `json:"memory_id,omitempty"`
	Type         string             `json:"type"`
	Content      string             `json:"content,omitempty"`
	Decision     string             `json:"decision,omitempty"`
	UtilityScore float64            `json:"utility_score,omitempty"`
	Components   map[string]float64 `json:"components,omitempty"`
	Reason       string             `json:"reason,omitempty"`
}
