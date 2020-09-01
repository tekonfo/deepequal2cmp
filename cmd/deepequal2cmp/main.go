package main

import (
	"deepequal2cmp"
	"golang.org/x/tools/go/analysis/unitchecker"
)

func main() { unitchecker.Main(deepequal2cmp.Analyzer) }

