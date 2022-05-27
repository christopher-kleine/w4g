package commands

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

func Init() *cli.Command {
	return &cli.Command{
		Name:   "init",
		Usage:  "Creates a new WASM-4 project inside the current directory",
		Action: initCmd,
	}
}

func initCmd(c *cli.Context) error {
	if len(c.FlagNames()) == 0 {
		return fmt.Errorf("a language is required")
	}

	return nil
}
