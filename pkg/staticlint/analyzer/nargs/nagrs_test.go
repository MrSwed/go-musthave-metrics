package nargs

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestNArgsAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), NArgsAnalyzer, "./...")
}
