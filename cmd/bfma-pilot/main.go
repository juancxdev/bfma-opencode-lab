package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"bfma-opencode-lab/internal/memory"
	"bfma-opencode-lab/internal/opencode"
	"bfma-opencode-lab/internal/runner"
)

func main() {
	var (
		scenarioID  = flag.String("scenario", "scenario_01", "scenario id without .json")
		groupsRaw   = flag.String("groups", "g1,g2,g3", "comma-separated groups: g1,g2,g3")
		reps        = flag.Int("reps", 1, "repetitions per group")
		model       = flag.String("model", "", "optional opencode model override provider/model")
		simulateLLM = flag.Bool("simulate-llm", false, "simulate LLM responses; validates runner/memory/logging only")
		dryRunAlias = flag.Bool("dry-run", false, "deprecated alias for --simulate-llm")
		timeout     = flag.Duration("timeout", 5*time.Minute, "timeout per OpenCode turn")
		retries     = flag.Int("retries", 0, "number of retries per OpenCode turn after a failed attempt")
		fromTurn    = flag.Int("from-turn", 0, "resume from this scenario turn after reconstructing prior memory without LLM calls")
	)
	defaultBFMA := memory.DefaultBFMAConfig()
	var (
		bfmaTokenBudget        = flag.Int("bfma-token-budget", defaultBFMA.TokenBudget, "BFMA active-context token budget")
		bfmaKeepMinScore       = flag.Float64("bfma-keep-min-score", defaultBFMA.KeepMinScore, "minimum BFMA utility required to keep a memory in active context")
		bfmaWeightRelevance    = flag.Float64("bfma-weight-relevance", defaultBFMA.Weights.Relevance, "antecedent score weight for relevance")
		bfmaWeightImportance   = flag.Float64("bfma-weight-importance", defaultBFMA.Weights.Importance, "antecedent score weight for importance")
		bfmaWeightRecency      = flag.Float64("bfma-weight-recency", defaultBFMA.Weights.Recency, "antecedent score weight for recency")
		bfmaWeightFrequency    = flag.Float64("bfma-weight-frequency", defaultBFMA.Weights.Frequency, "BFMA extension weight for frequency bonus")
		bfmaWeightTokenCost    = flag.Float64("bfma-weight-token-cost", defaultBFMA.Weights.TokenCost, "BFMA extension weight for token cost penalty")
		bfmaWeightObsolescence = flag.Float64("bfma-weight-obsolescence", defaultBFMA.Weights.Obsolescence, "BFMA extension weight for obsolescence penalty")
	)
	flag.Parse()

	projectDir, err := os.Getwd()
	if err != nil {
		fatal(err)
	}
	groups, err := runner.ParseGroups(*groupsRaw)
	if err != nil {
		fatal(err)
	}
	bfmaConfig := memory.BFMAConfig{
		TokenBudget:  *bfmaTokenBudget,
		KeepMinScore: *bfmaKeepMinScore,
		Weights: memory.ScoreWeights{
			Relevance:    *bfmaWeightRelevance,
			Importance:   *bfmaWeightImportance,
			Recency:      *bfmaWeightRecency,
			Frequency:    *bfmaWeightFrequency,
			TokenCost:    *bfmaWeightTokenCost,
			Obsolescence: *bfmaWeightObsolescence,
		},
	}.Normalize()

	cfg := runner.Config{
		ProjectDir:  projectDir,
		ScenarioDir: filepath.Join(projectDir, "scenarios"),
		LogDir:      filepath.Join(projectDir, "logs"),
		DataDir:     filepath.Join(projectDir, "data"),
		ScenarioID:  *scenarioID,
		Groups:      groups,
		Reps:        *reps,
		Model:       *model,
		DryRun:      *simulateLLM || *dryRunAlias,
		Timeout:     *timeout,
		Retries:     *retries,
		FromTurn:    *fromTurn,
		BFMAConfig:  bfmaConfig,
	}
	r := runner.New(cfg, opencode.NewCLIClient("opencode"))
	result, err := r.Run(context.Background())
	if err != nil {
		fatal(err)
	}
	fmt.Printf("run_id=%s\n", result.RunID)
	for _, logFile := range result.LogFiles {
		fmt.Printf("log=%s\n", logFile)
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
