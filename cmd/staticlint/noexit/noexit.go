// Package noexit provides an analysis pass that reports
// calls to os.Exit within main packages, encouraging
// explicit error reporting instead of abrupt termination.
package noexit

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var NoExitCodeAnalyzer = &analysis.Analyzer{
	Name: "noexitcode",
	Doc:  "reports calls to os.Exit() in main packages",
	Run:  runNoExitCode,
}

func runNoExitCode(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		pkgName := pass.Pkg.Name()
		if pkgName != "main" {
			continue
		}
		ast.Inspect(file, func(n ast.Node) bool {
			callExpr, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			pkgIdent, ok := selExpr.X.(*ast.Ident)
			if !ok {
				return true
			}
			if pkgIdent.Name == "os" && selExpr.Sel.Name == "Exit" {
				pass.Reportf(callExpr.Lparen, "do not call os.Exit in main packages; use error reporting instead")
			}

			return true
		})
	}
	return nil, nil
}
