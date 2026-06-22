package main

// make static-lint
import (
	"github.com/Vla8islav/metrics-aggregator/internal/analyzer/noosexit"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"

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
		noosexit.Analyzer,
	)
	for _, v := range staticcheck.Analyzers {
		analyzers = append(analyzers, v.Analyzer)
	}
	//не менее одного анализатора остальных классов пакета staticcheck.io;
	analyzers = append(analyzers, simple.Analyzers[0].Analyzer)
	analyzers = append(analyzers, stylecheck.Analyzers[0].Analyzer)
	analyzers = append(analyzers, quickfix.Analyzers[0].Analyzer)

	multichecker.Main(analyzers...)
}
