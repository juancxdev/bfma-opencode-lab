package report

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"bfma-opencode-lab/internal/instrumentation"
	"bfma-opencode-lab/internal/memory"
	"bfma-opencode-lab/internal/scenario"
)

func TestLoadRunParsesJSONLAndAggregatesBFMADecisions(t *testing.T) {
	runDir := writeRunFixture(t)

	data, err := LoadRun(runDir)
	if err != nil {
		t.Fatal(err)
	}

	if data.RunID != "run_test" {
		t.Fatalf("run id = %q", data.RunID)
	}
	if data.TotalTurns != 2 {
		t.Fatalf("total turns = %d, want 2", data.TotalTurns)
	}
	if data.SuccessTurns != 2 || data.FailedTurns != 0 {
		t.Fatalf("success/fail = %d/%d", data.SuccessTurns, data.FailedTurns)
	}
	if data.TotalKeep != 1 || data.TotalDiscard != 2 {
		t.Fatalf("keep/discard = %d/%d, want 1/2", data.TotalKeep, data.TotalDiscard)
	}
	if data.ReasonCounts["below_min_score"] != 1 || data.ReasonCounts["token_budget_exceeded"] != 1 {
		t.Fatalf("reason counts = %#v", data.ReasonCounts)
	}
	if data.FinalMemoryCount != 3 {
		t.Fatalf("final memory count = %d, want 3", data.FinalMemoryCount)
	}
	if len(data.Groups) != 1 || data.Groups[0] != "g3" {
		t.Fatalf("groups = %#v", data.Groups)
	}
	if !data.HasBFMA {
		t.Fatal("expected HasBFMA")
	}
	if len(data.FinalQuestions) != 1 {
		t.Fatalf("final questions = %d, want 1", len(data.FinalQuestions))
	}
}

func TestLoadRunSupportsFailedAttemptsAndRetries(t *testing.T) {
	runDir := t.TempDir()
	logDir := filepath.Join(runDir, "g3")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeJSONL(t, filepath.Join(logDir, "scenario_rep_01.jsonl"), []instrumentation.TurnLog{
		{
			RunID:       "run_failed",
			Status:      "failed",
			Group:       "g3",
			ScenarioID:  "scenario",
			Rep:         1,
			Turn:        1,
			Attempt:     1,
			MaxAttempts: 2,
			Error:       "opencode run failed; likely timeout or provider hang",
			MemoryAfter: memory.Snapshot{
				"memory_count": 0,
			},
		},
		{
			RunID:       "run_failed",
			Status:      "success",
			Group:       "g3",
			ScenarioID:  "scenario",
			Rep:         1,
			Turn:        1,
			Attempt:     2,
			MaxAttempts: 2,
			LatencyMS:   1000,
			MemoryAfter: memory.Snapshot{
				"memory_count": 1,
			},
		},
	})

	data, err := LoadRun(runDir)
	if err != nil {
		t.Fatal(err)
	}
	if data.FailedTurns != 1 || data.SuccessTurns != 1 {
		t.Fatalf("success/fail = %d/%d", data.SuccessTurns, data.FailedTurns)
	}
	if !data.HasFailures {
		t.Fatal("expected failures")
	}
	if data.RetryCount != 1 {
		t.Fatalf("retry count = %d, want 1", data.RetryCount)
	}
}

