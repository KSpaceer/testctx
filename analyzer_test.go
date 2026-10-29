package testctx_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/KSpaceer/testctx"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.RunWithSuggestedFixes(t, testdata, testctx.Analyzer, "test")
}

func TestContainerSuggestedFixes(t *testing.T) {
	analysistest.RunWithSuggestedFixes(t, analysistest.TestData(), testctx.Analyzer,
		"containers", "indirectcontainer")
}

func TestAccessorSuggestedFixes(t *testing.T) {
	analysistest.RunWithSuggestedFixes(t, analysistest.TestData(), testctx.Analyzer, "accessors")
}

func TestRemoveUnusedImport(t *testing.T) {
	analysistest.RunWithSuggestedFixes(t, analysistest.TestData(), testctx.Analyzer, "unusedimport")
}
