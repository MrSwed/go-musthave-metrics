package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

const mainFuncBadName = "os.Exit"
const errOSExitMainCheck = "os.Exit is not allowed at main func of main package"

// OsExitCheckAnalyzer
// Deny use os.Exit expression
var OsExitCheckAnalyzer = &analysis.Analyzer{
	Name: "osexitcheck",
	Doc:  errOSExitMainCheck,
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	exprIsDeny := func(call *ast.CallExpr) bool {
		chName := ""
		if f, ok := call.Fun.(*ast.SelectorExpr); ok {
			if pkg, ok := f.X.(*ast.Ident); ok {
				chName = pkg.Name + "." + f.Sel.Name
			}
		}
		return chName == mainFuncBadName
	}
	for _, file := range pass.Files {
		if file.Name.String() != "main" {
			continue
		}
		ast.Inspect(file, func(node ast.Node) bool {
			switch x := node.(type) {
			case *ast.FuncDecl:
				if x.Name.String() != "main" {
					return false
				}
			case *ast.ExprStmt:
				if call, ok := x.X.(*ast.CallExpr); ok && exprIsDeny(call) {
					pass.Reportf(x.Pos(), errOSExitMainCheck)
				}
			case *ast.GoStmt:
				if exprIsDeny(x.Call) {
					pass.Reportf(x.Pos(), errOSExitMainCheck)
				}
			case *ast.DeferStmt:
				if exprIsDeny(x.Call) {
					pass.Reportf(x.Pos(), errOSExitMainCheck)
				}
			}
			return true
		})
	}
	return nil, nil
}
