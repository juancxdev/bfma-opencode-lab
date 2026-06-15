package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

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
