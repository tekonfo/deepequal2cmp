package deepequal2cmp

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/astutil"
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

func detectDeepEqual(n *ast.IfStmt) bool {
	unaryExpr, ok := n.Cond.(*ast.UnaryExpr)
	if !ok {
		return false
	}
	callExpr, ok := unaryExpr.X.(*ast.CallExpr)
	if !ok {
		return false
	}
	selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	if selectorExpr.Sel.Name == "DeepEqual" {
		return true
	}
	return false
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.IfStmt)(nil),
	}

	// if got := f(); !reflect.DeepEqual(m1, m2) { ← これを
	// if got := f(); true { ← これに書き換える
	inspect.Preorder(nodeFilter, func(n ast.Node) {
		switch n := n.(type) {
		case *ast.IfStmt:
			// DeepEqualの検知
			isUsedDeepEqual := detectDeepEqual(n)
			if !isUsedDeepEqual {
				return
			}

			// // DeepEqualをtrue
			// unaryExpr, ok := n.Cond.(*ast.UnaryExpr)
			// if !ok {
			// 	return
			// }

			fset := token.NewFileSet()
			var buf1 bytes.Buffer
			err := format.Node(&buf1, fset, n)
			if err != nil {
				panic(err)
			}
			fmt.Println(buf1.String())

			d := astutil.Apply(n, func(cr *astutil.Cursor) bool {
				switch cr.Name() {
				case "Cond":
					cr.Replace(&ast.UnaryExpr{Op: token.NOT, X: ast.NewIdent("true")})
				}
				return true
			}, nil)

			fset = token.NewFileSet()
			var buf2 bytes.Buffer
			err = format.Node(&buf2, fset, d)
			if err != nil {
				panic(err)
			}
			fmt.Println(buf2.String())
		}
	})

	return nil, nil
}
