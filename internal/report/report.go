package report

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"bfma-opencode-lab/internal/instrumentation"
	"bfma-opencode-lab/internal/memory"
	"bfma-opencode-lab/internal/scenario"
)

type Data struct {
	RunID              string           `json:"run_id"`
	ScenarioID         string           `json:"scenario_id"`
	Groups             []string         `json:"groups"`
	GroupSummaries     []GroupSummary   `json:"group_summaries"`
	FinalByGroup       []GroupFinal     `json:"final_by_group"`
	TotalTurns         int              `json:"total_turns"`
	SuccessTurns       int              `json:"success_turns"`
	FailedTurns        int              `json:"failed_turns"`
	RetryCount         int              `json:"retry_count"`
	AvgLatencyMS       int64            `json:"avg_latency_ms"`
	TotalKeep          int              `json:"total_keep"`
	TotalDiscard       int              `json:"total_discard"`
	FinalMemoryCount   int              `json:"final_memory_count"`
	GeneratedAt        string           `json:"generated_at"`
	ReasonCounts       map[string]int   `json:"reason_counts"`
	Turns              []TurnSummary    `json:"turns"`
	FinalContext       string           `json:"final_context"`
	FinalAnswer        string           `json:"final_answer"`
	FinalQuestions     []Question       `json:"final_questions"`
	Findings           []string         `json:"findings"`
	SourceFiles        []string         `json:"source_files"`
	HasFailures        bool             `json:"has_failures"`
	HasBFMA            bool             `json:"has_bfma"`
	LatencyByTurn      []ChartPoint     `json:"latency_by_turn"`
	MemoryByTurn       []ChartPoint     `json:"memory_by_turn"`
	KeepDiscardByTurn  []DecisionPoint  `json:"keep_discard_by_turn"`
	BudgetByTurn       []BudgetPoint    `json:"budget_by_turn"`
	ReasonDistribution []Distribution   `json:"reason_distribution"`
	GroupLatency       []ChartPoint     `json:"group_latency"`
	GroupAnswerSize    []ChartPoint     `json:"group_answer_size"`
	GroupCoverage      []ChartPoint     `json:"group_coverage"`
	GroupContextSize   []ChartPoint     `json:"group_context_size"`
	MemoryPressure     []MemoryPressure `json:"memory_pressure"`
}

type TurnSummary struct {
	Group       string `json:"group"`
	Turn        int    `json:"turn"`
	Status      string `json:"status"`
	Attempt     int    `json:"attempt"`
	MaxAttempts int    `json:"max_attempts"`
	LatencyMS   int64  `json:"latency_ms"`
	MemoryCount int    `json:"memory_count"`
	Keep        int    `json:"keep"`
	Discard     int    `json:"discard"`
	TokenBudget int    `json:"token_budget"`
	TokenUsed   int    `json:"token_used"`
	PromptShort string `json:"prompt_short"`
	AnswerShort string `json:"answer_short"`
	Error       string `json:"error,omitempty"`
	FinalTurn   bool   `json:"final_turn"`
}

type GroupSummary struct {
	Group                 string         `json:"group"`
	Label                 string         `json:"label"`
	Strategy              string         `json:"strategy"`
	Turns                 int            `json:"turns"`
	SuccessTurns          int            `json:"success_turns"`
	FailedTurns           int            `json:"failed_turns"`
	AvgLatencyMS          int64          `json:"avg_latency_ms"`
	FinalMemoryCount      int            `json:"final_memory_count"`
	FinalSummaryLines     int            `json:"final_summary_lines"`
	FinalContextItems     int            `json:"final_context_items"`
	FinalAnswerChars      int            `json:"final_answer_chars"`
	FinalAnswerWords      int            `json:"final_answer_words"`
	TotalKeep             int            `json:"total_keep"`
	TotalDiscard          int            `json:"total_discard"`
	FinalKeep             int            `json:"final_keep"`
	FinalDiscard          int            `json:"final_discard"`
	TokenBudget           int            `json:"token_budget"`
	TokenUsed             int            `json:"token_used"`
	CoverageDetected      int            `json:"coverage_detected"`
	CoverageTotal         int            `json:"coverage_total"`
	CoveragePercent       int            `json:"coverage_percent"`
	Coverage              []CoverageItem `json:"coverage"`
	Omissions             []CoverageItem `json:"omissions"`
	FinalPromptShort      string         `json:"final_prompt_short"`
	FinalAnswerShort      string         `json:"final_answer_short"`
	MethodologicalReading string         `json:"methodological_reading"`
}

