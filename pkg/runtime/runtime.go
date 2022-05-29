package runtime

import (
	"context"
	"fmt"

	"github.com/christopher-kleine/w4g/pkg/encoders"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

const (
	MemPalette      uint32 = 0x0004
	MemDrawColors   uint32 = 0x0014
	MemGamepads     uint32 = 0x0016
	MemMouseX       uint32 = 0x001a
	MemMouseY       uint32 = 0x001c
	MemMouseButtons uint32 = 0x001e
	MemSystemFlags  uint32 = 0x001f
	MemReserved     uint32 = 0x0020
	MemFramebuffer  uint32 = 0x00a0
	MemUser         uint32 = 0x19a0
)

const (
	SizePalette      uint32 = 16
	SizeDrawColors   uint32 = 2
	SizeGamepads     uint32 = 1
	SizeMouseX       uint32 = 2
	SizeMouseY       uint32 = 2
	SizeMouseButtons uint32 = 1
	SizeSystemFlags  uint32 = 1
	SizeReserved     uint32 = 128
	SizeFramebuffer  uint32 = 6400
	SizeUser         uint32 = 58976
)

const (
	FlagPreserveScreen byte = 1
)

const (
	PadIdle  byte = 0
	PadX     byte = 1
	PadY     byte = 1 << 1
	PadLeft  byte = 1 << 4
	PadRight byte = 1 << 5
	PadUp    byte = 1 << 6
	PadDown  byte = 1 << 7
)

var (
	Player1Keys = map[ebiten.Key]byte{
		ebiten.KeyLeft:  PadLeft,
		ebiten.KeyRight: PadRight,
		ebiten.KeyUp:    PadUp,
		ebiten.KeyDown:  PadDown,
		ebiten.KeyX:     PadX,
		ebiten.KeySpace: PadX,
		ebiten.KeyY:     PadY,
		ebiten.KeyZ:     PadY,
		ebiten.KeyC:     PadY,
	}
	Player2Keys = map[ebiten.Key]byte{
		ebiten.KeyS:   PadLeft,
		ebiten.KeyF:   PadRight,
		ebiten.KeyE:   PadUp,
		ebiten.KeyD:   PadDown,
		ebiten.KeyQ:   PadX,
		ebiten.KeyTab: PadY,
	}
)

type Runtime struct {
	runtime wazero.Runtime
	env     api.Module
	cart    api.Module
	ctx     context.Context
	showFPS bool
	Encoder encoders.Encoder
}

func NewRuntime(showFPS bool) (*Runtime, error) {
	var err error

	result := &Runtime{
		showFPS: showFPS,
	}

	result.ctx = context.Background()

	result.runtime = wazero.NewRuntime()

	builder := result.runtime.NewModuleBuilder("env")
	result.env, err = builder.
		// Drawing
		ExportFunction("blit", result.Blit).
		ExportFunction("blitSub", result.BlitSub).
		ExportFunction("line", result.Line).
		ExportFunction("hline", result.HLine).
		ExportFunction("vline", result.VLine).
		ExportFunction("oval", result.Oval).
		ExportFunction("rect", result.Rect).
		ExportFunction("text", result.Text).
		ExportFunction("textUtf8", result.TextUTF8).
		// Sound
		ExportFunction("tone", result.Tone).
		// Storage
		ExportFunction("diskw", result.DiskW).
		ExportFunction("diskr", result.DiskR).
		// Other
		ExportFunction("trace", result.Trace).
		ExportFunction("traceUtf8", result.TraceUtf8).
		// Memory
		ExportMemoryWithMax("memory", 1, 1).
		Instantiate(result.ctx)

	if err != nil {
		return result, err
	}

	return result, nil
}

func (rt *Runtime) LoadCart(code []byte) error {
	var err error

	rt.env.Memory().Write(rt.ctx, MemPalette, []byte{
		0xcf, 0xf8, 0xe0, 0,
		0x6c, 0xc0, 0x86, 0,
		0x50, 0x68, 0x30, 0,
		0x21, 0x18, 0x07, 0,
	})

	rt.cart, err = rt.runtime.InstantiateModuleFromCode(rt.ctx, code)
	if err != nil {
		return err
	}

	fn := rt.cart.ExportedFunction("start")
	if fn != nil {
		fn.Call(rt.ctx)
	}

	return nil
}

func (rt *Runtime) Close() {
	if rt.env != nil {
		rt.env.Close(rt.ctx)
	}

	if rt.cart != nil {
		rt.cart.Close(rt.ctx)
	}
}

func (rt *Runtime) Draw(screen *ebiten.Image) {
	fmt.Print("\r")
	rt.RenderFB(screen)
	if rt.showFPS {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%.f", ebiten.CurrentFPS()), 0, 0)
	}
	//screen.DrawImage(img, nil)
}

func (rt *Runtime) KeyState(keys map[ebiten.Key]byte) byte {
	result := PadIdle

	for key, value := range keys {
		if ebiten.IsKeyPressed(key) {
			result = result | value
		}
	}

	return result
}

func (rt *Runtime) Update() error {
	SystemFlags, _ := rt.env.Memory().ReadByte(rt.ctx, MemSystemFlags)
	if SystemFlags&FlagPreserveScreen == 0 {
		rt.ClearFB()
	}

	rt.env.Memory().WriteByte(rt.ctx, MemGamepads+0*SizeGamepads, rt.KeyState(Player1Keys))
	rt.env.Memory().WriteByte(rt.ctx, MemGamepads+1*SizeGamepads, rt.KeyState(Player2Keys))
	// rt.env.Memory().WriteByte(rt.ctx, MemGamepads+2*SizeGamepads, rt.KeyState(Player3Keys))
	// rt.env.Memory().WriteByte(rt.ctx, MemGamepads+3*SizeGamepads, rt.KeyState(Player4Keys))

	_, err := rt.cart.ExportedFunction("update").Call(rt.ctx)
	if err != nil {
		return err
	}

	return nil
}

func (rt *Runtime) Layout(int, int) (int, int) { return 160, 160 }
