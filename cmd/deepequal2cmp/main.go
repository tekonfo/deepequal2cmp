package main

import (
	"deepequal2cmp"
	"flag"
)

func main() {
	flag.Parse()
	deepequal2cmp.Rewrite(flag.Args())
}
