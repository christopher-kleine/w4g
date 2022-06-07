package main

import (
	"log"
	"os"

	"github.com/christopher-kleine/w4g/cmd/w4g/commands"
	"github.com/urfave/cli/v2"
)

var (
	GitCommit string
)

func main() {
	app := &cli.App{
		Name:    "w4g",
		Usage:   "Chris' version of the WASM-4 CLI",
		Version: GitCommit,
		Action:  commands.SurfAction,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "fps",
				Usage: "Show the current FPS",
				Value: false,
			},
			&cli.IntFlag{
				Name:  "scale",
				Usage: "Sets the window scale compared to the game",
				Value: 5,
			},
			&cli.StringFlag{
				Name:    "encoder",
				Aliases: []string{"enc"},
				Usage:   "Encoder for video recordings (y4m, mjpeg)",
				Value:   "y4m",
			},
			&cli.IntFlag{
				Name:    "quality",
				Aliases: []string{"q"},
				Usage:   "Quality setting for the MJPEG encoder",
				Value:   80,
			},
		},
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
