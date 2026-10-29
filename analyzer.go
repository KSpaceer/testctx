package testctx

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/ssa"
)

var Analyzer = &analysis.Analyzer{
	Name:     "testctx",
	Doc:      "find context.Background()/context.TODO() usages in tests",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer, buildssa.Analyzer},
}

func run(pass *analysis.Pass) (any, error) {
	tbIface, ok := getTestingTBInterface(pass)
	if !ok {
		return nil, nil
	}

	inspectResult := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	astFilter := []ast.Node{(*ast.CallExpr)(nil)}

	inspectResult.Preorder(astFilter, func(node ast.Node) {
		callExpr, ok := node.(*ast.CallExpr)
		if !ok {
			return
		}

		var ident *ast.Ident
		switch fun := callExpr.Fun.(type) {
		case *ast.Ident:
			ident = fun
		case *ast.SelectorExpr:
			ident = fun.Sel
		default:
			return
		}

		obj := pass.TypesInfo.ObjectOf(ident)
		if obj == nil {
			return
		}

		fn, ok := obj.(*types.Func)
		if !ok {
			return
		}
		if !isBasicContextFunction(fn) {
			return
		}

		innerScope := pass.Pkg.Scope().Innermost(callExpr.Pos())
		if innerScope == nil {
			return
		}

		var tbVarName string
	tbVarSearch:
		for s := innerScope; s != nil; s = s.Parent() {
			for _, name := range s.Names() {
				scopeObj := s.Lookup(name)

				variable, ok := scopeObj.(*types.Var)
				if !ok {
					continue
				}

				varType := variable.Type()

				if types.Implements(varType, tbIface) {
					tbVarName = variable.Name()
					break tbVarSearch
				}

				if _, ok := varType.(*types.Pointer); !ok {
					ptrType := types.NewPointer(varType)
					if types.Implements(ptrType, tbIface) {
						tbVarName = variable.Name()
						break tbVarSearch
					}
				}
			}
		}

		if tbVarName == "" {
			return
		}

		// calls of context.Background/context.TODO
		// are fine within testing.TB.Cleanup because testing.TB.Context is already cancelled
		if isWithinCleanup(pass, callExpr) {
			return
		}

		pass.Report(analysis.Diagnostic{
			Pos:      callExpr.Pos(),
			End:      callExpr.End(),
			Category: "context",
			Message:  "using context.Background/context.TODO when testing.TB is present",
			SuggestedFixes: []analysis.SuggestedFix{
				{
					TextEdits: []analysis.TextEdit{
						{
							Pos:     callExpr.Pos(),
							End:     callExpr.End(),
							NewText: []byte(tbVarName + ".Context()"),
						},
					},
				},
			},
		})
	})

	return nil, nil
}

var basicContextFunctions = [...][2]string{
	{"context", "Background"},
	{"context", "TODO"},
}

func isBasicContextFunction(fn *types.Func) bool {
	for _, tuple := range basicContextFunctions {
		if fn.Pkg() == nil || fn.Pkg().Path() != tuple[0] {
			continue
		}
		if fn.Name() != tuple[1] {
			continue
		}
		return true
	}
	return false
}

const (
	testingPackagePath     = "testing"
	testingTBInterfaceName = "TB"
	testingCleanupMethod   = "Cleanup"
)

func getTestingTBInterface(pass *analysis.Pass) (*types.Interface, bool) {
	for _, imp := range pass.Pkg.Imports() {
		if imp.Path() == testingPackagePath {
			obj := imp.Scope().Lookup(testingTBInterfaceName)
			if typeName, ok := obj.(*types.TypeName); ok {
				if iface, ok := typeName.Type().Underlying().(*types.Interface); ok {
					return iface, true
				}
			}
		}
	}
	return nil, false
}

func isWithinCleanup(pass *analysis.Pass, call *ast.CallExpr) bool {
	ssaResult := pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA)

	ssaInstr := callToSSA(ssaResult, call)
	if ssaInstr == nil {
		return false
	}

	return isInstructionWithinCleanup(ssaInstr)
}

func callToSSA(ssaResult *buildssa.SSA, call *ast.CallExpr) ssa.Instruction {
	pos := call.Pos()

	var enclosingFn *ssa.Function
	for _, fn := range ssaResult.SrcFuncs {
		if pos >= fn.Pos() && pos <= fn.Syntax().End() {
			enclosingFn = fn
			break
		}
	}
	if enclosingFn == nil {
		return nil
	}
	enclosingFn = findDeepestEnclosingFunc(enclosingFn, pos)

	for _, block := range enclosingFn.Blocks {
		for _, instr := range block.Instrs {
			if instr.Pos() == call.Lparen {
				return instr
			}
		}
	}
	return nil
}

func findDeepestEnclosingFunc(fn *ssa.Function, pos token.Pos) *ssa.Function {
	for _, anon := range fn.AnonFuncs {
		if pos >= anon.Pos() && pos <= anon.Syntax().End() {
			return findDeepestEnclosingFunc(anon, pos)
		}
	}
	return fn
}

func isInstructionWithinCleanup(instr ssa.Instruction) bool {
	block := instr.Block()
	if block == nil {
		return false
	}
	currentFn := block.Parent()
	if currentFn == nil {
		return false
	}

	queue := []*ssa.Function{currentFn}
	visited := make(map[*ssa.Function]bool)
	visited[currentFn] = true

	for len(queue) > 0 {
		fn := queue[0]
		queue = queue[1:]

		refs := fn.Referrers()
		if refs == nil {
			continue
		}

		for _, ref := range *refs {
			if call, ok := ref.(*ssa.Call); ok && isCleanupCall(call, fn) {
				return true
			}

			if closure, ok := ref.(*ssa.MakeClosure); ok {
				closureRefs := closure.Referrers()
				if closureRefs == nil {
					continue
				}

				for _, closureRef := range *closureRefs {
					if call, ok := closureRef.(*ssa.Call); ok && isCleanupCall(call, closure) {
						return true
					}

					if b := closureRef.Block(); b != nil {
						callerFn := b.Parent()
						if callerFn != nil && !visited[callerFn] {
							visited[callerFn] = true
							queue = append(queue, callerFn)
						}
					}
				}
				continue
			}

			if b := ref.Block(); b != nil {
				callerFn := b.Parent()
				if callerFn != nil && !visited[callerFn] {
					visited[callerFn] = true
					queue = append(queue, callerFn)
				}
			}
		}
	}

	return false
}

func isCleanupCall(call *ssa.Call, passedValue ssa.Value) bool {
	common := call.Common()
	if common == nil {
		return false
	}
	callee := common.Value
	if callee == nil {
		return false
	}

	var isCleanup bool
	if fn, ok := callee.(*ssa.Function); ok && fn.Object() != nil {
		obj := fn.Object()
		if obj.Name() == testingCleanupMethod && obj.Pkg() != nil && obj.Pkg().Path() == testingPackagePath {
			isCleanup = true
		}
	}
	if !isCleanup && common.Method != nil && common.Method.Name() == testingCleanupMethod {
		if named, ok := common.Value.Type().(*types.Named); ok {
			if pkg := named.Obj().Pkg(); pkg != nil && pkg.Path() == testingPackagePath {
				isCleanup = true
			}
		}
	}

	if !isCleanup {
		return false
	}

	for _, arg := range common.Args {
		if arg == passedValue {
			return true
		}
	}
	return false
}
