package main

import (
	"log"
	"os"

	"github.com/christopher-kleine/w4g/cmd/w4g/commands"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:                 "w4g",
		Usage:                "Chris' version of the WASM-4 CLI",
		Version:              "0.0.1",
		Action:               commands.SurfAction,
		EnableBashCompletion: true,
		Authors: []*cli.Author{
			{
				Name:  "Christopher Kleine",
				Email: "chris@suletuxe.de",
			},
			{
				Name:  "Bruno Garcia",
				Email: "b@aduros.com",
			},
		},
		Copyright: "(c) 2022 by Christopher Kleine",
		Commands: []*cli.Command{
			commands.Create(),
			commands.Init(),
			//commands.Watch(),
			//commands.Web(),
			commands.Run(),
			//commands.Img2Src(),
			//commands.Install(),
			//commands.Build(),
			//commands.Bundle(),
			commands.Surf(),
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
