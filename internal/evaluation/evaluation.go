package evaluation

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"bfma-opencode-lab/internal/instrumentation"
	"bfma-opencode-lab/internal/memory"
	"bfma-opencode-lab/internal/scenario"
)

type Options struct {
	RunDir      string
	ScenarioDir string
	OutDir      string
}

type Result struct {
	RunID                string         `json:"run_id"`
	ScenarioID           string         `json:"scenario_id"`
	Rows                 []MetricRow    `json:"rows"`
	Groups               []GroupMetrics `json:"groups"`
	Conclusion           []string       `json:"conclusion"`
	Warnings             []string       `json:"warnings"`
	FormulaVersions      []string       `json:"formula_versions,omitempty"`
	AvgAntecedentScore   float64        `json:"avg_antecedent_score,omitempty"`
	AvgBFMAUtility       float64        `json:"avg_bfma_utility,omitempty"`
	ObsoleteDiscardCount int            `json:"obsolete_discard_count,omitempty"`
	MetricsCSVPath       string         `json:"metrics_csv_path,omitempty"`
	SummaryJSONPath      string         `json:"summary_json_path,omitempty"`
	ConclusionMDPath     string         `json:"conclusion_md_path,omitempty"`
}

type MetricRow struct {
	RunID              string  `json:"run_id"`
	ScenarioID         string  `json:"scenario_id"`
	Architecture       string  `json:"architecture"`
	Rep                int     `json:"replica"`
	Turn               int     `json:"turn"`
	QuestionID         string  `json:"question_id"`
	ExpectedAnswer     string  `json:"expected_answer"`
	F1                 float64 `json:"f1"`
	Precision          float64 `json:"precision"`
	Recall             float64 `json:"recall"`
	SubEM              bool    `json:"subem"`
	SubEMHits          int     `json:"subem_hits"`
	SubEMTotal         int     `json:"subem_total"`
	FalseMemoryHits    int     `json:"false_memory_hits"`
	FalseMemoryTotal   int     `json:"false_memory_total"`
	ContradictionHits  int     `json:"contradiction_hits"`
	ContradictionTotal int     `json:"contradiction_total"`
	MemoryTokens       int     `json:"memory_tokens"`
	OutputTokens       int     `json:"output_tokens"`
	StorageItems       int     `json:"storage_items"`
	LatencyMS          int64   `json:"latency_ms"`
	ValidExecution     bool    `json:"valid_execution"`
	ExclusionReason    string  `json:"exclusion_reason,omitempty"`
}

type GroupMetrics struct {
	Architecture       string  `json:"architecture"`
	Rows               int     `json:"rows"`
	AvgF1              float64 `json:"avg_f1"`
	AvgPrecision       float64 `json:"avg_precision"`
	AvgRecall          float64 `json:"avg_recall"`
	SubEMRate          float64 `json:"subem_rate"`
	FalseMemoryRate    float64 `json:"false_memory_rate"`
	ContradictionRate  float64 `json:"contradiction_rate"`
	AvgMemoryTokens    float64 `json:"avg_memory_tokens"`
	AvgOutputTokens    float64 `json:"avg_output_tokens"`
	AvgStorageItems    float64 `json:"avg_storage_items"`
	AvgLatencyMS       float64 `json:"avg_latency_ms"`
	ValidRows          int     `json:"valid_rows"`
	InvalidRows        int     `json:"invalid_rows"`
	SubEMHits          int     `json:"subem_hits"`
	SubEMTotal         int     `json:"subem_total"`
	FalseMemoryHits    int     `json:"false_memory_hits"`
	FalseMemoryTotal   int     `json:"false_memory_total"`
	ContradictionHits  int     `json:"contradiction_hits"`
	ContradictionTotal int     `json:"contradiction_total"`
	Interpretation     string  `json:"interpretation"`
}

type extendedScenario struct {
	scenario.Scenario
	Evaluation struct {
		FalseMemoryTraps []string `json:"false_memory_traps"`
		SubEMEntities    []string `json:"subem_entities"`
	} `json:"evaluation"`
}

