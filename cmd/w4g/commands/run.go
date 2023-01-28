package commands

import (
	"errors"
	"fmt"
	"os"

	"github.com/christopher-kleine/lorca"
	"github.com/christopher-kleine/w4g/pkg/encoders"
	"github.com/christopher-kleine/w4g/pkg/runtime"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/urfave/cli/v2"
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

	showFPS := c.Bool("fps")
	scale := c.Int("scale")

	rt, err := runtime.NewRuntime(showFPS)
	if err != nil {
		return err
	}

	enc := c.String("encoder")
	switch enc {
	case "y4m":
		rt.Encoder = encoders.NewY4M()

	case "mjpeg":
		rt.Encoder = encoders.NewMJPEG(c.Int("quality"))

	default:
		return fmt.Errorf("unknown encoder %q selected", enc)
	}

	code, err := os.ReadFile(cart)
	if err != nil {
		return err
	}

	err = rt.LoadCart(code, cart)
	if err != nil {
		return err
	}

	ebiten.SetWindowSize(160*scale, 160*scale)
	ebiten.SetWindowTitle("WASM-4 (Go)")
	ebiten.SetMaxTPS(60)
	ebiten.SetFPSMode(ebiten.FPSModeVsyncOn)
	err = ebiten.RunGame(rt)

	return rt.Close()
}
