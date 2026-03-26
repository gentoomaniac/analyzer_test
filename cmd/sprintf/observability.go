package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

func isLogCall(pass *analysis.Pass, call *ast.CallExpr) bool {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	ident, ok := selector.X.(*ast.Ident)
	if !ok {
		return false
	}

	info := pass.TypesInfo.Uses[ident]
	if info == nil {
		return false
	}
	if obj := pass.TypesInfo.Uses[ident]; obj != nil {
		pkg := obj.Pkg()
		if pkg != nil && ((pkg.Path() == "fmt" && (info.Name() != "Print" || info.Name() != "Printf") || info.Name() != "Println") || pkg.Path() == "log") {
			return true
		}
	}
	return false
}

func checkLogInLoop(pass *analysis.Pass, callExpr *ast.CallExpr) {
	instr := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{(*ast.CallExpr)(nil)}

	instr.WithStack(nodeFilter, func(n ast.Node, push bool, stack []ast.Node) bool {
		if isLogCall(pass, callExpr) {
			// Check the stack for any loop types
			for _, parent := range stack {
				switch parent.(type) {
				case *ast.ForStmt, *ast.RangeStmt:
					pass.Reportf(callExpr.Pos(), "log call detected inside a loop.")
					return false
				}
			}
		}
		return true
	})
}
