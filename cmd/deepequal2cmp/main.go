package main

import (
	"errors"
	"log"
	"os"

	"github.com/tekonfo/deepequal2cmp"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "deepequal2cmp",
		Usage: "convert DeepEqual to cmp.Diff",
		Action: func(c *cli.Context) error {
			var dir string
			if c.NArg() == 0 {
				dir = "./"
			} else if c.NArg() == 1 {
				dir = c.Args().Get(0)
			} else {
				return errors.New("please 1 args! ")
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
