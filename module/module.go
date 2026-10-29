package module

import (
	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"

	"github.com/KSpaceer/testctx"
)

func init() {
	register.Plugin("testctx", New)
}

// Linter is a register.LinterPlugin implementation for golangci module plugins system
type Linter struct{}

// New creates a module plugin linter
func New(any) (register.LinterPlugin, error) {
	return &Linter{}, nil
}

// BuildAnalyzers builds analyzers used in golangci
func (l *Linter) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{testctx.Analyzer}, nil
}

// GetLoadMode returns load mode which the module plugin uses
func (l *Linter) GetLoadMode() string {
	return register.LoadModeTypesInfo
}