func Evaluate(opts Options) (Result, error) {
	if strings.TrimSpace(opts.RunDir) == "" {
		return Result{}, fmt.Errorf("run dir is required")
	}
	entries, err := readRunLogs(opts.RunDir)
	if err != nil {
		return Result{}, err
	}
	if len(entries) == 0 {
		return Result{}, fmt.Errorf("no JSONL turn logs found under %s", opts.RunDir)
	}
	if strings.TrimSpace(opts.ScenarioDir) == "" {
		opts.ScenarioDir = inferScenarioDir(opts.RunDir)
	}
	sc, err := loadExtendedScenario(opts.ScenarioDir, entries[0].ScenarioID)
	if err != nil {
		return Result{}, err
	}
	res := buildResult(entries, sc)
	if strings.TrimSpace(opts.OutDir) != "" {
		if err := os.MkdirAll(opts.OutDir, 0o755); err != nil {
			return Result{}, fmt.Errorf("create output dir: %w", err)
		}
		if err := writeArtifacts(opts.OutDir, &res); err != nil {
			return Result{}, err
		}
	}
	return res, nil
}

func buildResult(entries []instrumentation.TurnLog, sc extendedScenario) Result {
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].Group != entries[j].Group {
			return entries[i].Group < entries[j].Group
		}
		if entries[i].Rep != entries[j].Rep {
			return entries[i].Rep < entries[j].Rep
		}
		return entries[i].Turn < entries[j].Turn
	})
	res := Result{RunID: entries[0].RunID, ScenarioID: entries[0].ScenarioID}
	res.AvgAntecedentScore, res.AvgBFMAUtility, res.ObsoleteDiscardCount, res.FormulaVersions = bfmaTraceSummary(entries)
	finals := finalSuccessfulEntries(entries)
	for _, entry := range finals {
		questions := entry.FinalQuestions
		if len(questions) == 0 {
			questions = sc.FinalQuestions
		}
		if len(questions) == 0 {
			res.Warnings = append(res.Warnings, fmt.Sprintf("%s rep %d no contiene preguntas finales", entry.Group, entry.Rep))
			continue
		}
		for _, q := range questions {
			subEMHits := countDetected(entry.AssistantResponse, sc.Evaluation.SubEMEntities)
			falseMemoryHits := countDetected(entry.AssistantResponse, sc.Evaluation.FalseMemoryTraps)
			contradictionHits := countDetected(entry.AssistantResponse, sc.GroundTruth.ExpectedContradictions)
			precision, recall, f1 := factualF1(entry.AssistantResponse, q.ExpectedAnswer, falseMemoryHits+contradictionHits)
			row := MetricRow{
				RunID:              entry.RunID,
				ScenarioID:         entry.ScenarioID,
				Architecture:       entry.Group,
				Rep:                entry.Rep,
				Turn:               entry.Turn,
				QuestionID:         q.ID,
				ExpectedAnswer:     q.ExpectedAnswer,
				F1:                 round4(f1),
				Precision:          round4(precision),
				Recall:             round4(recall),
				SubEM:              subEMHits == len(sc.Evaluation.SubEMEntities) && len(sc.Evaluation.SubEMEntities) > 0,
				SubEMHits:          subEMHits,
				SubEMTotal:         len(sc.Evaluation.SubEMEntities),
				FalseMemoryHits:    falseMemoryHits,
				FalseMemoryTotal:   len(sc.Evaluation.FalseMemoryTraps),
				ContradictionHits:  contradictionHits,
				ContradictionTotal: len(sc.GroundTruth.ExpectedContradictions),
				MemoryTokens:       estimateTokens(entry.MemoryContextInjected),
				OutputTokens:       estimateTokens(entry.AssistantResponse),
				StorageItems:       storageItems(entry.MemoryAfter),
				LatencyMS:          entry.LatencyMS,
				ValidExecution:     entry.Status == "" || entry.Status == "success",
			}
			if !row.ValidExecution {
				row.ExclusionReason = entry.Error
			}
			res.Rows = append(res.Rows, row)
		}
	}
	if len(sc.Evaluation.FalseMemoryTraps) == 0 {
		res.Warnings = append(res.Warnings, "El escenario no define evaluation.false_memory_traps; la tasa de falsas memorias queda en 0 por falta de trampas explícitas.")
	}
	if len(sc.Evaluation.SubEMEntities) == 0 {
		res.Warnings = append(res.Warnings, "El escenario no define evaluation.subem_entities; SubEM no puede interpretarse como recuperación exacta de entidades críticas.")
	}
	res.Groups = summarizeGroups(res.Rows)
	res.Conclusion = conclusionFor(res.Groups, res.Warnings)
	return res
}

