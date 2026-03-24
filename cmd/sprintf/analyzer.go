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

	for call := range ins.PreorderSeq((*ast.CallExpr)(nil)) {
		callExpr, ok := call.(*ast.CallExpr)
		if !ok {
			continue
		}
		checkSprintf(pass, callExpr)
	}
	return nil, nil
}

func checkSprintf(pass *analysis.Pass, callExpr *ast.CallExpr) {
	selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}

	ident := selExpr.Sel
	info := pass.TypesInfo.Uses[ident]
	if info == nil {
		return
	}

	pkg := info.Pkg()
	if pkg == nil {
		return
	}

	if pkg.Path() != "fmt" || info.Name() != "Sprintf" {
		return
	}

	formatString := callExpr.Args[0]
	lit, ok := formatString.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return
	}

	strValue, err := strconv.Unquote(lit.Value)
	if err != nil {
		return
	}

	sCount := strings.Count(strValue, "%s")
	if sCount == 0 || len(strValue) != sCount*2 {
		return
	}

	if len(callExpr.Args) < sCount+1 {
		return
	}

	var args []string
	for _, node := range callExpr.Args[1 : sCount+1] {
		var buf bytes.Buffer
		err := printer.Fprint(&buf, pass.Fset, node)
		if err != nil {
			return
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
}
