package main

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/defers" // импортируем дополнительный анализатор
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift" // импортируем дополнительный анализатор
	"golang.org/x/tools/go/analysis/passes/structtag"
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
		shift.Analyzer, // добавляем анализатор в вызов multichecker
		defers.Analyzer,
		loopclosure.Analyzer,

		structtag.Analyzer,
	}
	checks := map[string]bool{
		"SA":    true,
		"S1006": true,
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
