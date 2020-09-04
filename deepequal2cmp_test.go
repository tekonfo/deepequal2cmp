package deepequal2cmp

import (
	"testing"
)

func Test_showDiff(t *testing.T) {
	type args struct {
		dirPath string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "ok",
			args: args{
				dirPath: "testdata/src/a",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			showDiff(tt.args.dirPath)
		})
	}
}
