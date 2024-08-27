package noosexit

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

const Doc = `check for not using os.Exit call`

var Analyzer = &analysis.Analyzer{
	Name: "noosexit0",
	Doc:  Doc,
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		i := 0
		ast.Inspect(file, func(node ast.Node) bool {
			switch x := node.(type) {
			case *ast.Ident:
				if x.Name == "os" {
					i++
				}
				if x.Name == "Exit" && i == 1 {
					pass.Reportf(x.Pos(), "os.Exit not allowed")
				}
			}
			return true
		})
	}
	return nil, nil
}
