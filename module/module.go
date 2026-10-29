package module

import (
	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"

	"github.com/KSpaceer/testctx"
)

func init() {
	register.Plugin("testctx", New)
}

type Linter struct{}

func New(any) (register.LinterPlugin, error) {
	return &Linter{}, nil
}

func (l *Linter) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{testctx.Analyzer}, nil
}

func (l *Linter) GetLoadMode() string {
	return register.LoadModeTypesInfo
}
