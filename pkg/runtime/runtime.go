package runtime

import (
	"context"
	_ "embed"
	"fmt"
	"image/png"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/christopher-kleine/w4g/pkg/encoders"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
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

// envWasm was compiled using `cd wasm; wat2wasm --debug-names env.wat`
//
//go:embed wasm/env.wasm
var envWasm []byte

type Runtime struct {
	runtime  wazero.Runtime
	cart     api.Module
	cartName string
	ctx      context.Context
	showFPS  bool
	Encoder  encoders.Encoder
	VPU      *VPU
	APU      *APU
	Storage  io.ReadWriter
}

var (
	i32 = api.ValueTypeI32
)

func NewRuntime(showFPS bool) (*Runtime, error) {
	var err error

	result := &Runtime{
		showFPS: showFPS,
	}

	result.ctx = context.Background()

	result.runtime = wazero.NewRuntime(result.ctx)

	builder := result.runtime.NewHostModuleBuilder("goenv")
	_, err = builder.
		// Drawing
		NewFunctionBuilder().
		WithGoModuleFunction(api.GoModuleFunc(result.blit), []api.ValueType{i32, i32, i32, i32, i32, i32}, []api.ValueType{}).
		WithParameterNames("sprite", "x", "y", "width", "height", "flags").
		Export("blit").
		NewFunctionBuilder().
		WithGoModuleFunction(api.GoModuleFunc(result.blitSub), []api.ValueType{i32, i32, i32, i32, i32, i32, i32, i32, i32}, []api.ValueType{}).
		WithParameterNames("sprite", "x", "y", "width", "height", "srcX", "srcY", "stride", "flags").
		Export("blitSub").
		NewFunctionBuilder().
		WithGoFunction(api.GoFunc(result.line), []api.ValueType{i32, i32, i32, i32}, []api.ValueType{}).
		WithParameterNames("x1", "y1", "x2", "y2").
		Export("line").
		NewFunctionBuilder().
		WithGoFunction(api.GoFunc(result.hline), []api.ValueType{i32, i32, i32}, []api.ValueType{}).
		WithParameterNames("x", "y", "len").
		Export("hline").
		NewFunctionBuilder().
		WithGoFunction(api.GoFunc(result.vline), []api.ValueType{i32, i32, i32}, []api.ValueType{}).
		WithParameterNames("x", "y", "len").
		Export("vline").
		NewFunctionBuilder().
		WithGoFunction(api.GoFunc(result.oval), []api.ValueType{i32, i32, i32, i32}, []api.ValueType{}).
		WithParameterNames("x", "y", "width", "height").
		Export("oval").
		NewFunctionBuilder().
		WithGoFunction(api.GoFunc(result.rect), []api.ValueType{i32, i32, i32, i32}, []api.ValueType{}).
		WithParameterNames("x", "y", "width", "height").
		Export("rect").
		NewFunctionBuilder().
		WithGoModuleFunction(api.GoModuleFunc(result.text), []api.ValueType{i32, i32, i32}, []api.ValueType{}).
		WithParameterNames("str", "x", "y").
		Export("text").
		NewFunctionBuilder().
		WithGoModuleFunction(api.GoModuleFunc(result.textUtf8), []api.ValueType{i32, i32, i32, i32}, []api.ValueType{}).
		WithParameterNames("str", "byteLength", "x", "y").
		Export("textUtf8").
		NewFunctionBuilder().
		WithGoModuleFunction(api.GoModuleFunc(result.textUtf16), []api.ValueType{i32, i32, i32, i32}, []api.ValueType{}).
		WithParameterNames("str", "byteLength", "x", "y").
		Export("textUtf16").
		//// Sound
		NewFunctionBuilder().
		WithGoFunction(api.GoFunc(result.tone), []api.ValueType{i32, i32, i32, i32}, []api.ValueType{}).
		WithParameterNames("frequency", "duration", "volume", "flags").
		Export("tone").
		//// Storage
		NewFunctionBuilder().
		WithGoModuleFunction(api.GoModuleFunc(result.diskr), []api.ValueType{i32, i32}, []api.ValueType{i32}).
		WithParameterNames("dest", "size").
		Export("diskr").
		NewFunctionBuilder().
		WithGoModuleFunction(api.GoModuleFunc(result.diskw), []api.ValueType{i32, i32}, []api.ValueType{i32}).
		WithParameterNames("src", "size").
		Export("diskw").
		//// Other
		NewFunctionBuilder().
		WithGoModuleFunction(api.GoModuleFunc(result.trace), []api.ValueType{i32}, []api.ValueType{}).
		WithParameterNames("str").
		Export("trace").
		NewFunctionBuilder().
		WithGoModuleFunction(api.GoModuleFunc(result.traceUtf8), []api.ValueType{i32, i32}, []api.ValueType{}).
		WithParameterNames("str", "byteLength").
		Export("traceUtf8").
		NewFunctionBuilder().
		WithGoModuleFunction(api.GoModuleFunc(result.traceUtf16), []api.ValueType{i32, i32}, []api.ValueType{}).
		WithParameterNames("str", "byteLength").
		Export("traceUtf16").
		NewFunctionBuilder().
		WithGoModuleFunction(api.GoModuleFunc(result.tracef), []api.ValueType{i32, i32}, []api.ValueType{}).
		WithParameterNames("str", "byteLength").
		Export("tracef").
		Instantiate(result.ctx, result.runtime)

	if err != nil {
		result.runtime.Close(result.ctx)
		return nil, err
	}

	_, err = result.runtime.InstantiateModuleFromBinary(result.ctx, envWasm)

	if err != nil {
		result.runtime.Close(result.ctx)
		return nil, err
	}

	return result, nil
}

func (rt *Runtime) LoadCart(code []byte, name string) error {
	var err error

	rt.cartName = name

	rt.cart, err = rt.runtime.InstantiateModuleFromBinary(rt.ctx, code)
	if err != nil {
		return err
	}

	rt.VPU = &VPU{
		Memory: rt.cart.Memory,
	}
	rt.APU = &APU{}
	rt.Storage = &Storage{
		Data: make([]byte, 1024),
	}

	fn := rt.cart.ExportedFunction("start")
	if fn != nil {
		fn.Call(rt.ctx)
	}

	return nil
}

func (rt *Runtime) Close() {
	if rt.runtime != nil {
		rt.runtime.Close(rt.ctx)
	}
}

func (rt *Runtime) Screenshot(screen *ebiten.Image) {
	udir, err := os.UserHomeDir()
	if err != nil {
		log.Println(err)
		return
	}

	fname := fmt.Sprintf("%s_%v.png", rt.cartName, time.Now().Format("2006-01-02_15-04-05"))
	fname = filepath.Join(udir, fname)
	f, err := os.Create(fname)
	if err != nil {
		log.Println(err)
		return
	}

	err = png.Encode(f, screen)
	if err != nil {
		log.Println(err)
	}
}

func (rt *Runtime) Draw(screen *ebiten.Image) {
	fmt.Print("\r")
	rt.VPU.RenderFB(screen)

	if inpututil.IsKeyJustPressed(ebiten.KeyF9) {
		rt.Screenshot(screen)
	}

	if rt.Encoder.IsRunning() {
		rt.Encoder.Encode(screen)
	}

	if rt.showFPS {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%.f", ebiten.CurrentFPS()), 0, 0)
	}

	if rt.Encoder.IsRunning() {
		ebitenutil.DebugPrintAt(screen, "REC", 160-24, 0)
	}
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
	SystemFlags, _ := rt.cart.Memory().ReadByte(MemSystemFlags)
	if SystemFlags&FlagPreserveScreen == 0 {
		rt.VPU.Clear()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyF10) {
		if rt.Encoder.IsRunning() {
			rt.Encoder.Stop()
		} else {
			rt.Encoder.Start(rt.cartName)
		}
	}

	rt.cart.Memory().WriteByte(MemGamepads+0*SizeGamepads, rt.KeyState(Player1Keys))
	rt.cart.Memory().WriteByte(MemGamepads+1*SizeGamepads, rt.KeyState(Player2Keys))
	// rt.cart.Memory().WriteByte(MemGamepads+2*SizeGamepads, rt.KeyState(Player3Keys))
	// rt.cart.Memory().WriteByte(MemGamepads+3*SizeGamepads, rt.KeyState(Player4Keys))

	_, err := rt.cart.ExportedFunction("update").Call(rt.ctx)
	if err != nil {
		return err
	}

	return nil
}

func (rt *Runtime) Layout(int, int) (int, int) { return 160, 160 }