func bfmaTraceSummary(entries []instrumentation.TurnLog) (float64, float64, int, []string) {
	var antecedentSum float64
	var utilitySum float64
	var count float64
	obsoleteDiscard := 0
	versions := []string{}
	for _, entry := range entries {
		for _, event := range entry.MemoryPreEvents {
			if event.Type != "bfma_decision" {
				continue
			}
			if event.AntecedentScore > 0 {
				antecedentSum += event.AntecedentScore
				count++
			}
			if event.BFMAUtility > 0 {
				utilitySum += event.BFMAUtility
			} else if event.UtilityScore > 0 {
				utilitySum += event.UtilityScore
			}
			if event.Reason == "obsolete_replaced" {
				obsoleteDiscard++
			}
			if event.FormulaVersion != "" {
				versions = appendUniqueString(versions, event.FormulaVersion)
			}
		}
	}
	if count == 0 {
		return 0, 0, obsoleteDiscard, versions
	}
	return round4(antecedentSum / count), round4(utilitySum / count), obsoleteDiscard, versions
}

func appendUniqueString(existing []string, value string) []string {
	for _, item := range existing {
		if item == value {
			return existing
		}
	}
	return append(existing, value)
}

func finalSuccessfulEntries(entries []instrumentation.TurnLog) []instrumentation.TurnLog {
	byKey := map[string]instrumentation.TurnLog{}
	for _, entry := range entries {
		if entry.Status == "failed" {
			continue
		}
		key := fmt.Sprintf("%s/%d", entry.Group, entry.Rep)
		prev, ok := byKey[key]
		if !ok || entry.Turn > prev.Turn || len(entry.FinalQuestions) > 0 {
			byKey[key] = entry
		}
	}
	out := make([]instrumentation.TurnLog, 0, len(byKey))
	for _, entry := range byKey {
		out = append(out, entry)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Group != out[j].Group {
			return out[i].Group < out[j].Group
		}
		return out[i].Rep < out[j].Rep
	})
	return out
}

func summarizeGroups(rows []MetricRow) []GroupMetrics {
	byGroup := map[string][]MetricRow{}
	for _, row := range rows {
		byGroup[row.Architecture] = append(byGroup[row.Architecture], row)
	}
	groups := make([]string, 0, len(byGroup))
	for group := range byGroup {
		groups = append(groups, group)
	}
	sort.Strings(groups)
	out := make([]GroupMetrics, 0, len(groups))
	for _, group := range groups {
		items := byGroup[group]
		gm := GroupMetrics{Architecture: group, Rows: len(items)}
		var f1, precision, recall, subemHits, subemTotal, falseHits, falseTotal, contradictionHits, contradictionTotal, memTokens, outTokens, storage, latency float64
		for _, row := range items {
			if row.ValidExecution {
				gm.ValidRows++
			} else {
				gm.InvalidRows++
			}
			f1 += row.F1
			precision += row.Precision
			recall += row.Recall
			subemHits += float64(row.SubEMHits)
			subemTotal += float64(row.SubEMTotal)
			falseHits += float64(row.FalseMemoryHits)
			falseTotal += float64(row.FalseMemoryTotal)
			contradictionHits += float64(row.ContradictionHits)
			contradictionTotal += float64(row.ContradictionTotal)
			memTokens += float64(row.MemoryTokens)
			outTokens += float64(row.OutputTokens)
			storage += float64(row.StorageItems)
			latency += float64(row.LatencyMS)
		}
		den := math.Max(float64(len(items)), 1)
		gm.AvgF1 = round4(f1 / den)
		gm.AvgPrecision = round4(precision / den)
		gm.AvgRecall = round4(recall / den)
		if subemTotal > 0 {
			gm.SubEMRate = round4(subemHits / subemTotal)
		}
		if len(items) > 0 {
			gm.SubEMHits = items[0].SubEMHits
			gm.SubEMTotal = items[0].SubEMTotal
			gm.FalseMemoryHits = items[0].FalseMemoryHits
			gm.FalseMemoryTotal = items[0].FalseMemoryTotal
			gm.ContradictionHits = items[0].ContradictionHits
			gm.ContradictionTotal = items[0].ContradictionTotal
		}
		if falseTotal > 0 {
			gm.FalseMemoryRate = round4(falseHits / falseTotal)
		}
		if contradictionTotal > 0 {
			gm.ContradictionRate = round4(contradictionHits / contradictionTotal)
		}
		gm.AvgMemoryTokens = round4(memTokens / den)
		gm.AvgOutputTokens = round4(outTokens / den)
		gm.AvgStorageItems = round4(storage / den)
		gm.AvgLatencyMS = round4(latency / den)
		gm.Interpretation = interpretationFor(gm)
		out = append(out, gm)
	}
	return out
}

