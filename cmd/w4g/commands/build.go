package commands

import (
	"github.com/urfave/cli/v2"
)

func Build() *cli.Command {
	return &cli.Command{
		Name:  "build",
		Usage: "Starts the build process of the project",
		Subcommands: []*cli.Command{
			{
				Name:  "native",
				Usage: "Builds the project for native targets",
			},
		},
		Action: build,
	}
}

func build(c *cli.Context) error {
	return nil
}
