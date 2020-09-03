package deepequal2cmp

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
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
						X:   ast.NewIdent("t"),
						Sel: ast.NewIdent("Errorf"),
					},
					Args: []ast.Expr{
						// TODO: ここのf()は動的な値なのであとで書き換える必要がある
						&ast.BasicLit{Kind: token.STRING, Value: "\"differs: (-got +want)\\n%s\""},
						ast.NewIdent("diff"),
					},
				},
			},
		},
	}
}

// got = f(); はifStmtの前に記述しなければならないので、その箇所を取得
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
func deepEqual2cmp(f *ast.File) error {
	astutil.Apply(f, func(cr *astutil.Cursor) bool {
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

		s, err := execFuncNode(ifStmt)
		// TODO: error処理
		if err != nil {

		}

		newIfStmt := &ast.IfStmt{
			Init: initNode(arg1, arg2),
			Cond: condNode(),
			Body: bodyNode(),
		}

		cr.Replace(newIfStmt)
		if s != nil {
			cr.InsertBefore(s)
		}

		return true
	}, nil)

	return nil
}

func makeFile(n interface{}, fileName string) error {
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}

	defer func() {
		f.Close()
	}()

	fset := token.NewFileSet()
	err = format.Node(f, fset, n)
	if err != nil {
		return err
	}

	return nil
}

func findTestFiles(dir string) []string {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	var paths []string
	for _, file := range files {
		if file.IsDir() {
			paths = append(paths, findTestFiles(filepath.Join(dir, file.Name()))...)
			continue
		}

		// *_test.goのみappendする
		if !strings.HasSuffix(file.Name(), "_test.go") {
			continue
		}

		paths = append(paths, filepath.Join(dir, file.Name()))
	}

	return paths
}

// Rewrite is 外部から呼び出され、DeepEqualをcmp.Diffに書き換える
func Rewrite(dirPath string) {

	// ディレクトリ以下の全testファイルを取得する
	files := findTestFiles(dirPath)

	mode := packages.NeedSyntax // 構文解析まで
	cfg := &packages.Config{Mode: mode, Tests: true}
	pkgs, err := packages.Load(cfg, files...)
	if err != nil { /* エラー処理 */
		fmt.Println("error")
		panic(err)
	}
	if packages.PrintErrors(pkgs) > 0 { /* エラー処理 */
		fmt.Println("error error")
		panic("error")
	}

	for _, pkg := range pkgs {
		for _, f := range pkg.Syntax {
			// mainパッケージは除外
			if f.Name.Name == "main" {
				continue
			}

			deepEqual2cmp(f)

			showBuf(f)
		}
	}

	// for _, _ = range fileNames {

	// }
	// fs := token.NewFileSet()
	// f, err := parser.ParseFile(fs, "testdata/src/a/a_test.go", nil, 0)
	// if err != nil {
	// 	panic(err)
	// }

	// deepEqual2cmp(f)

	// showBuf(f)

	// makeFile(f, "testdata/src/a/a_test_test.go")

	// // goimportsをapplyしてreflectとgo-cmpをimport処理する
	// // TODO: 自分でもできる。。。
	// err = exec.Command("goimports", "-w", "-l", "testdata/src/a/a_test_test.go").Run()
	// if err != nil {
	// 	panic(err)
	// }
}