func interpretationFor(g GroupMetrics) string {
	switch g.Architecture {
	case "g1":
		return "Resumen incremental: línea base para observar pérdida factual por compresión sucesiva."
	case "g2":
		return "Memoria persistente acumulativa: línea comparadora para observar recuperación amplia con riesgo de ruido u obsolescencia."
	case "g3":
		return "Arquitectura BFMA: modalidad experimental; selecciona contexto por utilidad bajo presupuesto de tokens."
	default:
		return "Arquitectura evaluada."
	}
}

func conclusionFor(groups []GroupMetrics, warnings []string) []string {
	out := []string{
		"Este análisis convierte los logs del demo técnico en métricas exploratorias alineadas con la matriz de consistencia: F1 factual, SubEM, falsas memorias, contradicciones, tokens, almacenamiento y latencia.",
		"Los valores se calculan sobre respuestas finales y ground truth del escenario; por tanto sirven para validar el flujo de medición, no para contrastar la hipótesis confirmatoria.",
		"El contraste formal requiere múltiples escenarios, réplicas emparejadas, orden aleatorizado y el modelo mixto F1 ~ arquitectura + orden + (1 | escenario/réplica).",
	}
	by := map[string]GroupMetrics{}
	for _, g := range groups {
		by[g.Architecture] = g
	}
	g3, hasG3 := by["g3"]
	g1, hasG1 := by["g1"]
	g2, hasG2 := by["g2"]
	if hasG3 && hasG1 {
		out = append(out, fmt.Sprintf("Comparación exploratoria BFMA vs resumen incremental: F1 promedio %.3f vs %.3f.", g3.AvgF1, g1.AvgF1))
	}
	if hasG3 && hasG2 {
		out = append(out, fmt.Sprintf("Comparación exploratoria BFMA vs memoria persistente: F1 promedio %.3f vs %.3f.", g3.AvgF1, g2.AvgF1))
	}
	if len(warnings) > 0 {
		out = append(out, "Existen advertencias de instrumentación; revisar summary.json antes de usar el resultado como anexo.")
	}
	return out
}

func writeArtifacts(outDir string, res *Result) error {
	metricsPath := filepath.Join(outDir, "metrics.csv")
	if err := writeMetricsCSV(metricsPath, res.Rows); err != nil {
		return err
	}
	res.MetricsCSVPath = metricsPath
	summaryPath := filepath.Join(outDir, "summary.json")
	body, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal summary: %w", err)
	}
	if err := os.WriteFile(summaryPath, append(body, '\n'), 0o644); err != nil {
		return fmt.Errorf("write summary: %w", err)
	}
	res.SummaryJSONPath = summaryPath
	conclusionPath := filepath.Join(outDir, "conclusion.md")
	if err := os.WriteFile(conclusionPath, []byte(renderConclusionMarkdown(*res)), 0o644); err != nil {
		return fmt.Errorf("write conclusion: %w", err)
	}
	res.ConclusionMDPath = conclusionPath
	return nil
}

