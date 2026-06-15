package runner

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"bfma-opencode-lab/internal/instrumentation"
	"bfma-opencode-lab/internal/logging"
	"bfma-opencode-lab/internal/memory"
	"bfma-opencode-lab/internal/opencode"
	"bfma-opencode-lab/internal/scenario"
)

type Config struct {
	ProjectDir  string
	ScenarioDir string
	LogDir      string
	DataDir     string
	ScenarioID  string
	Groups      []memory.Group
	Reps        int
	Model       string
	DryRun      bool
	Timeout     time.Duration
}

type Result struct {
	RunID    string
	LogFiles []string
}

type Runner struct {
	cfg    Config
	client opencode.Client
}

func New(cfg Config, client opencode.Client) *Runner {
	if cfg.Reps <= 0 {
		cfg.Reps = 1
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 5 * time.Minute
	}
	if cfg.LogDir == "" {
		cfg.LogDir = filepath.Join(cfg.ProjectDir, "logs")
	}
	if cfg.DataDir == "" {
		cfg.DataDir = filepath.Join(cfg.ProjectDir, "data")
	}
	return &Runner{cfg: cfg, client: client}
}

func (r *Runner) Run(ctx context.Context) (Result, error) {
	sc, err := scenario.Load(r.cfg.ScenarioDir, r.cfg.ScenarioID)
	if err != nil {
		return Result{}, err
	}
	runID := "run_" + time.Now().Format("20060102_150405")
	result := Result{RunID: runID}
	for _, group := range r.cfg.Groups {
		for rep := 1; rep <= r.cfg.Reps; rep++ {
			logPath, err := r.runGroupRep(ctx, runID, sc, group, rep)
			if err != nil {
				return result, err
			}
			result.LogFiles = append(result.LogFiles, logPath)
		}
	}
	return result, nil
}

func (r *Runner) runGroupRep(ctx context.Context, runID string, sc scenario.Scenario, group memory.Group, rep int) (string, error) {
	manager, agent, err := r.newManager(runID, group)
	if err != nil {
		return "", err
	}
	logPath := filepath.Join(r.cfg.LogDir, runID, string(group), fmt.Sprintf("%s_rep_%02d.jsonl", sc.ID, rep))
	writer, err := logging.NewJSONL(logPath)
	if err != nil {
		return "", err
	}
	defer writer.Close()

	for _, turn := range sc.Turns {
		memoryBefore := manager.Snapshot()
		memCtx, err := manager.BuildContext(turn)
		if err != nil {
			return logPath, fmt.Errorf("build memory context group=%s turn=%d: %w", group, turn.Turn, err)
		}
		prompt := BuildPrompt(string(group), memCtx.Text, turn.UserPrompt)

		turnCtx, cancel := context.WithTimeout(ctx, r.cfg.Timeout)
		resp, err := r.client.Run(turnCtx, opencode.Request{
			ProjectDir: r.cfg.ProjectDir,
			Agent:      agent,
			Model:      r.cfg.Model,
			Prompt:     prompt,
			DryRun:     r.cfg.DryRun,
		})
		cancel()
		if err != nil {
			return logPath, fmt.Errorf("run opencode group=%s turn=%d: %w", group, turn.Turn, err)
		}

		postEvents, err := manager.Observe(turn, memory.AgentResponse{Text: resp.Text})
		if err != nil {
			return logPath, fmt.Errorf("observe memory group=%s turn=%d: %w", group, turn.Turn, err)
		}
		memoryAfter := manager.Snapshot()
		combinedEvents := append([]memory.Event{}, memCtx.Events...)
		combinedEvents = append(combinedEvents, postEvents...)
		entry := instrumentation.TurnLog{
			RunID:                 runID,
			Group:                 string(group),
			ScenarioID:            sc.ID,
			Rep:                   rep,
			Turn:                  turn.Turn,
			Agent:                 agent,
			MemoryContextInjected: memCtx.Text,
			UserPrompt:            turn.UserPrompt,
			AssistantResponse:     resp.Text,
			LatencyMS:             resp.Latency.Milliseconds(),
			OpenCodeEvents:        resp.Events,
			MemoryBefore:          memoryBefore,
			MemoryAfter:           memoryAfter,
			MemoryEvents:          combinedEvents,
			MemoryPreEvents:       memCtx.Events,
			MemoryPostEvents:      postEvents,
		}
		if turn.Turn == sc.Turns[len(sc.Turns)-1].Turn {
			entry.FinalQuestions = sc.FinalQuestions
		}
		if err := writer.Write(entry); err != nil {
			return logPath, err
		}
	}
	return logPath, nil
}

func (r *Runner) newManager(runID string, group memory.Group) (memory.Manager, string, error) {
	switch group {
	case memory.GroupG1:
		return memory.NewIncrementalSummary(8), "g1-summary-agent", nil
	case memory.GroupG2:
		store := filepath.Join(r.cfg.DataDir, runID, string(group), "memories.jsonl")
		return memory.NewPersistent(store, 8), "g2-persistent-agent", nil
	case memory.GroupG3:
		return memory.NewBFMA(220, 0.28), "g3-bfma-agent", nil
	default:
		return nil, "", fmt.Errorf("unsupported group %q", group)
	}
}

func BuildPrompt(group string, memoryContext string, userPrompt string) string {
	var b strings.Builder
	b.WriteString("EXPERIMENT_GROUP: ")
	b.WriteString(group)
	b.WriteString("\n\nMEMORY_CONTEXT:\n")
	b.WriteString(memoryContext)
	b.WriteString("\n\nCURRENT_USER_TURN:\n")
	b.WriteString(userPrompt)
	b.WriteString("\n\nINSTRUCTIONS:\nResponde únicamente con la información disponible en MEMORY_CONTEXT y CURRENT_USER_TURN. No incluyas métricas, JSON experimental ni trazas internas.\n")
	return b.String()
}

func ParseGroups(raw string) ([]memory.Group, error) {
	parts := strings.Split(raw, ",")
	out := make([]memory.Group, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(strings.ToLower(part))
		if part == "" {
			continue
		}
		group := memory.Group(part)
		switch group {
		case memory.GroupG1, memory.GroupG2, memory.GroupG3:
			out = append(out, group)
		default:
			return nil, fmt.Errorf("unknown group %q", part)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("at least one group is required")
	}
	return out, nil
}
