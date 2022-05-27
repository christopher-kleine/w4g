package commands

import (
	"github.com/urfave/cli/v2"
)

func Watch() *cli.Command {
	return &cli.Command{
		Name:   "watch",
		Usage:  "Starts the build process and watches for changes",
		Action: watch,
	}
}

func watch(c *cli.Context) error {
	return nil
}
