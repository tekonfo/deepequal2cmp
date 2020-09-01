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
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ff() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("ff() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
