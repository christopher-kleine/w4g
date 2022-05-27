package commands

import (
	"embed"

	"github.com/urfave/cli/v2"
)

//go:embed templates
var templates embed.FS

func Create() *cli.Command {
	return &cli.Command{
		Name:      "create",
		ArgsUsage: "<TEMPLATE> <PROJECT>",
		Usage:     "Creates a new WASM-4 project in a new directory",
	}
}
