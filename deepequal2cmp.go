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

func showBuf(n interface{}) {
	fset := token.NewFileSet()
	var buf1 bytes.Buffer
	err := format.Node(&buf1, fset, n)
	if err != nil {
		panic(err)
	}
	fmt.Println(buf1.String())
}

// if diff := cmp.Diff(m1, m2); diff != "" の中の diff := cmp.Diff(m1, m2)の部分を作成
func initNode(firstArgName string, secondArgName string) ast.Node {
	return &ast.AssignStmt{
		Lhs: []ast.Expr{
			ast.NewIdent("diff"),
		},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   ast.NewIdent("cmp"),
					Sel: ast.NewIdent("Diff"),
				},
				Args: []ast.Expr{
					ast.NewIdent(firstArgName),
					ast.NewIdent(secondArgName),
				},
			},
		},
	}
}

// if diff := cmp.Diff(m1, m2); diff != "" の中の diff != ""の部分を作成
func condNode() ast.Node {
	return &ast.BinaryExpr{
		Op: token.NEQ,
		X:  ast.NewIdent("diff"),
		Y:  &ast.BasicLit{Kind: token.STRING, Value: "\"\""},
	}
}

// fmt.Printf("f() differs: (-got +want)\n%s", diff)を作成
func bodyNode() ast.Node {
	return &ast.BlockStmt{
		List: []ast.Stmt{
			&ast.ExprStmt{
				X: &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   ast.NewIdent("fmt"),
						Sel: ast.NewIdent("Printf"),
					},
					Args: []ast.Expr{
						&ast.BasicLit{Kind: token.STRING, Value: "\"f() differs: (-got +want)\\n%s\""},
						ast.NewIdent("diff"),
					},
				},
			},
		},
	}
}

func getArg(node *ast.IfStmt) (string, string) {
	unaryExpr, ok := node.Cond.(*ast.UnaryExpr)
	if !ok {
		return "a", "b"
	}
	callExpr, ok := unaryExpr.X.(*ast.CallExpr)
	if !ok {
		return "a", "b"
	}
	args := callExpr.Args
	return args[0].(*ast.Ident).Name, args[1].(*ast.Ident).Name
}

// if _ = f(); !reflect.DeepEqual(m1, m2) { ← これを
// 	fmt.Printf("f() = %v, want %v", m1, m2)
// }

// if diff := cmp.Diff(m1, m2); diff != "" { ← これにする
// 	fmt.Printf("f() differs: (-got +want)\n%s", diff)
// }
func deepEqual2cmp(n ast.Node) (ast.Node, error) {
	d := astutil.Apply(n, func(cr *astutil.Cursor) bool {
		// ifstmtかどうかを確認
		pNode := cr.Parent()
		pIfStmt, ok := pNode.(*ast.IfStmt)
		if !ok {
			return true
		}

		arg1, arg2 := getArg(pIfStmt)

		switch cr.Name() {
		case "Init":
			cr.Replace(
				initNode(arg1, arg2),
			)
		case "Cond":
			cr.Replace(
				condNode(),
			)
		case "Body":
			cr.Replace(
				bodyNode(),
			)
		}
		return true
	}, nil)

	return d, nil
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.IfStmt)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		switch n := n.(type) {
		case *ast.IfStmt:
			// DeepEqualの検知
			isUsedDeepEqual := detectDeepEqual(n)
			if !isUsedDeepEqual {
				return
			}

			showBuf(n)

			d, _ := deepEqual2cmp(n)

			showBuf(d)
		}
	})

	return nil, nil
}