type GroupFinal struct {
	Group             string         `json:"group"`
	Label             string         `json:"label"`
	FinalContext      string         `json:"final_context"`
	FinalAnswer       string         `json:"final_answer"`
	Coverage          []CoverageItem `json:"coverage"`
	Omissions         []CoverageItem `json:"omissions"`
	FinalContextItems int            `json:"final_context_items"`
	FinalAnswerChars  int            `json:"final_answer_chars"`
}

type CoverageItem struct {
	ID             string `json:"id"`
	Question       string `json:"question"`
	ExpectedAnswer string `json:"expected_answer"`
	Detected       bool   `json:"detected"`
	MatchedTokens  int    `json:"matched_tokens"`
	TotalTokens    int    `json:"total_tokens"`
}

type Question struct {
	ID             string `json:"id"`
	Question       string `json:"question"`
	ExpectedAnswer string `json:"expected_answer"`
}

type ChartPoint struct {
	Label string `json:"label"`
	Value int64  `json:"value"`
	Group string `json:"group"`
}

type DecisionPoint struct {
	Label   string `json:"label"`
	Keep    int    `json:"keep"`
	Discard int    `json:"discard"`
	Group   string `json:"group"`
}

type BudgetPoint struct {
	Label  string `json:"label"`
	Used   int    `json:"used"`
	Budget int    `json:"budget"`
	Group  string `json:"group"`
}

type Distribution struct {
	Label string `json:"label"`
	Value int    `json:"value"`
}

type MemoryPressure struct {
	Group    string `json:"group"`
	Label    string `json:"label"`
	Stored   int    `json:"stored"`
	Selected int    `json:"selected"`
}

type Options struct {
	RunDir string
	Out    string
}

func Generate(opts Options) error {
	data, err := LoadRun(opts.RunDir)
	if err != nil {
		return err
	}
	if strings.TrimSpace(opts.Out) == "" {
		return fmt.Errorf("output path is required")
	}
	if err := os.MkdirAll(filepath.Dir(opts.Out), 0o755); err != nil {
		return fmt.Errorf("create report dir: %w", err)
	}
	file, err := os.Create(opts.Out)
	if err != nil {
		return fmt.Errorf("create report: %w", err)
	}
	defer file.Close()
	if err := Render(file, data); err != nil {
		return err
	}
	return nil
}

func LoadRun(runDir string) (Data, error) {
	if strings.TrimSpace(runDir) == "" {
		return Data{}, fmt.Errorf("run dir is required")
	}
	entries, files, err := readRunLogs(runDir)
	if err != nil {
		return Data{}, err
	}
	if len(entries) == 0 {
		return Data{}, fmt.Errorf("no JSONL turn logs found under %s", runDir)
	}
	return aggregate(entries, files), nil
}

func Render(w io.Writer, data Data) error {
	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal report data: %w", err)
	}
	payload := base64.StdEncoding.EncodeToString(body)
	tpl, err := template.New("report").Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("parse report template: %w", err)
	}
	return tpl.Execute(w, struct {
		Title   string
		Payload string
	}{
		Title:   "Reporte BFMA - " + data.RunID,
		Payload: payload,
	})
}

func readRunLogs(runDir string) ([]instrumentation.TurnLog, []string, error) {
	info, err := os.Stat(runDir)
	if err != nil {
		return nil, nil, fmt.Errorf("stat run dir: %w", err)
	}
	if !info.IsDir() {
		return nil, nil, fmt.Errorf("run path is not a directory: %s", runDir)
	}
	files := []string{}
	if err := filepath.WalkDir(runDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".jsonl") {
			return nil
		}
		files = append(files, path)
		return nil
	}); err != nil {
		return nil, nil, fmt.Errorf("walk run logs: %w", err)
	}
	sort.Strings(files)

	entries := []instrumentation.TurnLog{}
	for _, file := range files {
		parsed, err := readJSONL(file)
		if err != nil {
			return nil, nil, err
		}
		entries = append(entries, parsed...)
	}
	return entries, files, nil
}

