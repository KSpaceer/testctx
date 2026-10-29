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

// Analyzer is the analysis.Analyzer for testctx
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
	ssaResult := pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA)
	cleanupFunctions := functionsCalledDuringCleanup(ssaResult)

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
		if instr := callToSSA(ssaResult, callExpr); instr != nil && cleanupFunctions[instr.Block().Parent()] {
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

// functionsCalledDuringCleanup follows calls forward from registered callbacks.
// A function is exempt if any cleanup path reaches it, even if it also has
// ordinary callers. Merely creating a nested closure does not call it.
func functionsCalledDuringCleanup(result *buildssa.SSA) map[*ssa.Function]bool {
	bindings := closureBindings(result)
	callees := make(map[*ssa.Function][]*ssa.Function)
	var queue []*ssa.Function
	for _, fn := range result.SrcFuncs {
		for _, block := range fn.Blocks {
			for _, instr := range block.Instrs {
				call, ok := instr.(ssa.CallInstruction)
				if !ok {
					continue
				}
				common := call.Common()
				callees[fn] = append(callees[fn], functionTargets(common.Value, bindings)...)
				if isCleanupCall(common) {
					callback := common.Args[len(common.Args)-1]
					queue = append(queue, functionTargets(callback, bindings)...)
				}
			}
		}
	}

	reachable := make(map[*ssa.Function]bool)
	for len(queue) > 0 {
		fn := queue[0]
		queue = queue[1:]
		if reachable[fn] {
			continue
		}
		reachable[fn] = true
		queue = append(queue, callees[fn]...)
	}
	return reachable
}

// closureBindings connects captured variables to their bindings and local
// storage to its possible values. Stores are combined without regard to order.
func closureBindings(result *buildssa.SSA) map[ssa.Value][]ssa.Value {
	bindings := make(map[ssa.Value][]ssa.Value)
	for _, fn := range result.SrcFuncs {
		for _, block := range fn.Blocks {
			for _, instr := range block.Instrs {
				switch instr := instr.(type) {
				case *ssa.Store:
					bindings[instr.Addr] = append(bindings[instr.Addr], instr.Val)
				case *ssa.MakeClosure:
					closure := instr.Fn.(*ssa.Function)
					for i, binding := range instr.Bindings {
						variable := closure.FreeVars[i]
						bindings[variable] = append(bindings[variable], binding)
					}
				}
			}
		}
	}
	return bindings
}

// functionTargets resolves local aliases, captured variables, and branch-selected
// function values. Parameters and returned functions are unresolved.
func functionTargets(value ssa.Value, bindings map[ssa.Value][]ssa.Value) []*ssa.Function {
	var functions []*ssa.Function
	queue := []ssa.Value{value}
	visited := make(map[ssa.Value]bool)
	for len(queue) > 0 {
		value := queue[0]
		queue = queue[1:]
		if visited[value] {
			continue
		}
		visited[value] = true
		queue = append(queue, bindings[value]...)
		switch value := value.(type) {
		case *ssa.Function:
			functions = append(functions, value)
		case *ssa.MakeClosure:
			queue = append(queue, value.Fn)
		case *ssa.Phi:
			queue = append(queue, value.Edges...)
		case *ssa.ChangeType:
			queue = append(queue, value.X)
		case *ssa.UnOp:
			if value.Op == token.MUL {
				queue = append(queue, value.X)
			}
		}
	}
	return functions
}

func isCleanupCall(common *ssa.CallCommon) bool {
	method := common.Method
	if fn := common.StaticCallee(); fn != nil {
		method, _ = fn.Object().(*types.Func)
	}
	return method != nil && method.Name() == testingCleanupMethod &&
		method.Pkg() != nil && method.Pkg().Path() == testingPackagePath
}
