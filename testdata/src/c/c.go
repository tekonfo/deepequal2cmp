package c

import "fmt"

type hoge struct {
	a int
	B string
}

func a() hoge {
	h := hoge{}
	fmt.Println(h)
	return h
}
