package deepequal2cmp

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"

	"golang.org/x/tools/go/ast/astutil"
)

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

func detectCmpDiff(n *ast.IfStmt) bool {
	assignStmt, ok := n.Init.(*ast.AssignStmt)
	if !ok {
		return false
	}
	callExpr, ok := assignStmt.Rhs[0].(*ast.CallExpr)
	if !ok {
		return false
	}
	selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	if selectorExpr.X.(*ast.Ident).Name == "cmp" && selectorExpr.Sel.Name == "Diff" {
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
func initNode(firstArg ast.Node, secondArg ast.Node) ast.Stmt {
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
func condNode() ast.Expr {
	return &ast.BinaryExpr{
		Op: token.NEQ,
		X:  ast.NewIdent("diff"),
		Y:  &ast.BasicLit{Kind: token.STRING, Value: "\"\""},
	}
}

// fmt.Printf("f() differs: (-got +want)\n%s", diff)を作成
func bodyNode() *ast.BlockStmt {
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

func execFuncNode(node *ast.IfStmt) (ast.Node, error) {
	assignStmt, ok := node.Init.(*ast.AssignStmt)
	if !ok {
		return nil, errors.New("cast error")
	}

	return assignStmt, nil
}

// !reflect.DeepEqual(m1, m2) で利用されているm1, m2を取得する
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
		n := cr.Node()
		ifStmt, ok := n.(*ast.IfStmt)
		if !ok {
			return true
		}

		isUsedDeepEqual := detectDeepEqual(ifStmt)
		if !isUsedDeepEqual {
			return true
		}

		arg1, arg2, err := getArg(ifStmt)
		if err != nil {
			panic(err)
		}

		newIfStmt := &ast.IfStmt{
			Init: initNode(arg1, arg2),
			Cond: condNode(),
			Body: bodyNode(),
		}

		// FIXME: この処理を加えても、コードが書き換わらない
		cr.Replace(newIfStmt)

		return true
	}, nil)

	return nil
}

// Rewrite is 外部から呼び出され、DeepEqualをcmp.Diffに書き換える
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