func TestLoadRunBuildsGroupComparisonForThreeGroups(t *testing.T) {
	runDir := writeComparisonRunFixture(t)

	data, err := LoadRun(runDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(data.GroupSummaries) != 3 {
		t.Fatalf("group summaries = %d, want 3", len(data.GroupSummaries))
	}
	byGroup := map[string]GroupSummary{}
	for _, summary := range data.GroupSummaries {
		byGroup[summary.Group] = summary
	}
	if byGroup["g1"].FinalSummaryLines != 2 {
		t.Fatalf("g1 summary lines = %d, want 2", byGroup["g1"].FinalSummaryLines)
	}
	if byGroup["g2"].FinalMemoryCount != 69 {
		t.Fatalf("g2 memory count = %d, want 69", byGroup["g2"].FinalMemoryCount)
	}
	if byGroup["g3"].TotalKeep != 2 || byGroup["g3"].TotalDiscard != 1 {
		t.Fatalf("g3 keep/discard = %d/%d", byGroup["g3"].TotalKeep, byGroup["g3"].TotalDiscard)
	}
	if byGroup["g3"].FinalContextItems != 2 {
		t.Fatalf("g3 final context items = %d, want 2", byGroup["g3"].FinalContextItems)
	}
	if byGroup["g2"].CoveragePercent != 100 {
		t.Fatalf("g2 coverage = %d, want 100", byGroup["g2"].CoveragePercent)
	}
	if byGroup["g1"].CoveragePercent >= byGroup["g2"].CoveragePercent {
		t.Fatalf("expected g1 coverage below g2: g1=%d g2=%d", byGroup["g1"].CoveragePercent, byGroup["g2"].CoveragePercent)
	}
	if len(data.FinalByGroup) != 3 {
		t.Fatalf("final by group = %d, want 3", len(data.FinalByGroup))
	}
	if len(data.GroupLatency) != 3 || len(data.GroupAnswerSize) != 3 || len(data.GroupCoverage) != 3 {
		t.Fatalf("comparison chart points missing")
	}
}

func TestExpectedAnswerDetectedNormalizesAccentsAndUsesTokens(t *testing.T) {
	detected, matched, total := expectedAnswerDetected(
		"El estudiante usa correo institucional y la matrícula valida prerrequisitos antes del registro.",
		"Correo institucional y validación de prerrequisitos antes de matrícula.",
	)
	if !detected {
		t.Fatalf("expected detection, matched=%d total=%d", matched, total)
	}
	detected, _, _ = expectedAnswerDetected("Respuesta sin relación", "Correo institucional")
	if detected {
		t.Fatal("unexpected detection")
	}
}

func TestRenderProducesSelfContainedHTML(t *testing.T) {
	data := Data{
		RunID:        "run_html",
		ScenarioID:   "scenario_html",
		Groups:       []string{"g3"},
		TotalTurns:   1,
		GeneratedAt:  "2026-06-15T00:00:00Z",
		ReasonCounts: map[string]int{},
		Turns: []TurnSummary{{
			Group:  "g3",
			Turn:   1,
			Status: "success",
		}},
	}
	var b strings.Builder
	if err := Render(&b, data); err != nil {
		t.Fatal(err)
	}
	html := b.String()
	for _, want := range []string{
		"Reporte experimental BFMA",
		"Comparación experimental por grupo",
		"Lectura metodológica",
		"Comparación de respuestas finales",
		"REPORT_PAYLOAD",
		"drawBar",
		"run_html",
		"notion-page",
		"notion-callout",
		"notion-toc",
		"details",
		"Modo captura",
		"canvas.width = Math.floor(cssWidth * ratio)",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html missing %q", want)
		}
	}
	if strings.Contains(html, "https://") || strings.Contains(html, "http://") {
		t.Fatal("report should not depend on remote assets")
	}
}

