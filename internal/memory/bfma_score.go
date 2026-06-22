package memory

import "strings"

const BFMAFormulaVersion = "bfma_base_extension_v1"

type BFMAConfig struct {
	TokenBudget  int          `json:"token_budget"`
	KeepMinScore float64      `json:"keep_min_score"`
	Weights      ScoreWeights `json:"weights"`
}

type ScoreWeights struct {
	Relevance    float64 `json:"relevance"`
	Importance   float64 `json:"importance"`
	Recency      float64 `json:"recency"`
	Frequency    float64 `json:"frequency"`
	TokenCost    float64 `json:"token_cost"`
	Obsolescence float64 `json:"obsolescence"`
}

type ScoreComponents struct {
	Relevance    float64 `json:"relevance"`
	Importance   float64 `json:"importance"`
	Recency      float64 `json:"recency"`
	Frequency    float64 `json:"frequency"`
	TokenCost    float64 `json:"token_cost"`
	Obsolescence float64 `json:"obsolescence"`
}

type ScoreBreakdown struct {
	FormulaVersion      string          `json:"formula_version"`
	AntecedentScore     float64         `json:"antecedent_score"`
	BFMAUtility         float64         `json:"bfma_utility"`
	FrequencyBonus      float64         `json:"frequency_bonus"`
	TokenCostPenalty    float64         `json:"token_cost_penalty"`
	ObsolescencePenalty float64         `json:"obsolescence_penalty"`
	Components          ScoreComponents `json:"components"`
	Weights             ScoreWeights    `json:"weights"`
}

type ForgetDecision struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
}

type ObsolescenceAssessment struct {
	Score  float64 `json:"score"`
	Reason string  `json:"reason,omitempty"`
}

func DefaultBFMAConfig() BFMAConfig {
	return BFMAConfig{
		TokenBudget:  220,
		KeepMinScore: 0.28,
		Weights: ScoreWeights{
			Relevance:    0.25,
			Importance:   0.30,
			Recency:      0.15,
			Frequency:    0.10,
			TokenCost:    0.10,
			Obsolescence: 0.10,
		},
	}
}

func NewBFMAConfig(tokenBudget int, keepMinScore float64) BFMAConfig {
	cfg := DefaultBFMAConfig()
	if tokenBudget > 0 {
		cfg.TokenBudget = tokenBudget
	}
	if keepMinScore > 0 {
		cfg.KeepMinScore = keepMinScore
	}
	return cfg.Normalize()
}

func (c BFMAConfig) Normalize() BFMAConfig {
	if c.TokenBudget <= 0 {
		c.TokenBudget = DefaultBFMAConfig().TokenBudget
	}
	if c.KeepMinScore <= 0 {
		c.KeepMinScore = DefaultBFMAConfig().KeepMinScore
	}
	defaults := DefaultBFMAConfig().Weights
	if c.Weights.Relevance == 0 && c.Weights.Importance == 0 && c.Weights.Recency == 0 {
		c.Weights.Relevance = defaults.Relevance
		c.Weights.Importance = defaults.Importance
		c.Weights.Recency = defaults.Recency
	}
	if c.Weights.Frequency == 0 {
		c.Weights.Frequency = defaults.Frequency
	}
	if c.Weights.TokenCost == 0 {
		c.Weights.TokenCost = defaults.TokenCost
	}
	if c.Weights.Obsolescence == 0 {
		c.Weights.Obsolescence = defaults.Obsolescence
	}
	return c
}

func CalculateScore(components ScoreComponents, weights ScoreWeights) ScoreBreakdown {
	components = clampComponents(components)
	antecedent := clamp01(
		weights.Relevance*components.Relevance +
			weights.Importance*components.Importance +
			weights.Recency*components.Recency,
	)
	frequencyBonus := weights.Frequency * components.Frequency
	tokenPenalty := weights.TokenCost * components.TokenCost
	obsolescencePenalty := weights.Obsolescence * components.Obsolescence
	utility := clamp01(antecedent + frequencyBonus - tokenPenalty - obsolescencePenalty)
	return ScoreBreakdown{
		FormulaVersion:      BFMAFormulaVersion,
		AntecedentScore:     round3(antecedent),
		BFMAUtility:         round3(utility),
		FrequencyBonus:      round3(frequencyBonus),
		TokenCostPenalty:    round3(tokenPenalty),
		ObsolescencePenalty: round3(obsolescencePenalty),
		Components:          roundScoreComponents(components),
		Weights:             roundWeights(weights),
	}
}

