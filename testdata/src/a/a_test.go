package a

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_f(t *testing.T) {
	tests := []struct {
		name string
		want person
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := f()
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("f() differs: (-got +want)\n%s", diff)
			}

		})
	}
}

func Test_ff(t *testing.T) {
	tests := []struct {
		name    string
		want    person
		want1   person
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := ff()
			if (err != nil) != tt.wantErr {
				t.Errorf("ff() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("ff() differs: (-got +want)\n%s", diff)
			}
			if diff := cmp.Diff(got1, tt.want1); diff != "" {
				t.Errorf("ff() differs: (-got +want)\n%s", diff)
			}

		})
	}
}
