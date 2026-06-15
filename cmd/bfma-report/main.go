package main

import (
	"flag"
	"fmt"
	"os"

	"bfma-opencode-lab/internal/report"
)

func main() {
	runDir := flag.String("run", "", "path to a run directory under logs/")
	out := flag.String("out", "", "output HTML file path")
	flag.Parse()

	if err := report.Generate(report.Options{RunDir: *runDir, Out: *out}); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	fmt.Println("report=" + *out)
}