func writeMetricsCSV(path string, rows []MetricRow) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create metrics csv: %w", err)
	}
	defer file.Close()
	w := csv.NewWriter(file)
	defer w.Flush()
	headers := []string{"run_id", "scenario_id", "architecture", "replica", "turn", "question_id", "expected_answer", "f1", "precision", "recall", "subem", "subem_hits", "subem_total", "false_memory_hits", "false_memory_total", "contradiction_hits", "contradiction_total", "memory_tokens", "output_tokens", "storage_items", "latency_ms", "valid_execution", "exclusion_reason"}
	if err := w.Write(headers); err != nil {
		return err
	}
	for _, row := range rows {
		record := []string{
			row.RunID, row.ScenarioID, row.Architecture, fmt.Sprint(row.Rep), fmt.Sprint(row.Turn), row.QuestionID, row.ExpectedAnswer,
			fmt.Sprintf("%.4f", row.F1), fmt.Sprintf("%.4f", row.Precision), fmt.Sprintf("%.4f", row.Recall), fmt.Sprint(row.SubEM),
			fmt.Sprint(row.SubEMHits), fmt.Sprint(row.SubEMTotal), fmt.Sprint(row.FalseMemoryHits), fmt.Sprint(row.FalseMemoryTotal), fmt.Sprint(row.ContradictionHits), fmt.Sprint(row.ContradictionTotal),
			fmt.Sprint(row.MemoryTokens), fmt.Sprint(row.OutputTokens), fmt.Sprint(row.StorageItems), fmt.Sprint(row.LatencyMS), fmt.Sprint(row.ValidExecution), row.ExclusionReason,
		}
		if err := w.Write(record); err != nil {
			return err
		}
	}
	return w.Error()
}

func renderConclusionMarkdown(res Result) string {
	var b strings.Builder
	b.WriteString("# Evaluación exploratoria del demo técnico BFMA\n\n")
	b.WriteString(fmt.Sprintf("- Run: `%s`\n- Escenario: `%s`\n\n", res.RunID, res.ScenarioID))
	b.WriteString("## Lectura metodológica\n\n")
	if len(res.FormulaVersions) > 0 {
		b.WriteString(fmt.Sprintf("- Fórmula BFMA registrada: `%s`; score antecedente promedio: %.3f; utilidad BFMA promedio: %.3f; descartes por obsolescencia: %d.\n", strings.Join(res.FormulaVersions, ", "), res.AvgAntecedentScore, res.AvgBFMAUtility, res.ObsoleteDiscardCount))
	}

	for _, line := range res.Conclusion {
		b.WriteString("- ")
		b.WriteString(line)
		b.WriteString("\n")
	}
	b.WriteString("\n## Resumen por arquitectura\n\n")
	b.WriteString("| Arquitectura | F1 | Recall | SubEM entidades | Falsas memorias detectadas | Contradicciones detectadas | Tokens memoria | Tokens salida | Latencia ms |\n")
	b.WriteString("|---|---:|---:|---:|---:|---:|---:|---:|---:|\n")
	for _, g := range res.Groups {
		b.WriteString(fmt.Sprintf("| %s | %.4f | %.4f | %d/%d (%.4f) | %d/%d (%.4f) | %d/%d (%.4f) | %.1f | %.1f | %.1f |\n", g.Architecture, g.AvgF1, g.AvgRecall, g.SubEMHits, g.SubEMTotal, g.SubEMRate, g.FalseMemoryHits, g.FalseMemoryTotal, g.FalseMemoryRate, g.ContradictionHits, g.ContradictionTotal, g.ContradictionRate, g.AvgMemoryTokens, g.AvgOutputTokens, g.AvgLatencyMS))
	}
	if len(res.Warnings) > 0 {
		b.WriteString("\n## Advertencias\n\n")
		for _, warning := range res.Warnings {
			b.WriteString("- ")
			b.WriteString(warning)
			b.WriteString("\n")
		}
	}
	return b.String()
}

func factualF1(answer, expected string, falsePositiveCount int) (float64, float64, float64) {
	answerTokens := tokenMultiset(answer)
	expectedTokens := tokenMultiset(expected)
	if len(expectedTokens) == 0 || len(answerTokens) == 0 {
		return 0, 0, 0
	}
	answerCounts := counts(answerTokens)
	matched := 0
	for _, token := range expectedTokens {
		if answerCounts[token] > 0 {
			matched++
			answerCounts[token]--
		}
	}
	precisionDenominator := matched + falsePositiveCount
	precision := 0.0
	if precisionDenominator > 0 {
		precision = float64(matched) / float64(precisionDenominator)
	}
	recall := float64(matched) / float64(len(expectedTokens))
	if precision+recall == 0 {
		return precision, recall, 0
	}
	return precision, recall, 2 * precision * recall / (precision + recall)
}

func substringExactMatch(answer, expected string) bool {
	expectedNorm := normalize(expected)
	if expectedNorm == "" {
		return false
	}
	return strings.Contains(normalize(answer), expectedNorm)
}

