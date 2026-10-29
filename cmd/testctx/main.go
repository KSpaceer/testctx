package main

import (
	"golang.org/x/tools/go/analysis/unitchecker"

	"github.com/KSpaceer/testctx"
)

func main() {
	unitchecker.Main(testctx.Analyzer)
}
