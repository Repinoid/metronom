// пакет статического анализа кода
package main

import (
	// Package analysis defines the interface between a modular static analysis and an analysis driver program.
	"golang.org/x/tools/go/analysis"
	// Package multichecker defines the main function for an analysis driver with several analyzers
	"golang.org/x/tools/go/analysis/multichecker"
	// Package defers defines an Analyzer that checks for common mistakes in defer statements.
	"golang.org/x/tools/go/analysis/passes/defers"
	// Package loopclosure defines an Analyzer that checks for references to enclosing loop variables from within nested functions.
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	// Package printf defines an Analyzer that checks consistency of Printf format strings and arguments.
	"golang.org/x/tools/go/analysis/passes/printf"
	// Package shadow defines an Analyzer that checks for shadowed variables.
	"golang.org/x/tools/go/analysis/passes/shadow"
	// импортируем дополнительный анализатор
	"golang.org/x/tools/go/analysis/passes/shift"
	// Package structtag defines an Analyzer that checks struct field tags are well formed.
	"golang.org/x/tools/go/analysis/passes/structtag"
	// Package staticcheck contains analyzes that find bugs and performance issues.
	"honnef.co/go/tools/staticcheck"
)

var ExitCheckAnalyzer = &analysis.Analyzer{
	Name: "exitchack",
	Doc:  "check for direct os.Exit call in main function",
	Run:  runner,
}

func main() {

	mychecks := []*analysis.Analyzer{
		ExitCheckAnalyzer,
		printf.Analyzer,
		shadow.Analyzer,
		shift.Analyzer, 
		defers.Analyzer,
		loopclosure.Analyzer,

		structtag.Analyzer,
	}
	checks := map[string]bool{
		"SA":    true,	// staticcheck, all SA-s
		"S1006": true,  // Use for { ... } for infinite loops
	}

	for _, v := range staticcheck.Analyzers {
		// добавляем в массив нужные проверки
		if checks[v.Analyzer.Name] {
			mychecks = append(mychecks, v.Analyzer)
		}
	}

	multichecker.Main(
		mychecks...,
	)
}
