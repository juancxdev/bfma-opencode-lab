package memory

import (
	"regexp"
	"sort"
	"strings"
)

var wordRE = regexp.MustCompile(`[\p{L}\p{N}]+`)

func estimateTokens(s string) int {
	words := wordRE.FindAllString(s, -1)
	if len(words) == 0 {
		return 0
	}
	return int(float64(len(words))*1.35) + 1
}

func overlapScore(a, b string) float64 {
	wa := tokenSet(a)
	wb := tokenSet(b)
	if len(wa) == 0 || len(wb) == 0 {
		return 0
	}
	matches := 0
	for w := range wa {
		if wb[w] {
			matches++
		}
	}
	den := len(wa)
	if len(wb) > den {
		den = len(wb)
	}
	return clamp01(float64(matches) / float64(den))
}

func tokenSet(s string) map[string]bool {
	stop := map[string]bool{
		"el": true, "la": true, "los": true, "las": true, "un": true, "una": true,
		"de": true, "del": true, "y": true, "o": true, "que": true, "en": true,
		"con": true, "por": true, "para": true, "se": true, "a": true, "es": true,
		"debe": true, "deben": true, "todo": true, "toda": true, "ahora": true,
	}
	out := map[string]bool{}
	for _, w := range wordRE.FindAllString(strings.ToLower(s), -1) {
		if len(w) < 3 || stop[w] {
			continue
		}
		out[w] = true
	}
	return out
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func topByRelevance(records []MemoryRecord, prompt string, max int) []MemoryRecord {
	type scored struct {
		r MemoryRecord
		s float64
	}
	scores := make([]scored, 0, len(records))
	for _, r := range records {
		scores = append(scores, scored{r: r, s: overlapScore(prompt, r.Content) + r.Importance*0.15})
	}
	sort.SliceStable(scores, func(i, j int) bool { return scores[i].s > scores[j].s })
	if max > len(scores) {
		max = len(scores)
	}
	out := make([]MemoryRecord, 0, max)
	for i := 0; i < max; i++ {
		out = append(out, scores[i].r)
	}
	return out
}
