package instrumentation

import (
	"bfma-opencode-lab/internal/memory"
	"bfma-opencode-lab/internal/scenario"
)

type TurnLog struct {
	RunID                 string                   `json:"run_id"`
	Group                 string                   `json:"group"`
	ScenarioID            string                   `json:"scenario_id"`
	Rep                   int                      `json:"rep"`
	Turn                  int                      `json:"turn"`
	Agent                 string                   `json:"agent"`
	MemoryContextInjected string                   `json:"memory_context_injected"`
	UserPrompt            string                   `json:"user_prompt"`
	AssistantResponse     string                   `json:"assistant_response"`
	LatencyMS             int64                    `json:"latency_ms"`
	OpenCodeEvents        []map[string]any         `json:"opencode_events"`
	MemoryBefore          memory.Snapshot          `json:"memory_before"`
	MemoryAfter           memory.Snapshot          `json:"memory_after"`
	MemoryEvents          []memory.Event           `json:"memory_events"`
	MemoryPreEvents       []memory.Event           `json:"memory_pre_events,omitempty"`
	MemoryPostEvents      []memory.Event           `json:"memory_post_events,omitempty"`
	FinalQuestions        []scenario.FinalQuestion `json:"final_questions,omitempty"`
}
