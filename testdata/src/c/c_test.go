package c

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_a(t *testing.T) {
	tests := []struct {
		name string
		want hoge
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := a()
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("a() differs: (-got +want)\n%s", diff)
			}

		})
	}
}
