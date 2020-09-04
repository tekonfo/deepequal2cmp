package deepequal2cmp

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

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