func readJSONL(path string) ([]instrumentation.TurnLog, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open jsonl %s: %w", path, err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1024), 10*1024*1024)
	out := []instrumentation.TurnLog{}
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var entry instrumentation.TurnLog
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			return nil, fmt.Errorf("decode %s:%d: %w", path, lineNo, err)
		}
		out = append(out, entry)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan jsonl %s: %w", path, err)
	}
	return out, nil
}

func aggregate(entries []instrumentation.TurnLog, files []string) Data {
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].Group != entries[j].Group {
			return entries[i].Group < entries[j].Group
		}
		if entries[i].Rep != entries[j].Rep {
			return entries[i].Rep < entries[j].Rep
		}
		if entries[i].Turn != entries[j].Turn {
			return entries[i].Turn < entries[j].Turn
		}
		return entries[i].Attempt < entries[j].Attempt
	})

	data := Data{
		RunID:        entries[0].RunID,
		ScenarioID:   entries[0].ScenarioID,
		GeneratedAt:  time.Now().Format(time.RFC3339),
		ReasonCounts: map[string]int{},
		SourceFiles:  append([]string(nil), files...),
	}
	groupSet := map[string]bool{}
	entriesByGroup := map[string][]instrumentation.TurnLog{}
	var latencySum int64
	var latencyCount int64
	var lastSuccessful *instrumentation.TurnLog

	for i := range entries {
		entry := entries[i]
		groupSet[entry.Group] = true
		entriesByGroup[entry.Group] = append(entriesByGroup[entry.Group], entry)
		if entry.Status == "" {
			entry.Status = "success"
		}
		if entry.Status == "failed" {
			data.FailedTurns++
			data.HasFailures = true
		} else {
			data.SuccessTurns++
			lastSuccessful = &entries[i]
		}
		if entry.Attempt > 1 {
			data.RetryCount += entry.Attempt - 1
		}
		if entry.LatencyMS > 0 && entry.Status != "failed" {
			latencySum += entry.LatencyMS
			latencyCount++
		}
		keep, discard := countDecisions(entry.MemoryPreEvents)
		data.TotalKeep += keep
		data.TotalDiscard += discard
		if keep+discard > 0 {
			data.HasBFMA = true
		}
		for _, event := range entry.MemoryPreEvents {
			if event.Type == "bfma_decision" && event.Decision == "discard" && event.Reason != "" {
				data.ReasonCounts[event.Reason]++
			}
		}
		memoryCount := snapshotInt(entry.MemoryAfter, "memory_count")
		tokenBudget := snapshotInt(entry.MemoryAfter, "token_budget")
		tokenUsed := parseTokenUsed(entry.MemoryContextInjected)
		label := fmt.Sprintf("%s T%d", entry.Group, entry.Turn)
		summary := TurnSummary{
			Group:       entry.Group,
			Turn:        entry.Turn,
			Status:      entry.Status,
			Attempt:     entry.Attempt,
			MaxAttempts: entry.MaxAttempts,
			LatencyMS:   entry.LatencyMS,
			MemoryCount: memoryCount,
			Keep:        keep,
			Discard:     discard,
			TokenBudget: tokenBudget,
			TokenUsed:   tokenUsed,
			PromptShort: shorten(entry.UserPrompt, 180),
			AnswerShort: shorten(entry.AssistantResponse, 220),
			Error:       entry.Error,
			FinalTurn:   len(entry.FinalQuestions) > 0,
		}
		data.Turns = append(data.Turns, summary)
		data.LatencyByTurn = append(data.LatencyByTurn, ChartPoint{Label: label, Value: entry.LatencyMS, Group: entry.Group})
		data.MemoryByTurn = append(data.MemoryByTurn, ChartPoint{Label: label, Value: int64(memoryCount), Group: entry.Group})
		data.KeepDiscardByTurn = append(data.KeepDiscardByTurn, DecisionPoint{Label: label, Keep: keep, Discard: discard, Group: entry.Group})
		if tokenBudget > 0 {
			data.BudgetByTurn = append(data.BudgetByTurn, BudgetPoint{Label: label, Used: tokenUsed, Budget: tokenBudget, Group: entry.Group})
		}
	}
	data.TotalTurns = len(data.Turns)
	if latencyCount > 0 {
		data.AvgLatencyMS = latencySum / latencyCount
	}
	for group := range groupSet {
		data.Groups = append(data.Groups, group)
	}
	sort.Strings(data.Groups)
	data.GroupSummaries, data.FinalByGroup = buildGroupComparisons(data.Groups, entriesByGroup)
	for _, summary := range data.GroupSummaries {
		data.GroupLatency = append(data.GroupLatency, ChartPoint{Label: strings.ToUpper(summary.Group), Value: summary.AvgLatencyMS, Group: summary.Group})
		data.GroupAnswerSize = append(data.GroupAnswerSize, ChartPoint{Label: strings.ToUpper(summary.Group), Value: int64(summary.FinalAnswerWords), Group: summary.Group})
		data.GroupCoverage = append(data.GroupCoverage, ChartPoint{Label: strings.ToUpper(summary.Group), Value: int64(summary.CoveragePercent), Group: summary.Group})
		contextMetric := summary.FinalContextItems
		if summary.Group == "g1" {
			contextMetric = summary.FinalSummaryLines
		}
		data.GroupContextSize = append(data.GroupContextSize, ChartPoint{Label: strings.ToUpper(summary.Group), Value: int64(contextMetric), Group: summary.Group})
		data.MemoryPressure = append(data.MemoryPressure, MemoryPressure{
			Group:    summary.Group,
			Label:    strings.ToUpper(summary.Group),
			Stored:   summary.FinalMemoryCount,
			Selected: summary.FinalContextItems,
		})
	}
	for reason, count := range data.ReasonCounts {
		data.ReasonDistribution = append(data.ReasonDistribution, Distribution{Label: reason, Value: count})
	}
	sort.Slice(data.ReasonDistribution, func(i, j int) bool {
		return data.ReasonDistribution[i].Value > data.ReasonDistribution[j].Value
	})
	if lastSuccessful != nil {
		data.FinalMemoryCount = snapshotInt(lastSuccessful.MemoryAfter, "memory_count")
		data.FinalContext = lastSuccessful.MemoryContextInjected
		data.FinalAnswer = lastSuccessful.AssistantResponse
		for _, q := range lastSuccessful.FinalQuestions {
			data.FinalQuestions = append(data.FinalQuestions, Question{ID: q.ID, Question: q.Question, ExpectedAnswer: q.ExpectedAnswer})
		}
	}
	data.Findings = buildFindings(data)
	return data
}