func countDetected(answer string, phrases []string) int {
	answerNorm := normalize(answer)
	count := 0
	for _, phrase := range phrases {
		if phrase = normalize(phrase); phrase != "" && strings.Contains(answerNorm, phrase) {
			count++
		}
	}
	return count
}

func tokenMultiset(s string) []string {
	stop := map[string]bool{"con": true, "del": true, "para": true, "por": true, "que": true, "los": true, "las": true, "una": true, "uno": true, "debe": true, "deben": true, "como": true, "solo": true, "salvo": true, "esta": true, "este": true, "sera": true, "seran": true}
	parts := wordRE.FindAllString(normalize(s), -1)
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if len([]rune(part)) < 4 || stop[part] {
			continue
		}
		out = append(out, part)
	}
	return out
}

func counts(tokens []string) map[string]int {
	out := map[string]int{}
	for _, token := range tokens {
		out[token]++
	}
	return out
}

func normalize(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		switch r {
		case 'á', 'à', 'ä', 'â':
			r = 'a'
		case 'é', 'è', 'ë', 'ê':
			r = 'e'
		case 'í', 'ì', 'ï', 'î':
			r = 'i'
		case 'ó', 'ò', 'ö', 'ô':
			r = 'o'
		case 'ú', 'ù', 'ü', 'û':
			r = 'u'
		case 'ñ':
			r = 'n'
		}
		if unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsSpace(r) {
			b.WriteRune(r)
		} else {
			b.WriteRune(' ')
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

func storageItems(snapshot memory.Snapshot) int {
	if snapshot == nil {
		return 0
	}
	if v, ok := snapshot["memory_count"]; ok {
		switch typed := v.(type) {
		case int:
			return typed
		case int64:
			return int(typed)
		case float64:
			return int(typed)
		case json.Number:
			i, _ := typed.Int64()
			return int(i)
		}
	}
	if raw, ok := snapshot["summary_current"].(string); ok {
		count := 0
		for _, line := range strings.Split(raw, "\n") {
			if strings.TrimSpace(line) != "" {
				count++
			}
		}
		return count
	}
	return 0
}

func estimateTokens(s string) int {
	words := wordRE.FindAllString(s, -1)
	if len(words) == 0 {
		return 0
	}
	return int(float64(len(words))*1.35) + 1
}

func round4(v float64) float64 { return math.Round(v*10000) / 10000 }

func readRunLogs(runDir string) ([]instrumentation.TurnLog, error) {
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
		return nil, fmt.Errorf("walk run logs: %w", err)
	}
	sort.Strings(files)
	entries := []instrumentation.TurnLog{}
	for _, path := range files {
		file, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("open jsonl %s: %w", path, err)
		}
		scanner := bufio.NewScanner(file)
		scanner.Buffer(make([]byte, 1024), 10*1024*1024)
		lineNo := 0
		for scanner.Scan() {
			lineNo++
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			var entry instrumentation.TurnLog
			if err := json.Unmarshal([]byte(line), &entry); err != nil {
				file.Close()
				return nil, fmt.Errorf("decode %s:%d: %w", path, lineNo, err)
			}
			entries = append(entries, entry)
		}
		if err := scanner.Err(); err != nil {
			file.Close()
			return nil, fmt.Errorf("scan jsonl %s: %w", path, err)
		}
		file.Close()
	}
	return entries, nil
}

func loadExtendedScenario(dir, id string) (extendedScenario, error) {
	path := filepath.Join(dir, id+".json")
	body, err := os.ReadFile(path)
	if err != nil {
		return extendedScenario{}, fmt.Errorf("read scenario %q: %w", path, err)
	}
	var out extendedScenario
	if err := json.Unmarshal(body, &out); err != nil {
		return extendedScenario{}, fmt.Errorf("decode scenario %q: %w", path, err)
	}
	if out.ID == "" {
		return extendedScenario{}, fmt.Errorf("scenario %q missing scenario_id", path)
	}
	return out, nil
}

func inferScenarioDir(runDir string) string {
	clean := filepath.Clean(runDir)
	if filepath.Base(filepath.Dir(clean)) == "logs" {
		return filepath.Join(filepath.Dir(filepath.Dir(clean)), "scenarios")
	}
	return "scenarios"
}

var wordRE = regexp.MustCompile(`[\p{L}\p{N}]+`)
