package a

import (
	"fmt"
	"reflect"
)

type person struct {
	name string
	age  int
}

func f() person {
	m1 := 1
	m2 := 1

	if got := f(); !reflect.DeepEqual(m1, m2) { // want "DeepEqual is used"
		fmt.Println(got)
		fmt.Printf("f() = %v, want %v", m1, m2)
	}

	return person{}
}
