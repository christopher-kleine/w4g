package commands

import (
	"github.com/urfave/cli/v2"
)

func Surf() *cli.Command {
	return &cli.Command{
		Name:   "surf",
		Usage:  "Shows all games available online (Default command)",
		Action: SurfAction,
	}
}

func SurfAction(c *cli.Context) error {
	return nil
}
