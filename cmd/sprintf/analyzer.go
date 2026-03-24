package main

import (
	"bytes"
	"go/ast"
	"go/printer"
	"go/token"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "printffinder",
	Doc:  "checks for pure string concatenation in Sprintf",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, astFile := range pass.Files {
		ast.Inspect(astFile, func(n ast.Node) bool {
			if n == nil {
				return true
			}

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
			// Guard: Ensure there is at least one "%s" AND the string contains nothing else
			if sCount == 0 || len(strValue) != sCount*2 {
				return true
			}

			// Not enough parameters for all the %s another analyzer should fetch this
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

			// If we survived the gauntlet of guards above, we have our match!
			pass.Report(analysis.Diagnostic{
				Pos:     ident.Pos(),
				Message: "unnecessary use of fmt.Sprintf",
				SuggestedFixes: []analysis.SuggestedFix{
					{
						Message: "replace with string concatenation",
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
	return nil, nil
}
