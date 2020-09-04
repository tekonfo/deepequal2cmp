package a

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_hoge(t *testing.T) {
	tests := []struct {
		name string
		want interface{}
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hoge()
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("hoge() differs: (-got +want)\n%s", diff)
			}

		})
	}
}
