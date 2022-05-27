package commands

import (
	"github.com/urfave/cli/v2"
)

func Img2Src() *cli.Command {
	return &cli.Command{
		Name:   "img2src",
		Usage:  "Converts an image to WASM-4 source code",
		Action: imgs2src,
	}
}

func imgs2src(c *cli.Context) error {
	return nil
}
