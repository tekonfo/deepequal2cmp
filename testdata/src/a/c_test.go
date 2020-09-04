package a

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type S2 struct {
	F1      int
	private int
}

func Test_hoge(t *testing.T) {
	v1 := S2{F1: 1, private: 1}
	v2 := S2{F1: 1, private: 1}
	v3 := S2{F1: 1, private: 2}

	if diff := cmp.Diff(v1, v2); diff != "" {
		fmt.Printf("v1 != v2\n%s\n", diff)
	} else {
		fmt.Println("v1 == v2")
	}

	if diff := cmp.Diff(v1, v3); diff != "" {
		fmt.Printf("v1 != v3\n%s\n", diff)
	} else {
		fmt.Println("v1 == v3")
	}
}
