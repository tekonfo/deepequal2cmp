package deepequal2cmp

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/tools/go/packages"
)

// test用: fileを書き換え、書き換えた後のファイルパスを返却
func convertFile(file string) string {
	fs := token.NewFileSet()
	mode := parser.ParseComments
	f, err := parser.ParseFile(fs, file, nil, mode)
	if err != nil {
		panic(err)
	}

	pkg := ParcePackage(file)

	_, _ = deepEqual2cmp(f, pkg)

	file = file[:len(file)-3] + "_test.go"

	makeFile(f, fs, file)

	err = exec.Command("goimports", "-w", "-l", file).Run()
	if err != nil {
		panic(err)
	}

	return file
}

func Test_convertFile(t *testing.T) {
	type args struct {
		file string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "ok",
			args: args{
				file: "testdata/src/a/a_test.go",
			},
			want: "testdata/src/b/a_test.go",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			bytes, err := ioutil.ReadFile(tt.want)
			if err != nil {
				t.Errorf("cannot read want file")
			}
			want := string(bytes)

			gotFile := convertFile(tt.args.file)
			defer func() {
				os.Remove(gotFile)
			}()
			bytes, err = ioutil.ReadFile(gotFile)
			if err != nil {
				t.Errorf("cannot read want file")
			}
			got := string(bytes)

			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("want != got\n%s\n", diff)
			}
		})
	}
}

type S2 struct {
	F1  int
	aaa int
}

func Test_hoge(t *testing.T) {
	v1 := S2{F1: 1, aaa: 1}
	v2 := S2{F1: 1, aaa: 1}

	opt := cmp.AllowUnexported(v1, v2)

	if diff := cmp.Diff(v1, v2, opt); diff != "" {
		fmt.Printf("v1 != v2\n%s\n", diff)
	} else {
		fmt.Println("v1 == v2")
	}
}

func Test_parcePackage(t *testing.T) {
	type args struct {
		packageName string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "OK",
			args: args{
				packageName: "deepequal2cmp/testdata/src/c",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ParcePackage(tt.args.packageName)
		})
	}
}

func Test_findTypesStruct(t *testing.T) {
	type args struct {
		pkg   *packages.Package
		ident *ast.Ident
	}
	tests := []struct {
		name string
		args args
		want *types.Struct
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := findTypesStruct(tt.args.pkg, tt.args.ident); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("findTypesStruct() = %v, want %v", got, tt.want)
			}
		})
	}
}
