package testctx_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/KSpaceer/testctx"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, testctx.Analyzer, "test")
}
