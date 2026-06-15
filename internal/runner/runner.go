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
	ProjectDir   string
	ScenarioDir  string
	LogDir       string
	DataDir      string
	ScenarioID   string
	Groups       []memory.Group
	Reps         int
	Model        string
	DryRun       bool
	Timeout      time.Duration
	Retries      int
	RetryBackoff time.Duration
	FromTurn     int
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
	if cfg.Retries < 0 {
		cfg.Retries = 0
	}
	if cfg.RetryBackoff < 0 {
		cfg.RetryBackoff = 0
	}
	if cfg.RetryBackoff == 0 {
		cfg.RetryBackoff = 2 * time.Second
	}
	if cfg.FromTurn < 0 {
		cfg.FromTurn = 0
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
			if logPath != "" {
				result.LogFiles = append(result.LogFiles, logPath)
			}
			if err != nil {
				return result, err
			}
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

	if err := r.warmMemoryUntilTurn(sc, manager); err != nil {
		return logPath, err
	}

	for _, turn := range sc.Turns {
		if r.cfg.FromTurn > 0 && turn.Turn < r.cfg.FromTurn {
			continue
		}
		memoryBefore := manager.Snapshot()
		memCtx, err := manager.BuildContext(turn)
		if err != nil {
			return logPath, fmt.Errorf("build memory context group=%s turn=%d: %w", group, turn.Turn, err)
		}
		prompt := BuildPrompt(string(group), memCtx.Text, turn.UserPrompt)

		if err := r.runTurnWithRetries(ctx, writer, runID, sc, group, rep, agent, turn, manager, memoryBefore, memCtx, prompt); err != nil {
			return logPath, err
		}
	}
	return logPath, nil
}

func (r *Runner) warmMemoryUntilTurn(sc scenario.Scenario, manager memory.Manager) error {
	if r.cfg.FromTurn <= 0 {
		return nil
	}
	found := false
	for _, turn := range sc.Turns {
		if turn.Turn == r.cfg.FromTurn {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("from-turn %d not found in scenario %s", r.cfg.FromTurn, sc.ID)
	}
	for _, turn := range sc.Turns {
		if turn.Turn >= r.cfg.FromTurn {
			break
		}
		if _, err := manager.BuildContext(turn); err != nil {
			return fmt.Errorf("warm memory build context turn=%d: %w", turn.Turn, err)
		}
		if _, err := manager.Observe(turn, memory.AgentResponse{}); err != nil {
			return fmt.Errorf("warm memory observe turn=%d: %w", turn.Turn, err)
		}
	}
	return nil
}

func (r *Runner) runTurnWithRetries(
	ctx context.Context,
	writer *logging.JSONL,
	runID string,
	sc scenario.Scenario,
	group memory.Group,
	rep int,
	agent string,
	turn scenario.Turn,
	manager memory.Manager,
	memoryBefore memory.Snapshot,
	memCtx memory.Context,
	prompt string,
) error {
	maxAttempts := r.cfg.Retries + 1
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
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
			lastErr = err
			entry := r.turnLog(runID, sc, group, rep, turn, agent, memoryBefore, memoryBefore, memCtx, nil, opencode.Response{}, attempt, maxAttempts)
			entry.Status = "failed"
			entry.Error = diagnoseRunError(err, r.cfg.Timeout, agent, turn.Turn)
			if err := writer.Write(entry); err != nil {
				return err
			}
			if attempt < maxAttempts {
				if err := sleepWithContext(ctx, r.cfg.RetryBackoff); err != nil {
					return err
				}
				continue
			}
			break
		}

		postEvents, err := manager.Observe(turn, memory.AgentResponse{Text: resp.Text})
		if err != nil {
			return fmt.Errorf("observe memory group=%s turn=%d: %w", group, turn.Turn, err)
		}
		memoryAfter := manager.Snapshot()
		entry := r.turnLog(runID, sc, group, rep, turn, agent, memoryBefore, memoryAfter, memCtx, postEvents, resp, attempt, maxAttempts)
		entry.Status = "success"
		if err := writer.Write(entry); err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("run opencode group=%s turn=%d after %d attempt(s): %w", group, turn.Turn, maxAttempts, lastErr)
}

func (r *Runner) turnLog(
	runID string,
	sc scenario.Scenario,
	group memory.Group,
	rep int,
	turn scenario.Turn,
	agent string,
	memoryBefore memory.Snapshot,
	memoryAfter memory.Snapshot,
	memCtx memory.Context,
	postEvents []memory.Event,
	resp opencode.Response,
	attempt int,
	maxAttempts int,
) instrumentation.TurnLog {
	combinedEvents := append([]memory.Event{}, memCtx.Events...)
	combinedEvents = append(combinedEvents, postEvents...)
	entry := instrumentation.TurnLog{
		RunID:                 runID,
		Group:                 string(group),
		ScenarioID:            sc.ID,
		Rep:                   rep,
		Turn:                  turn.Turn,
		Attempt:               attempt,
		MaxAttempts:           maxAttempts,
		Agent:                 agent,
		MemoryContextInjected: memCtx.Text,
		UserPrompt:            turn.UserPrompt,
		AssistantResponse:     resp.Text,
		TimeoutMS:             r.cfg.Timeout.Milliseconds(),
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
	return entry
}

func diagnoseRunError(err error, timeout time.Duration, agent string, turn int) string {
	msg := err.Error()
	if strings.Contains(msg, "context deadline exceeded") || strings.Contains(msg, "signal: killed") || strings.Contains(msg, "stderr empty") {
		return fmt.Sprintf("opencode run failed; likely timeout or provider hang; timeout=%s; agent=%s; turn=%d; cause=%s", timeout, agent, turn, msg)
	}
	return fmt.Sprintf("opencode run failed; timeout=%s; agent=%s; turn=%d; cause=%s", timeout, agent, turn, msg)
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
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
