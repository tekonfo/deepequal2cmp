package deepequal2cmp

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const doc = "deepequal2cmp is ..."

// Analyzer is ...
var Analyzer = &analysis.Analyzer{
	Name: "deepequal2cmp",
	Doc:  doc,
	Run:  run,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		switch n := n.(type) {
		case *ast.CallExpr:
			f := n.Fun.(*ast.SelectorExpr)
			if f.Sel.Name == "DeepEqual" {
				pass.Reportf(n.Pos(), "DeepEqual is used")
			}
		}
	})

	return nil, nil
}
