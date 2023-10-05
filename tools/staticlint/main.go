package main

import (
	"regexp"

	"github.com/alexkohler/nakedret"
	"github.com/breml/bidichk/pkg/bidichk"
	"github.com/kvvPro/metric-collector/tools/staticlint/exitcheck"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift" // импортируем дополнительный анализатор
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/staticcheck"
)

func main() {
	mychecks := []*analysis.Analyzer{
		exitcheck.Analyzer,
		shift.Analyzer,
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		nakedret.NakedReturnAnalyzer(5),
		bidichk.NewAnalyzer(),
	}

	re := regexp.MustCompile(`(SA*|S1*|ST1*)`)
	for _, v := range staticcheck.Analyzers {
		if re.MatchString(v.Analyzer.Name) {
			mychecks = append(mychecks, v.Analyzer)
		}
	}

	multichecker.Main(
		mychecks...,
	)
}
