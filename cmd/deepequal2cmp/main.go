package main

import (
	"log"
	"os"

	"github.com/tekonfo/deepequal2cmp"
	"github.com/urfave/cli/v2"
)

func main() {

	deepequal2cmp.ParcePackage("testdata/src/c/c_test.go")

	return

	app := &cli.App{
		Name:  "deepequal2cmp",
		Usage: "convert DeepEqual to cmp.Diff to target test files",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "dir", Aliases: []string{"d"}},
		},
		Action: func(c *cli.Context) error {
			var dir string
			if c.String("dir") == "" {
				dir = "./"
			} else {
				dir = c.String("dir")
			}

			deepequal2cmp.Rewrite(dir)

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
