package a

import (
	"reflect"
	"testing"
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
			if got := f(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("f() = %v, want %v", got, tt.want)
			}
		})
	}
}
