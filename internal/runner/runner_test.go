package runner

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

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

func firstLine(body []byte) []byte {
	for i, b := range body {
		if b == '\n' {
			return body[:i]
		}
	}
	return body
}
