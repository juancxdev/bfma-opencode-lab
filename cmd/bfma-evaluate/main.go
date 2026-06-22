package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"bfma-opencode-lab/internal/evaluation"
)

func main() {
	var (
		runDir      = flag.String("run", "", "run directory containing JSONL logs")
		scenarioDir = flag.String("scenario-dir", "", "directory containing scenario JSON files; defaults to project scenarios dir inferred from --run")
		outDir      = flag.String("out", "", "output directory for metrics.csv, summary.json and conclusion.md")
		printJSON   = flag.Bool("json", false, "print full evaluation summary as JSON")
	)
	flag.Parse()
	if strings.TrimSpace(*runDir) == "" {
		fatal(fmt.Errorf("--run is required"))
	}
	out := *outDir
	if strings.TrimSpace(out) == "" {
		out = filepath.Join("results", filepath.Base(filepath.Clean(*runDir)))
	}
	result, err := evaluation.Evaluate(evaluation.Options{RunDir: *runDir, ScenarioDir: *scenarioDir, OutDir: out})
	if err != nil {
		fatal(err)
	}
	if *printJSON {
		body, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			fatal(err)
		}
		fmt.Println(string(body))
		return
	}
	fmt.Printf("metrics=%s\n", result.MetricsCSVPath)
	fmt.Printf("summary=%s\n", result.SummaryJSONPath)
	fmt.Printf("conclusion=%s\n", result.ConclusionMDPath)
	for _, line := range result.Conclusion {
		fmt.Printf("- %s\n", line)
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