func writeComparisonRunFixture(t *testing.T) string {
	t.Helper()
	runDir := t.TempDir()
	questions := []scenario.FinalQuestion{
		{ID: "q1", Question: "¿Autenticación?", ExpectedAnswer: "Correo institucional."},
		{ID: "q2", Question: "¿Validación?", ExpectedAnswer: "Prerrequisitos antes de matrícula."},
	}
	for _, group := range []string{"g1", "g2", "g3"} {
		logDir := filepath.Join(runDir, group)
		if err := os.MkdirAll(logDir, 0o755); err != nil {
			t.Fatal(err)
		}
		entry := instrumentation.TurnLog{
			RunID:                 "run_compare",
			Status:                "success",
			Group:                 group,
			ScenarioID:            "scenario_compare",
			Rep:                   1,
			Turn:                  18,
			Attempt:               1,
			MaxAttempts:           1,
			LatencyMS:             1000,
			MemoryContextInjected: "- [mem_001] Correo institucional.\n- [mem_002] Prerrequisitos antes de matrícula.",
			AssistantResponse:     "Arquitectura con correo institucional y prerrequisitos antes de matrícula.",
			FinalQuestions:        questions,
		}
		switch group {
		case "g1":
			entry.AssistantResponse = "Arquitectura académica general sin mencionar autenticación."
			entry.MemoryAfter = memory.Snapshot{
				"summary_current": "- regla uno\n- regla dos",
				"version_count":   18,
			}
		case "g2":
			entry.MemoryAfter = memory.Snapshot{"memory_count": 69}
		case "g3":
			entry.MemoryAfter = memory.Snapshot{"memory_count": 69, "token_budget": 220}
			entry.MemoryContextInjected = "Memorias seleccionadas por BFMA (presupuesto 220 tokens, usados 80):\n- [mem_001] Correo institucional.\n- [mem_002] Prerrequisitos antes de matrícula."
			entry.MemoryPreEvents = []memory.Event{
				{Type: "bfma_decision", Decision: "keep"},
				{Type: "bfma_decision", Decision: "keep"},
				{Type: "bfma_decision", Decision: "discard", Reason: "below_min_score"},
			}
		}
		writeJSONL(t, filepath.Join(logDir, "scenario_rep_01.jsonl"), []instrumentation.TurnLog{entry})
	}
	return runDir
}

func writeRunFixture(t *testing.T) string {
	t.Helper()
	runDir := t.TempDir()
	logDir := filepath.Join(runDir, "g3")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeJSONL(t, filepath.Join(logDir, "scenario_rep_01.jsonl"), []instrumentation.TurnLog{
		{
			RunID:                 "run_test",
			Status:                "success",
			Group:                 "g3",
			ScenarioID:            "scenario_test",
			Rep:                   1,
			Turn:                  1,
			Attempt:               1,
			MaxAttempts:           1,
			LatencyMS:             2000,
			MemoryContextInjected: "Memorias seleccionadas por BFMA (presupuesto 220 tokens, usados 100):",
			MemoryPreEvents: []memory.Event{
				{Type: "bfma_decision", Decision: "keep", Reason: "within_budget"},
				{Type: "bfma_decision", Decision: "discard", Reason: "below_min_score"},
			},
			MemoryAfter: memory.Snapshot{
				"memory_count": 2,
				"token_budget": 220,
			},
		},
		{
			RunID:                 "run_test",
			Status:                "success",
			Group:                 "g3",
			ScenarioID:            "scenario_test",
			Rep:                   1,
			Turn:                  2,
			Attempt:               1,
			MaxAttempts:           1,
			LatencyMS:             4000,
			MemoryContextInjected: "Memorias seleccionadas por BFMA (presupuesto 220 tokens, usados 218):",
			AssistantResponse:     "respuesta final",
			MemoryPreEvents: []memory.Event{
				{Type: "bfma_decision", Decision: "discard", Reason: "token_budget_exceeded"},
			},
			MemoryAfter: memory.Snapshot{
				"memory_count": 3,
				"token_budget": 220,
			},
			FinalQuestions: []scenario.FinalQuestion{
				{ID: "q1", Question: "Pregunta", ExpectedAnswer: "Respuesta"},
			},
		},
	})
	return runDir
}

func writeJSONL(t *testing.T, path string, rows []instrumentation.TurnLog) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	for i, row := range rows {
		if i == 1 {
			if _, err := file.WriteString("\n"); err != nil {
				t.Fatal(err)
			}
		}
		body, err := json.Marshal(row)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := file.Write(append(body, '\n')); err != nil {
			t.Fatal(err)
		}
	}
}
