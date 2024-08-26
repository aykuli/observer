// Package noosexit defines an Analyzer that checks for not using os.Exit
package noosexit

import (
	"fmt"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

const Doc = `check for not using os.Exit call`

var Analyzer = &analysis.Analyzer{
	Name:     "sortslice",
	Doc:      Doc,
	URL:      "https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/sortslice",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	//inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	for id, obj := range pass.TypesInfo.Defs {
		fmt.Println(id)
		fmt.Println(obj)
	}
	fmt.Printf("pass: %v\n\n", pass)
	return nil, nil
}
