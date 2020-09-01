package deepequal2cmp_test

import (
	"testing"

	"deepequal2cmp"

	"golang.org/x/tools/go/analysis/analysistest"
)

// TestAnalyzer is a test for Analyzer.
func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, deepequal2cmp.Analyzer, "a")
}
