package deepequal2cmp

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
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
func initNode(firstArg ast.Node, secondArg ast.Node) ast.Node {
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
					firstArg.(ast.Expr),
					secondArg.(ast.Expr),
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

func getArg(node *ast.IfStmt) (ast.Node, ast.Node, error) {
	unaryExpr, ok := node.Cond.(*ast.UnaryExpr)
	if !ok {
		return nil, nil, errors.New("fail unaryExpr cast")
	}
	callExpr, ok := unaryExpr.X.(*ast.CallExpr)
	if !ok {
		return nil, nil, errors.New("fail callExpr cast")
	}
	args := callExpr.Args
	if len(args) != 2 {
		return nil, nil, errors.New("invalid args number")
	}

	return args[0], args[1], nil
}

// if _ = f(); !reflect.DeepEqual(m1, m2) { ← これを
// 	fmt.Printf("f() = %v, want %v", m1, m2)
// }

// if diff := cmp.Diff(m1, m2); diff != "" { ← これにする
// 	fmt.Printf("f() differs: (-got +want)\n%s", diff)
// }
func deepEqual2cmp(n ast.Node) error {
	astutil.Apply(n, func(cr *astutil.Cursor) bool {
		// ifstmtかどうかを確認
		pNode := cr.Parent()
		pIfStmt, ok := pNode.(*ast.IfStmt)
		if !ok {
			return true
		}

		isUsedDeepEqual := detectDeepEqual(pIfStmt)
		if !isUsedDeepEqual {
			return false
		}

		arg1, arg2, err := getArg(pIfStmt)
		if err != nil {
			panic(err)
		}

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

	return nil
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

			_ = deepEqual2cmp(n)
		}
	})

	return nil, nil
}

func Rewrite() {
	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, "testdata/src/a/a_test.go", nil, 0)
	if err != nil {
		panic(err)
	}

	ast.Inspect(f, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.IfStmt:
			// DeepEqualの検知
			isUsedDeepEqual := detectDeepEqual(n)
			if !isUsedDeepEqual {
				return true
			}
			deepEqual2cmp(n)
		}
		// falseを返すと子ノードの探索をしない
		return true
	})

	showBuf(f)

}
