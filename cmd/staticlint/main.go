package main

import (
	"honnef.co/go/tools/staticcheck"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/ctrlflow"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
)

func main() {
	var analyzers []*analysis.Analyzer
	analyzers = append(analyzers,
		printf.Analyzer,
		shadow.Analyzer,
		bools.Analyzer,
		ctrlflow.Analyzer,
		copylock.Analyzer,
		lostcancel.Analyzer,
	)
	for _, v := range staticcheck.Analyzers {
		analyzers = append(analyzers, v.Analyzer)
	}
	multichecker.Main(analyzers...)
}