func buildGroupComparisons(groups []string, entriesByGroup map[string][]instrumentation.TurnLog) ([]GroupSummary, []GroupFinal) {
	summaries := []GroupSummary{}
	finals := []GroupFinal{}
	for _, group := range groups {
		entries := entriesByGroup[group]
		if len(entries) == 0 {
			continue
		}
		summary := GroupSummary{
			Group:    group,
			Label:    groupLabel(group),
			Strategy: groupStrategy(group),
		}
		var latencySum int64
		var latencyCount int64
		var final *instrumentation.TurnLog
		for i := range entries {
			entry := entries[i]
			if entry.Status == "" {
				entry.Status = "success"
			}
			summary.Turns++
			if entry.Status == "failed" {
				summary.FailedTurns++
			} else {
				summary.SuccessTurns++
				final = &entries[i]
				if entry.LatencyMS > 0 {
					latencySum += entry.LatencyMS
					latencyCount++
				}
			}
			keep, discard := countDecisions(entry.MemoryPreEvents)
			summary.TotalKeep += keep
			summary.TotalDiscard += discard
		}
		if latencyCount > 0 {
			summary.AvgLatencyMS = latencySum / latencyCount
		}
		if final != nil {
			finalKeep, finalDiscard := countDecisions(final.MemoryPreEvents)
			summary.FinalKeep = finalKeep
			summary.FinalDiscard = finalDiscard
			summary.FinalMemoryCount = snapshotInt(final.MemoryAfter, "memory_count")
			summary.FinalSummaryLines = countSummaryLines(final.MemoryAfter)
			summary.FinalContextItems = countContextItems(final.MemoryContextInjected)
			summary.FinalAnswerChars = len([]rune(final.AssistantResponse))
			summary.FinalAnswerWords = len(strings.Fields(final.AssistantResponse))
			summary.TokenBudget = snapshotInt(final.MemoryAfter, "token_budget")
			summary.TokenUsed = parseTokenUsed(final.MemoryContextInjected)
			summary.FinalPromptShort = shorten(final.UserPrompt, 220)
			summary.FinalAnswerShort = shorten(final.AssistantResponse, 260)
			coverage, omissions, detected, total := coverageFor(final.AssistantResponse, final.FinalQuestions)
			summary.Coverage = coverage
			summary.Omissions = omissions
			summary.CoverageDetected = detected
			summary.CoverageTotal = total
			if total > 0 {
				summary.CoveragePercent = int(float64(detected)/float64(total)*100 + 0.5)
			}
			summary.MethodologicalReading = methodologicalReading(summary)
			finals = append(finals, GroupFinal{
				Group:             group,
				Label:             summary.Label,
				FinalContext:      final.MemoryContextInjected,
				FinalAnswer:       final.AssistantResponse,
				Coverage:          coverage,
				Omissions:         omissions,
				FinalContextItems: summary.FinalContextItems,
				FinalAnswerChars:  summary.FinalAnswerChars,
			})
		}
		summaries = append(summaries, summary)
	}
	return summaries, finals
}

