package commands

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

func Install() *cli.Command {
	return &cli.Command{
		Name:      "install",
		Usage:     "Installs the given package",
		ArgsUsage: "<PACKAGE>",
		Action:    install,
	}
}

func install(c *cli.Context) error {
	if !c.Args().Present() {
		return fmt.Errorf("at least one package must be provided")
	}

	return nil
}
