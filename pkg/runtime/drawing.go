package runtime

import (
	"context"
	"github.com/tetratelabs/wazero/api"
)

func (rt *Runtime) Blit(ctx context.Context, mod api.Module, spr, x, y, w, h, f int32) {
	rt.BlitSub(ctx, mod, spr, x, y, w, h, 0, 0, w, f)
}

func (rt *Runtime) BlitSub(ctx context.Context, mod api.Module, spr, x, y, w, h, srcX, srcY, stride, f int32) {
	sprite, _ := mod.Memory().Read(uint32(spr), uint32(stride*(srcY+h)))

	bpp2 := f&1 == 1
	flipX := f&2 == 2
	flipY := f&4 == 4
	rotate := f&8 == 8

	rt.BlitFB(sprite, x, y, w, h, srcX, srcY, stride, bpp2, flipX, flipY, rotate)
}

func (rt *Runtime) Line(x1, y1, x2, y2 int32) {
	dc0 := rt.GetColorByIndex(0)
	if dc0 == 0 {
		return
	}
	var strokeColor uint8 = (dc0 - 1) & 0x3
	rt.LineFB(strokeColor, x1, y1, x2, y2)
}

func (rt *Runtime) HLine(x, y, l int32) {
	dc0 := rt.GetColorByIndex(0)
	if dc0 == 0 {
		return
	}
	strokeColor := (dc0 - 1) & 0x3
	rt.HLineFB(strokeColor, x, y, l)
}

func (rt *Runtime) VLine(x, y, l int32) {
	dc0 := rt.GetColorByIndex(0)
	if dc0 == 0 {
		return
	}
	strokeColor := (dc0 - 1) & 0x3
	rt.VLineFB(strokeColor, x, y, l)
}

func (rt *Runtime) Oval(x, y, w, h int32) {
}

func (rt *Runtime) Rect(x, y, w, h int32) {
	rt.RectFB(x, y, w, h)
}

func (rt *Runtime) Text(ctx context.Context, mod api.Module, txt, x, y int32) {
	rt.TextFB(rt.getString(ctx, mod, txt), x, y)
}

func (rt *Runtime) TextUTF8(ctx context.Context, mod api.Module, textPtr, byteLength, x, y int32) {
	// const text = new Uint8Array(this.memory.buffer, textPtr, byteLength);
	// this.framebuffer.drawText(text, x, y);
	text, _ := mod.Memory().Read(uint32(textPtr), uint32(byteLength))
	rt.TextFB(string(text), x, y)
}

func (rt *Runtime) getString(ctx context.Context, mod api.Module, txt int32) string {
	letter, _ := mod.Memory().ReadByte(uint32(txt))
	text := ""
	offset := 0
	for letter != 0 {
		text += string(letter)
		offset++
		letter, _ = mod.Memory().ReadByte(uint32(txt) + uint32(offset))
	}

	return text
}
