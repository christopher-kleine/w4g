package commands

import (
	"errors"
	"os"

	"github.com/christopher-kleine/w4g/pkg/runtime"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/urfave/cli/v2"
	"github.com/zserge/lorca"
)

func Run() *cli.Command {
	return &cli.Command{
		Name:      "run",
		Usage:     "Starts a WASM-4 cart in the native client",
		ArgsUsage: "<CART>",
		Subcommands: []*cli.Command{
			{
				Name:      "web",
				Usage:     "Starts a WASM-4 cart in chrome/chromium/edge",
				Action:    runWeb,
				ArgsUsage: "<CART>",
				Flags:     nil,
			},
			{
				Name:      "native",
				Usage:     "Starts a WASM-4 cart in the native client",
				Action:    runNative,
				ArgsUsage: "<CART>",
				UsageText: "DASD",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "fps",
						Usage: "Show the current FPS",
						Value: false,
					},
				},
			},
		},
	}
}

func runWeb(c *cli.Context) error {
	ui, err := lorca.New("", "", 640, 640)
	if err != nil {
		return err
	}

	<-ui.Done()

	return nil
}

func runNative(c *cli.Context) error {
	cart := c.Args().First()
	if cart == "" {
		return errors.New("no file provided")
	}

	rt, err := runtime.NewRuntime(true)
	if err != nil {
		return err
	}

	code, err := os.ReadFile(cart)
	if err != nil {
		return err
	}

	err = rt.LoadCart(code)
	if err != nil {
		return err
	}

	ebiten.SetWindowSize(160*5, 160*5)
	err = ebiten.RunGame(rt)

	return err
}
