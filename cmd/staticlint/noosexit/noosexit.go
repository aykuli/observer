// Package noosexit defines an Analyzer that checks if main goroutine
// in main function doesn't call os.Exit
package noosexit

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

const Doc = `check for not using os.Exit call`

var Analyzer = &analysis.Analyzer{
	Name: "noosexit",
	Doc:  Doc,
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if file.Name.Name != "main" {
			return nil, nil
		}
		warnCount := 0
		for _, decl := range file.Decls {
			ast.Inspect(decl, func(node ast.Node) bool {
				switch y := decl.(type) {
				case *ast.FuncDecl:
					if y.Name.Name != "main" {
						break
					}

					for _, body := range y.Body.List {
						switch z := body.(type) {
						case *ast.ExprStmt:
							switch a := z.X.(type) {
							case *ast.CallExpr:
								switch fun := a.Fun.(type) {
								case *ast.SelectorExpr:
									x := fun.X.(*ast.Ident)
									if fun.Sel.Name == "Exit" && x.Name == "os" && warnCount == 0 {
										warnCount++
										pass.Reportf(x.Pos(), "os.Exit doesnt allowed in main function of main package")
									}
								}
							}
						}

					}

					return true
				}
				return true
			})
		}
	}

	return nil, nil
}
