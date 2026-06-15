package opencode

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type Request struct {
	ProjectDir string
	Agent      string
	Model      string
	Prompt     string
	DryRun     bool
}

type Response struct {
	Text      string           `json:"text"`
	Events    []map[string]any `json:"events"`
	RawOutput string           `json:"raw_output"`
	Latency   time.Duration    `json:"latency"`
}

type Client interface {
	Run(ctx context.Context, req Request) (Response, error)
}

type CLIClient struct {
	Binary string
}

func NewCLIClient(binary string) *CLIClient {
	if binary == "" {
		binary = "opencode"
	}
	return &CLIClient{Binary: binary}
}

func (c *CLIClient) Run(ctx context.Context, req Request) (Response, error) {
	if req.DryRun {
		return dryRun(req), nil
	}
	if strings.TrimSpace(req.Agent) == "" {
		return Response{}, errors.New("opencode agent is required")
	}
	args := []string{"run", "--format", "json", "--pure", "--agent", req.Agent, "--dir", req.ProjectDir}
	if strings.TrimSpace(req.Model) != "" {
		args = append(args, "--model", req.Model)
	}
	args = append(args, req.Prompt)

	cmd := exec.CommandContext(ctx, c.Binary, args...)
	cmd.Dir = req.ProjectDir
	cmd.Env = append(cmd.Environ(), "OPENCODE_DISABLE_AUTOCOMPACT=true")

	started := time.Now()
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		stderrText := strings.TrimSpace(stderr.String())
		if ctxErr := ctx.Err(); ctxErr != nil {
			return Response{}, fmt.Errorf("opencode run failed: %w; context=%v; likely timeout or provider hang; agent=%s stderr=%s", err, ctxErr, req.Agent, stderrText)
		}
		if stderrText == "" {
			return Response{}, fmt.Errorf("opencode run failed: %w; stderr empty; process may have been killed by OS or provider runtime; agent=%s", err, req.Agent)
		}
		return Response{}, fmt.Errorf("opencode run failed: %w stderr=%s", err, stderrText)
	}

	events, text := ParseJSONEvents(stdout.String())
	if strings.TrimSpace(text) == "" {
		text = strings.TrimSpace(stdout.String())
	}
	return Response{Text: text, Events: events, RawOutput: stdout.String(), Latency: time.Since(started)}, nil
}

func dryRun(req Request) Response {
	text := "Respuesta simulada para " + req.Agent + ": se atiende el turno usando exclusivamente el MEMORY_CONTEXT entregado."
	events := []map[string]any{{"type": "dry_run", "agent": req.Agent, "text": text}}
	return Response{Text: text, Events: events, RawOutput: text, Latency: time.Millisecond}
}

func ParseJSONEvents(raw string) ([]map[string]any, string) {
	events := []map[string]any{}
	texts := []string{}
	scanner := bufio.NewScanner(strings.NewReader(raw))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var event map[string]any
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}
		events = append(events, event)
		texts = append(texts, extractText(event)...)
	}
	return events, strings.TrimSpace(strings.Join(dedupeText(texts), "\n"))
}

func extractText(v any) []string {
	out := []string{}
	switch typed := v.(type) {
	case map[string]any:
		for key, val := range typed {
			lk := strings.ToLower(key)
			if lk == "text" || lk == "content" || lk == "message" || lk == "output" || lk == "response" {
				if s, ok := val.(string); ok && looksLikeAssistantText(s) {
					out = append(out, s)
				}
			}
			out = append(out, extractText(val)...)
		}
	case []any:
		for _, item := range typed {
			out = append(out, extractText(item)...)
		}
	}
	return out
}

func looksLikeAssistantText(s string) bool {
	s = strings.TrimSpace(s)
	if len(s) < 12 {
		return false
	}
	if strings.HasPrefix(s, "{") || strings.HasPrefix(s, "[") {
		return false
	}
	return true
}

func dedupeText(items []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" || seen[item] {
			continue
		}
		seen[item] = true
		out = append(out, item)
	}
	return out
}