func DecideForget(utility float64, keepMinScore float64, usedTokens int, cost int, tokenBudget int, obsolescence ObsolescenceAssessment) ForgetDecision {
	if obsolescence.Score >= 0.95 {
		return ForgetDecision{Decision: "discard", Reason: "obsolete_replaced"}
	}
	if utility < keepMinScore {
		return ForgetDecision{Decision: "discard", Reason: "below_min_score"}
	}
	if usedTokens+cost > tokenBudget {
		return ForgetDecision{Decision: "discard", Reason: "token_budget_exceeded"}
	}
	return ForgetDecision{Decision: "keep", Reason: "within_budget"}
}

const replacementOverlapThreshold = 0.35

func AssessObsolescence(record MemoryRecord, all []MemoryRecord) ObsolescenceAssessment {
	content := strings.ToLower(record.Content)
	if containsAny(content, []string{"ya no aplica", "obsoleto", "reemplazada", "reemplazado"}) {
		return ObsolescenceAssessment{Score: 0.75, Reason: "self_marked_obsolete"}
	}
	for _, other := range all {
		obsolete, reason := isObsoleteByReplacement(record, other)
		if obsolete {
			return ObsolescenceAssessment{Score: 1.0, Reason: reason}
		}
	}
	return ObsolescenceAssessment{}
}

func isObsoleteByReplacement(old MemoryRecord, newer MemoryRecord) (bool, string) {
	if newer.ID == old.ID || newer.SourceTurn <= old.SourceTurn {
		return false, "not_obsolete_not_newer"
	}
	if !replacementSignal(newer.Content) {
		return false, "not_obsolete_no_replacement_signal"
	}
	sharedTags := sharedSpecificTags(old.Tags, newer.Tags)
	if len(sharedTags) == 0 {
		return false, "not_obsolete_generic_tag_only"
	}
	if replacementOverlap(
		old.Content+" "+strings.Join(sharedTags, " "),
		newer.Content+" "+strings.Join(sharedTags, " "),
	) < replacementOverlapThreshold {
		return false, "not_obsolete_low_overlap"
	}
	return true, "replaced_by_" + newer.ID
}

func replacementSignal(text string) bool {
	text = strings.ToLower(text)
	return containsAny(text, []string{"ya no", "reemplaza", "reemplaz", "cambia", "cambio", "no aplica", "pasa a"})
}

func sharedSpecificTags(a, b []string) []string {
	seen := map[string]bool{}
	for _, tag := range a {
		normalized := normalizeTag(tag)
		if normalized == "" || isGenericTag(normalized) {
			continue
		}
		seen[normalized] = true
	}
	shared := []string{}
	for _, tag := range b {
		normalized := normalizeTag(tag)
		if normalized == "" || isGenericTag(normalized) || !seen[normalized] {
			continue
		}
		shared = append(shared, normalized)
	}
	return shared
}

func replacementOverlap(old string, newer string) float64 {
	oldTokens := replacementTokenSet(old)
	newTokens := replacementTokenSet(newer)
	if len(oldTokens) == 0 || len(newTokens) == 0 {
		return 0
	}
	matches := 0
	for token := range oldTokens {
		if newTokens[token] {
			matches++
		}
	}
	return clamp01(float64(matches) / float64(len(oldTokens)))
}

