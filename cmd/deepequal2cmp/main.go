package main

import (
	"flag"
	"fmt"

	"github.com/tekonfo/deepequal2cmp"
)

func main() {
	flag.Parse()

	args := flag.Args()

	var dir string

	if len(args) == 0 {
		// 指定がない場合はカレントディレクトリを参照する
		dir = "./"
	} else if len(args) == 1 {
		dir = args[0]
	} else {
		fmt.Println("please fill in only one dir path!")
		return
	}

	deepequal2cmp.Rewrite(dir)
}
