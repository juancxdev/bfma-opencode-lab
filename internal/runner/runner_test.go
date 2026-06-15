package runner

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"bfma-opencode-lab/internal/memory"
	"bfma-opencode-lab/internal/opencode"
)

type fakeClient struct{}

func (fakeClient) Run(ctx context.Context, req opencode.Request) (opencode.Response, error) {
	return opencode.Response{
		Text:      "respuesta fake del agente",
		Events:    []map[string]any{{"type": "fake", "agent": req.Agent}},
		RawOutput: "fake",
		Latency:   time.Millisecond,
	}, nil
}

type flakyClient struct {
	failuresRemaining int
	calls             int
}

func (c *flakyClient) Run(ctx context.Context, req opencode.Request) (opencode.Response, error) {
	c.calls++
	if c.failuresRemaining > 0 {
		c.failuresRemaining--
		return opencode.Response{}, errors.New("opencode run failed: signal: killed stderr=")
	}
	return opencode.Response{
		Text:      "respuesta recuperada del agente",
		Events:    []map[string]any{{"type": "fake", "agent": req.Agent}},
		RawOutput: "fake",
		Latency:   time.Millisecond,
	}, nil
}

type alwaysFailClient struct {
	calls int
}

func (c *alwaysFailClient) Run(ctx context.Context, req opencode.Request) (opencode.Response, error) {
	c.calls++
	return opencode.Response{}, errors.New("opencode run failed: signal: killed stderr=")
}

func TestRunnerGeneratesSeparatedLogsForGroups(t *testing.T) {
	projectDir := copyScenarioFixture(t)
	groups, err := ParseGroups("g1,g2,g3")
	if err != nil {
		t.Fatal(err)
	}
	r := New(Config{
		ProjectDir:  projectDir,
		ScenarioDir: filepath.Join(projectDir, "scenarios"),
		LogDir:      filepath.Join(projectDir, "logs"),
		DataDir:     filepath.Join(projectDir, "data"),
		ScenarioID:  "scenario_01",
		Groups:      groups,
		Reps:        1,
	}, fakeClient{})
	result, err := r.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(result.LogFiles) != 3 {
		t.Fatalf("log files = %d, want 3", len(result.LogFiles))
	}
	for _, path := range result.LogFiles {
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		line := firstLine(body)
		if !json.Valid(line) {
			t.Fatalf("first line in %s is not valid json", path)
		}
		var decoded map[string]any
		if err := json.Unmarshal(line, &decoded); err != nil {
			t.Fatal(err)
		}
		if decoded["memory_events"] == nil {
			t.Fatalf("memory_events missing in %s", path)
		}
	}
}

func TestRunnerRetriesFailedTurnAndLogsFailureThenSuccess(t *testing.T) {
	projectDir := copyScenarioFixture(t)
	client := &flakyClient{failuresRemaining: 1}
	r := New(Config{
		ProjectDir:   projectDir,
		ScenarioDir:  filepath.Join(projectDir, "scenarios"),
		LogDir:       filepath.Join(projectDir, "logs"),
		DataDir:      filepath.Join(projectDir, "data"),
		ScenarioID:   "scenario_01",
		Groups:       []memory.Group{memory.GroupG2},
		Reps:         1,
		Retries:      1,
		RetryBackoff: time.Nanosecond,
	}, client)
	result, err := r.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if client.calls < 2 {
		t.Fatalf("calls = %d, want at least 2", client.calls)
	}
	rows := readTurnLogs(t, result.LogFiles[0])
	if len(rows) != 5 {
		t.Fatalf("log rows = %d, want 5 (failed attempt + 4 successful turns)", len(rows))
	}
	if rows[0]["status"] != "failed" {
		t.Fatalf("first status = %v, want failed", rows[0]["status"])
	}
	if rows[1]["status"] != "success" {
		t.Fatalf("second status = %v, want success", rows[1]["status"])
	}
	if rows[0]["memory_after"].(map[string]any)["memory_count"].(float64) != 0 {
		t.Fatalf("failed attempt mutated memory: %#v", rows[0]["memory_after"])
	}
	if rows[1]["memory_after"].(map[string]any)["memory_count"].(float64) != 3 {
		t.Fatalf("successful retry did not update memory: %#v", rows[1]["memory_after"])
	}
}

func TestRunnerWritesFailedLogWithDiagnosticWhenRetriesExhausted(t *testing.T) {
	projectDir := copyScenarioFixture(t)
	client := &alwaysFailClient{}
	r := New(Config{
		ProjectDir:   projectDir,
		ScenarioDir:  filepath.Join(projectDir, "scenarios"),
		LogDir:       filepath.Join(projectDir, "logs"),
		DataDir:      filepath.Join(projectDir, "data"),
		ScenarioID:   "scenario_01",
		Groups:       []memory.Group{memory.GroupG3},
		Reps:         1,
		Timeout:      time.Minute,
		RetryBackoff: time.Nanosecond,
	}, client)
	result, err := r.Run(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "after 1 attempt") {
		t.Fatalf("error = %q, want exhausted attempts", err)
	}
	rows := readTurnLogs(t, result.LogFiles[0])
	if len(rows) != 1 {
		t.Fatalf("log rows = %d, want 1", len(rows))
	}
	if rows[0]["status"] != "failed" {
		t.Fatalf("status = %v, want failed", rows[0]["status"])
	}
	if got := rows[0]["error"].(string); !strings.Contains(got, "likely timeout or provider hang") {
		t.Fatalf("diagnostic = %q", got)
	}
}

func TestRunnerFromTurnReconstructsPriorMemoryWithoutLoggingPriorTurns(t *testing.T) {
	projectDir := copyScenarioFixture(t)
	r := New(Config{
		ProjectDir:  projectDir,
		ScenarioDir: filepath.Join(projectDir, "scenarios"),
		LogDir:      filepath.Join(projectDir, "logs"),
		DataDir:     filepath.Join(projectDir, "data"),
		ScenarioID:  "scenario_01",
		Groups:      []memory.Group{memory.GroupG2},
		Reps:        1,
		FromTurn:    3,
	}, fakeClient{})
	result, err := r.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	rows := readTurnLogs(t, result.LogFiles[0])
	if len(rows) != 2 {
		t.Fatalf("log rows = %d, want turns 3 and 4 only", len(rows))
	}
	if rows[0]["turn"].(float64) != 3 {
		t.Fatalf("first logged turn = %v, want 3", rows[0]["turn"])
	}
	if rows[0]["memory_before"].(map[string]any)["memory_count"].(float64) != 4 {
		t.Fatalf("warm memory count = %#v, want 4", rows[0]["memory_before"])
	}
}

func copyScenarioFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "scenarios"), 0o755); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join("..", "..", "scenarios", "scenario_01.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "scenarios", "scenario_01.json"), body, 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func readTurnLogs(t *testing.T, path string) []map[string]any {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(body)), "\n")
	rows := make([]map[string]any, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			continue
		}
		var row map[string]any
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			t.Fatal(err)
		}
		rows = append(rows, row)
	}
	return rows
}

func firstLine(body []byte) []byte {
	for i, b := range body {
		if b == '\n' {
			return body[:i]
		}
	}
	return body
}