func replacementTokenSet(text string) map[string]bool {
	stop := map[string]bool{
		"el": true, "la": true, "los": true, "las": true, "un": true, "una": true,
		"de": true, "del": true, "y": true, "o": true, "que": true, "en": true,
		"con": true, "por": true, "para": true, "se": true, "a": true, "es": true,
		"debe": true, "deben": true, "todo": true, "toda": true, "ahora": true,
		"anterior": true, "regla": true, "aplica": true,
	}
	out := map[string]bool{}
	for _, token := range wordRE.FindAllString(normalizeTag(text), -1) {
		token = normalizeReplacementToken(token)
		if len(token) < 3 || stop[token] {
			continue
		}
		out[token] = true
	}
	return out
}

func normalizeReplacementToken(token string) string {
	switch {
	case strings.HasSuffix(token, "ciones") && len(token) > 8:
		return strings.TrimSuffix(strings.TrimSuffix(token, "ciones"), "a")
	case strings.HasSuffix(token, "cion") && len(token) > 6:
		return strings.TrimSuffix(strings.TrimSuffix(token, "cion"), "a")
	case strings.HasSuffix(token, "ando") && len(token) > 7:
		return strings.TrimSuffix(token, "ando")
	case strings.HasSuffix(token, "iendo") && len(token) > 8:
		return strings.TrimSuffix(token, "iendo")
	case strings.HasSuffix(token, "an") && len(token) > 5:
		return strings.TrimSuffix(token, "an")
	case strings.HasSuffix(token, "ar") && len(token) > 5:
		return strings.TrimSuffix(token, "ar")
	case strings.HasSuffix(token, "er") && len(token) > 5:
		return strings.TrimSuffix(token, "er")
	case strings.HasSuffix(token, "ir") && len(token) > 5:
		return strings.TrimSuffix(token, "ir")
	}
	return token
}

func isGenericTag(tag string) bool {
	generic := map[string]bool{
		"docentes": true, "docente": true, "estudiantes": true, "estudiante": true,
		"arquitectura": true, "academico": true, "académico": true, "datos": true,
		"modulos": true, "módulos": true, "modulo": true, "módulo": true,
		"cambio": true, "cambios": true, "vigente": true, "fact": true,
		"constraint": true, "decision": true,
	}
	return generic[normalizeTag(tag)]
}

func normalizeTag(tag string) string {
	tag = strings.ToLower(strings.TrimSpace(tag))
	replacer := strings.NewReplacer(
		"á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u", "ñ", "n",
	)
	return strings.Join(strings.Fields(replacer.Replace(tag)), " ")
}

func clampComponents(c ScoreComponents) ScoreComponents {
	c.Relevance = clamp01(c.Relevance)
	c.Importance = clamp01(c.Importance)
	c.Recency = clamp01(c.Recency)
	c.Frequency = clamp01(c.Frequency)
	c.TokenCost = clamp01(c.TokenCost)
	c.Obsolescence = clamp01(c.Obsolescence)
	return c
}

func roundScoreComponents(c ScoreComponents) ScoreComponents {
	c.Relevance = round3(c.Relevance)
	c.Importance = round3(c.Importance)
	c.Recency = round3(c.Recency)
	c.Frequency = round3(c.Frequency)
	c.TokenCost = round3(c.TokenCost)
	c.Obsolescence = round3(c.Obsolescence)
	return c
}

func roundWeights(w ScoreWeights) ScoreWeights {
	w.Relevance = round3(w.Relevance)
	w.Importance = round3(w.Importance)
	w.Recency = round3(w.Recency)
	w.Frequency = round3(w.Frequency)
	w.TokenCost = round3(w.TokenCost)
	w.Obsolescence = round3(w.Obsolescence)
	return w
}

func hasSharedTags(a, b []string) bool {
	seen := map[string]bool{}
	for _, tag := range a {
		seen[strings.ToLower(strings.TrimSpace(tag))] = true
	}
	for _, tag := range b {
		if seen[strings.ToLower(strings.TrimSpace(tag))] {
			return true
		}
	}
	return false
}

func containsAny(s string, needles []string) bool {
	for _, needle := range needles {
		if strings.Contains(s, needle) {
			return true
		}
	}
	return false
}
