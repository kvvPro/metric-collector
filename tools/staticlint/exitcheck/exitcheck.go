// Module exitcheck set check for calling os.Exit()
// via inspecting full AST
package exitcheck

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  "check for os.Exit() in main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	check := func(node ast.Node) {
		if expr, ok := node.(*ast.ExprStmt); ok {
			if c, ok := expr.X.(*ast.CallExpr); ok {
				if s, ok := c.Fun.(*ast.SelectorExpr); ok {
					if i, ok := s.X.(*ast.Ident); ok {
						// только для вызова функции os.Exit()
						if i.Name == "os" && s.Sel.Name == "Exit" {
							pass.Reportf(s.Pos(), "not allowed using of os.Exit()")
						}
					}
				}
			}
		}
	}

	// реализация будет ниже
	for _, file := range pass.Files {
		if file.Name.Name == "main" {
			// функцией ast.Inspect проходим по всем узлам AST
			ast.Inspect(file, func(node ast.Node) bool {
				if m, ok := node.(*ast.FuncDecl); ok {
					if m.Name.Name == "main" {
						// интересуют только вызовы функций
						for _, st := range m.Body.List {
							check(st)
						}
					}
				}
				return true
			})
		}
	}
	return nil, nil
}
