package deepequal2cmp

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

// Rewrite is 外部から呼び出され、DeepEqualをcmp.Diffに書き換える
func Rewrite(dirPath string) {

	// ディレクトリ以下の全testファイルを取得する
	files := findTestFiles(dirPath)

	fmt.Println("changed files:")

	for _, file := range files {
		fs := token.NewFileSet()
		mode := parser.ParseComments
		f, err := parser.ParseFile(fs, file, nil, mode)
		if err != nil {
			panic(err)
		}

		pkg := ParcePackage(file)

		isChanged, err := deepEqual2cmp(f, pkg)
		if !isChanged {
			continue
		}

		fmt.Println("\t", file)

		makeFile(f, fs, file)

		// goimportsをapplyしてreflectとgo-cmpをimport処理する
		err = exec.Command("goimports", "-w", "-l", file).Run()
		if err != nil {
			panic(err)
		}
	}
}

func detectDeepEqual(n *ast.IfStmt) bool {
	var isDetect bool

	ast.Inspect(n, func(n ast.Node) bool {
		if n == nil {
			return false
		}

		if selectorExpr, ok := n.(*ast.SelectorExpr); ok {
			if selectorExpr.Sel.Name == "DeepEqual" {
				isDetect = true
				return false
			}
			return true
		}

		return true
	})

	return isDetect
}

func detectErrorf(n *ast.CallExpr) bool {
	selectorExpr, ok := n.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	if selectorExpr.Sel.Name == "Errorf" {
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
func bodyNode(funcName string) *ast.BlockStmt {

	bodyStr := "\"" + funcName + " differs: (-got +want)\\n%s\""

	return &ast.BlockStmt{
		List: []ast.Stmt{
			&ast.ExprStmt{
				X: &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   ast.NewIdent("t"),
						Sel: ast.NewIdent("Errorf"),
					},
					Args: []ast.Expr{
						&ast.BasicLit{Kind: token.STRING, Value: bodyStr},
						ast.NewIdent("diff"),
					},
				},
			},
		},
	}
}

// got = f(); はifStmtの前に記述しなければならないので、その箇所を取得
func execFuncNode(node *ast.IfStmt) ast.Node {
	assignStmt, ok := node.Init.(*ast.AssignStmt)
	if !ok {
		return nil
	}

	return assignStmt
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

// 実行する関数名を取得する
// t.Errorf("ff() error = %v, wantErr %v", err, tt.wantErr)
// この文の中のff()を取得する
func getFuncName(node *ast.IfStmt) (string, error) {

	stmts := node.Body.List

	for _, stmt := range stmts {
		exprStmt, ok := stmt.(*ast.ExprStmt)
		if !ok {
			continue
		}
		callExpr, ok := exprStmt.X.(*ast.CallExpr)
		if !ok {
			continue
		}

		selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
		if !ok {
			continue
		}
		if selectorExpr.Sel.Name != "Errorf" {
			continue
		}

		// get string from arg such as "\"f() = %v, want %v\""
		arg := callExpr.Args[0]
		basicLit, ok := arg.(*ast.BasicLit)
		if !ok {
			continue
		}

		// "\"f() = %v, want %v\""
		argStr := basicLit.Value

		// get "\"f()"
		funcName := strings.Split(argStr, " ")[0]
		// get "f()"
		funcName = strings.Replace(funcName, "\"", "", -1)

		return funcName, nil
	}

	return "", nil
}

// if _ = f(); !reflect.DeepEqual(m1, m2) { ← これを
// 	fmt.Printf("f() = %v, want %v", m1, m2)
// }

// if diff := cmp.Diff(m1, m2); diff != "" { ← これにする
// 	fmt.Printf("f() differs: (-got +want)\n%s", diff)
// }
func deepEqual2cmp(f *ast.File, pkg *packages.Package) (bool, error) {
	var isChanged bool

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

		isChanged = true

		arg1, arg2, err := getArg(ifStmt)
		if err != nil {
			fmt.Printf("getArg error: %s", err)
			return true
		}

		target := arg1.(*ast.Ident)
		fmt.Println(target)
		str := findTypesStruct(pkg, target)
		if str != nil {
			fmt.Println(isStructHasUnExportedField(pkg, str))
		} else {
			fmt.Println(str)
		}

		execF := execFuncNode(ifStmt)

		funcName, err := getFuncName(ifStmt)
		if err != nil {
			return true
		}

		newIfStmt := &ast.IfStmt{
			Init: initNode(arg1, arg2),
			Cond: condNode(),
			Body: bodyNode(funcName),
		}

		cr.Replace(newIfStmt)
		if execF != nil {
			cr.InsertBefore(execF)
		}

		return true
	}, nil)

	return isChanged, nil
}

func makeFile(n interface{}, fset *token.FileSet, fileName string) error {
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}

	defer func() {
		f.Close()
	}()

	err = format.Node(f, fset, n)
	if err != nil {
		return err
	}

	return nil
}

func ParcePackage(packageName string) *packages.Package {

	fmt.Println(packageName)

	mode := packages.LoadTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedFiles
	cfg := &packages.Config{Mode: mode, Tests: true}
	pkgs, err := packages.Load(cfg, packageName)
	fmt.Println(pkgs)
	if err != nil { /* エラー処理 */
	}
	if packages.PrintErrors(pkgs) > 0 { /* エラー処理 */
		panic("err")
	}
	for _, pkg := range pkgs {
		for _, f := range pkg.Syntax {
			ast.Print(pkg.Fset, f)
		}
	}
	return pkgs[0]
}

func findTypesStruct(pkg *packages.Package, ident *ast.Ident) *types.Struct {

	fmt.Println(pkg.TypesInfo.Defs)

	obj, ok := pkg.TypesInfo.Defs[ident]
	if !ok {
		fmt.Println(obj)
		return nil
	}

	if a, ok := obj.Type().Underlying().(*types.Struct); ok {
		return a
	}

	return nil
}

func isStructHasUnExportedField(pkg *packages.Package, str *types.Struct) bool {

	for i := 0; i < str.NumFields(); i++ {
		f := str.Field(i)
		if f.Exported() {
			return true
		}
	}

	return false
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
