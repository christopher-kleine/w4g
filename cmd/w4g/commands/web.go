package commands

import (
	"github.com/urfave/cli/v2"
)

func Web() *cli.Command {
	return &cli.Command{
		Name:   "web",
		Usage:  "Starts a WASM-4 cart in the browser",
		Action: web,
	}
}

func web(c *cli.Context) error {
	return nil
}