func countDecisions(events []memory.Event) (int, int) {
	keep := 0
	discard := 0
	for _, event := range events {
		if event.Type != "bfma_decision" {
			continue
		}
		switch event.Decision {
		case "keep":
			keep++
		case "discard":
			discard++
		}
	}
	return keep, discard
}

func groupLabel(group string) string {
	switch group {
	case "g1":
		return "G1 — Resumen incremental"
	case "g2":
		return "G2 — Memoria persistente"
	case "g3":
		return "G3 — BFMA"
	default:
		return strings.ToUpper(group)
	}
}

func groupStrategy(group string) string {
	switch group {
	case "g1":
		return "Baseline con resumen incremental. Evalúa riesgo de summary drift por compresión sucesiva."
	case "g2":
		return "Memoria persistente aislada. Evalúa acumulación y recuperación amplia de recuerdos."
	case "g3":
		return "BFMA con olvido presupuestado. Evalúa selección de contexto relevante bajo presupuesto."
	default:
		return "Grupo experimental."
	}
}

func methodologicalReading(summary GroupSummary) string {
	switch summary.Group {
	case "g1":
		return "G1 sirve como línea base: conserva un resumen compacto, pero puede perder detalles históricos o mezclar reglas vigentes con reemplazadas."
	case "g2":
		return "G2 conserva memoria amplia: puede recuperar más información, pero también arriesga arrastrar ruido, detalles secundarios o información obsoleta."
	case "g3":
		return fmt.Sprintf("G3 selecciona contexto bajo presupuesto: en el turno final mantuvo %d memorias y descartó %d, con cobertura estimada de %d%%.", summary.FinalKeep, summary.FinalDiscard, summary.CoveragePercent)
	default:
		return "Interpretación no disponible."
	}
}

func countSummaryLines(snapshot memory.Snapshot) int {
	if snapshot == nil {
		return 0
	}
	raw, ok := snapshot["summary_current"].(string)
	if !ok {
		return 0
	}
	count := 0
	for _, line := range strings.Split(raw, "\n") {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}
	return count
}

func countContextItems(context string) int {
	count := 0
	for _, line := range strings.Split(context, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "- [") || strings.HasPrefix(line, "- ") {
			count++
		}
	}
	return count
}

func coverageFor(answer string, questions []scenario.FinalQuestion) ([]CoverageItem, []CoverageItem, int, int) {
	items := []CoverageItem{}
	omissions := []CoverageItem{}
	for _, question := range questions {
		detected, matched, total := expectedAnswerDetected(answer, question.ExpectedAnswer)
		item := CoverageItem{
			ID:             question.ID,
			Question:       question.Question,
			ExpectedAnswer: question.ExpectedAnswer,
			Detected:       detected,
			MatchedTokens:  matched,
			TotalTokens:    total,
		}
		items = append(items, item)
		if !detected {
			omissions = append(omissions, item)
		}
	}
	return items, omissions, len(items) - len(omissions), len(items)
}

