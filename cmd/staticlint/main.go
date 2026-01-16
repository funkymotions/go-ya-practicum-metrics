// Package main wires up a custom static analysis tool using
// golang.org/x/tools/go/analysis/multichecker. It includes:
//   - printf, shadow, structtag: standard analyzers from x/tools
//   - SA* and S1000: correctness and simplification checks from staticcheck
//   - bodyclose: ensures http.Response.Body is properly closed
//   - sqlrows: ensures *sql.Rows are closed and errors are checked
//   - noexit: project-specific rule forbidding os.Exit in main packages
package main

import (
	"github.com/funkymotions/go-ya-practicum-metrics/cmd/staticlint/noexit"
	"github.com/gostaticanalysis/sqlrows/passes/sqlrows"
	"github.com/timakin/bodyclose/passes/bodyclose"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
)

func main() {
	// analyzers from go/analysis package
	analyzers := []*analysis.Analyzer{
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		noexit.NoExitCodeAnalyzer,
		bodyclose.Analyzer,
		sqlrows.Analyzer,
	}

	// add all SA-class analyzers from staticcheck
	for _, a := range staticcheck.Analyzers {
		analyzers = append(analyzers, a.Analyzer)
	}

	for _, a := range simple.Analyzers {
		if a.Analyzer.Name == "S1000" {
			analyzers = append(analyzers, a.Analyzer)
			break
		}
	}

	multichecker.Main(analyzers...)
}
