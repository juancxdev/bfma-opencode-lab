package scenario

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Scenario struct {
	ID             string          `json:"scenario_id"`
	Title          string          `json:"title"`
	Description    string          `json:"description"`
	Turns          []Turn          `json:"turns"`
	FinalQuestions []FinalQuestion `json:"final_questions"`
	GroundTruth    GroundTruth     `json:"ground_truth"`
}

type Turn struct {
	Turn        int          `json:"turn"`
	UserPrompt  string       `json:"user_prompt"`
	MemoryHints []MemoryHint `json:"memory_hints"`
}

type MemoryHint struct {
	Content    string   `json:"content"`
	Type       string   `json:"type"`
	Importance float64  `json:"importance"`
	Tags       []string `json:"tags"`
}

type FinalQuestion struct {
	ID             string `json:"id"`
	Question       string `json:"question"`
	ExpectedAnswer string `json:"expected_answer"`
}

type GroundTruth struct {
	CriticalFacts          []string `json:"critical_facts"`
	ExpectedConstraints    []string `json:"expected_constraints"`
	ExpectedContradictions []string `json:"expected_contradictions"`
}

func Load(dir, id string) (Scenario, error) {
	path := filepath.Join(dir, id+".json")
	body, err := os.ReadFile(path)
	if err != nil {
		return Scenario{}, fmt.Errorf("read scenario %q: %w", path, err)
	}
	var out Scenario
	if err := json.Unmarshal(body, &out); err != nil {
		return Scenario{}, fmt.Errorf("decode scenario %q: %w", path, err)
	}
	if out.ID == "" {
		return Scenario{}, fmt.Errorf("scenario %q missing scenario_id", path)
	}
	if len(out.Turns) == 0 {
		return Scenario{}, fmt.Errorf("scenario %q has no turns", path)
	}
	return out, nil
}