func expectedAnswerDetected(answer string, expected string) (bool, int, int) {
	answerNorm := normalizeForCoverage(answer)
	expectedNorm := normalizeForCoverage(expected)
	if expectedNorm == "" {
		return false, 0, 0
	}
	if strings.Contains(answerNorm, expectedNorm) {
		tokens := coverageTokens(expectedNorm)
		if len(tokens) == 0 {
			return true, 1, 1
		}
		return true, len(tokens), len(tokens)
	}
	tokens := coverageTokens(expectedNorm)
	if len(tokens) == 0 {
		return false, 0, 0
	}
	matched := 0
	for _, token := range tokens {
		if strings.Contains(answerNorm, token) {
			matched++
		}
	}
	threshold := 1.0
	if len(tokens) > 2 {
		threshold = 0.6
	}
	return float64(matched)/float64(len(tokens)) >= threshold, matched, len(tokens)
}

func normalizeForCoverage(s string) string {
	replacer := strings.NewReplacer(
		"á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u", "ñ", "n",
		"Á", "a", "É", "e", "Í", "i", "Ó", "o", "Ú", "u", "Ü", "u", "Ñ", "n",
	)
	s = strings.ToLower(replacer.Replace(s))
	parts := coverageWordRE.FindAllString(s, -1)
	return strings.Join(parts, " ")
}

func coverageTokens(s string) []string {
	stop := map[string]bool{
		"con": true, "del": true, "para": true, "por": true, "que": true, "los": true, "las": true,
		"una": true, "uno": true, "debe": true, "deben": true, "como": true, "solo": true, "salvo": true,
	}
	seen := map[string]bool{}
	out := []string{}
	for _, token := range strings.Fields(s) {
		if len([]rune(token)) < 4 || stop[token] || seen[token] {
			continue
		}
		seen[token] = true
		out = append(out, token)
	}
	return out
}

func snapshotInt(snapshot memory.Snapshot, key string) int {
	if snapshot == nil {
		return 0
	}
	switch v := snapshot[key].(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case json.Number:
		i, _ := v.Int64()
		return int(i)
	}
	return 0
}

var (
	tokenUsedRE    = regexp.MustCompile(`usados\s+(\d+)`)
	coverageWordRE = regexp.MustCompile(`[\p{L}\p{N}]+`)
)

func parseTokenUsed(text string) int {
	matches := tokenUsedRE.FindStringSubmatch(text)
	if len(matches) != 2 {
		return 0
	}
	var out int
	fmt.Sscanf(matches[1], "%d", &out)
	return out
}

func shorten(s string, max int) string {
	s = strings.Join(strings.Fields(s), " ")
	if len([]rune(s)) <= max {
		return s
	}
	runes := []rune(s)
	return string(runes[:max]) + "…"
}

func buildFindings(data Data) []string {
	findings := []string{}
	if data.HasBFMA {
		findings = append(findings, fmt.Sprintf("BFMA registró %d decisiones keep y %d decisiones discard durante la corrida.", data.TotalKeep, data.TotalDiscard))
	}
	if data.FinalMemoryCount > 0 {
		findings = append(findings, fmt.Sprintf("La memoria final contiene %d memorias experimentales registradas.", data.FinalMemoryCount))
	}
	if len(data.BudgetByTurn) > 0 {
		last := data.BudgetByTurn[len(data.BudgetByTurn)-1]
		findings = append(findings, fmt.Sprintf("En el último turno con presupuesto BFMA se usaron %d/%d tokens de contexto seleccionado.", last.Used, last.Budget))
	}
	if data.HasFailures {
		findings = append(findings, "La corrida contiene intentos fallidos registrados; revisar la tabla por turno para trazabilidad.")
	} else {
		findings = append(findings, "No se registraron turnos fallidos en los logs analizados.")
	}
	if len(data.Groups) == 1 {
		findings = append(findings, "Este reporte corresponde a una corrida parcial de un solo grupo; para comparación experimental completa generar reportes con G1, G2 y G3.")
	}
	return findings
}
