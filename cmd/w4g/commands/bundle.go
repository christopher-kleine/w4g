package commands

import (
	"github.com/urfave/cli/v2"
)

func Bundle() *cli.Command {
	return &cli.Command{
		Name:   "surf",
		Usage:  "Shows all games on the website and let's you play them",
		Action: bundle,
	}
}

func bundle(c *cli.Context) error {
	return nil
}
