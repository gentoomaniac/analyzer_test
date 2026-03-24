package main

import (
	"bytes"
	"go/ast"
	"go/printer"
	"go/token"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "printffinder",
	Doc:      "checks for pure string concatenation in Sprintf",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {

	ins := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// nodeFilter := []ast.Node{
	// 	(*ast.BinaryExpr)(nil),
	// 	(*ast.FuncDecl)(nil),
	// 	(*ast.CallExpr)(nil),
	// 	(*ast.SelectorExpr)(nil),
	// }

	//ins.Preorder(nodeFilter, checkSprintfConcat(pass))

	for call := range ins.PreorderSeq((*ast.CallExpr)(nil)) {
		checkSprintf(pass, call)
	}
	return nil, nil
}

func checkSprintfConcat(pass *analysis.Pass) func(n ast.Node) {
	return func(n ast.Node) {
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			return
		}

		ast.Inspect(fn.Body, func(node ast.Node) bool {
			callExpr, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			ident := selExpr.Sel
			info := pass.TypesInfo.Uses[ident]
			if info == nil {
				return true
			}

			pkg := info.Pkg()
			if pkg == nil {
				return true
			}

			if pkg.Path() != "fmt" || info.Name() != "Sprintf" {
				return true
			}

			formatString := callExpr.Args[0]
			lit, ok := formatString.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}

			strValue, err := strconv.Unquote(lit.Value)
			if err != nil {
				return true
			}

			sCount := strings.Count(strValue, "%s")
			if sCount == 0 || len(strValue) != sCount*2 {
				return true
			}

			if len(callExpr.Args) < sCount+1 {
				return true
			}

			var args []string
			for _, node := range callExpr.Args[1 : sCount+1] {
				var buf bytes.Buffer
				err := printer.Fprint(&buf, pass.Fset, node)
				if err != nil {
					return true
				}
				args = append(args, buf.String())
			}
			newText := strings.Join(args, " + ")

			pass.Report(analysis.Diagnostic{
				Pos:     ident.Pos(),
				Message: "unnecessary use of fmt.Sprintf",
				SuggestedFixes: []analysis.SuggestedFix{
					{
						Message: "replace with + operator string concatenation",
						TextEdits: []analysis.TextEdit{
							{
								Pos:     ident.Pos(),
								End:     ident.End(),
								NewText: []byte(newText),
							},
						},
					},
				},
			})

			return true
		})
	}
}

func checkSprintf(pass *analysis.Pass, n ast.Node) {
	fn, ok := n.(*ast.FuncDecl)
	if !ok || fn.Body == nil {
		return
	}

	ast.Inspect(fn.Body, func(node ast.Node) bool {
		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		ident := selExpr.Sel
		info := pass.TypesInfo.Uses[ident]
		if info == nil {
			return true
		}

		pkg := info.Pkg()
		if pkg == nil {
			return true
		}

		if pkg.Path() != "fmt" || info.Name() != "Sprintf" {
			return true
		}

		formatString := callExpr.Args[0]
		lit, ok := formatString.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}

		strValue, err := strconv.Unquote(lit.Value)
		if err != nil {
			return true
		}

		sCount := strings.Count(strValue, "%s")
		if sCount == 0 || len(strValue) != sCount*2 {
			return true
		}

		if len(callExpr.Args) < sCount+1 {
			return true
		}

		var args []string
		for _, node := range callExpr.Args[1 : sCount+1] {
			var buf bytes.Buffer
			err := printer.Fprint(&buf, pass.Fset, node)
			if err != nil {
				return true
			}
			args = append(args, buf.String())
		}
		newText := strings.Join(args, " + ")

		pass.Report(analysis.Diagnostic{
			Pos:     ident.Pos(),
			Message: "unnecessary use of fmt.Sprintf",
			SuggestedFixes: []analysis.SuggestedFix{
				{
					Message: "replace with + operator string concatenation",
					TextEdits: []analysis.TextEdit{
						{
							Pos:     ident.Pos(),
							End:     ident.End(),
							NewText: []byte(newText),
						},
					},
				},
			},
		})

		return true
	})
}
