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
		(*ast.IfStmt)(nil),
	}

	// if got := f(); !reflect.DeepEqual(m1, m2) { ← これを検知したい
	inspect.Preorder(nodeFilter, func(n ast.Node) {
		switch n := n.(type) {
		case *ast.IfStmt:
			unaryExpr, ok := n.Cond.(*ast.UnaryExpr)
			if !ok {
				return
			}
			callExpr, ok := unaryExpr.X.(*ast.CallExpr)
			if !ok {
				return
			}
			selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
			if !ok {
				return
			}
			if selectorExpr.Sel.Name == "DeepEqual" {
				pass.Reportf(n.Pos(), "DeepEqual is used")
			}
		}
	})

	return nil, nil
}
