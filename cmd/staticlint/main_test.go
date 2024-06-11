package main

import (
	"testing"

	gocritic "github.com/go-critic/go-critic/checkers/analyzer"
	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/appends"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/timeformat"
)

func Test_analyzers(t *testing.T) {
	tests := []struct {
		name       string
		wantChecks []*analysis.Analyzer
	}{
		{
			name:       "check contain analyzers 1",
			wantChecks: []*analysis.Analyzer{appends.Analyzer, asmdecl.Analyzer, OsExitCheckAnalyzer},
		},
		{
			name:       "check contain analyzers 2",
			wantChecks: []*analysis.Analyzer{gocritic.Analyzer, bools.Analyzer, timeformat.Analyzer},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, w := range tt.wantChecks {
				assert.Contains(t, analyzers(), w)
			}
		})
	}
}
