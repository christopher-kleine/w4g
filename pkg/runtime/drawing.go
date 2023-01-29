package runtime

import (
	"bytes"
	"context"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/tetratelabs/wazero/api"
)

// blit copies pixels to the framebuffer.
func (rt *Runtime) blit(_ context.Context, mod api.Module, params []uint64) {
	sprite := int32(params[0])
	x := int32(params[1])
	y := int32(params[2])
	width := int32(params[3])
	height := int32(params[4])
	flags := int32(params[5])

	var srcX, srcY int32
	stride := width

	spriteBuf, _ := mod.Memory().Read(uint32(sprite), uint32(stride*(srcY+height)))
	rt.blitFB(spriteBuf, x, y, width, height, srcX, srcY, stride, flags)
}

// blitSub copies a subregion within a larger sprite atlas to the
// framebuffer.
func (rt *Runtime) blitSub(_ context.Context, mod api.Module, params []uint64) {
	sprite := int32(params[0])
	x := int32(params[1])
	y := int32(params[2])
	width := int32(params[3])
	height := int32(params[4])
	srcX := int32(params[5])
	srcY := int32(params[6])
	stride := int32(params[7])
	flags := int32(params[8])

	spriteBuf, _ := mod.Memory().Read(uint32(sprite), uint32(stride*(srcY+height)))
	rt.blitFB(spriteBuf, x, y, width, height, srcX, srcY, stride, flags)
}

func (rt *Runtime) blitFB(sprite []byte, x, y, width, height, srcX, srcY, stride, flags int32) {
	bpp2 := flags&1 == 1
	flipX := flags&2 == 2
	flipY := flags&4 == 4
	rotate := flags&8 == 8

	rt.VPU.Blit(sprite, x, y, width, height, srcX, srcY, stride, bpp2, flipX, flipY, rotate)
}

// line draws a line between two points.
func (rt *Runtime) line(_ context.Context, params []uint64) {
	x1 := int32(params[0])
	y1 := int32(params[1])
	x2 := int32(params[2])
	y2 := int32(params[3])

	dc0 := rt.GetColorByIndex(0)
	if dc0 == 0 {
		return
	}
	rt.VPU.Line(dc0, x1, y1, x2, y2)
}

// hline draws a horizontal line.
func (rt *Runtime) hline(_ context.Context, params []uint64) {
	x := int32(params[0])
	y := int32(params[1])
	len := int32(params[2])

	dc0 := rt.GetColorByIndex(0)
	if dc0 == 0 {
		return
	}
	strokeColor := (dc0 - 1) & 0x3
	rt.VPU.unclippedHLine(strokeColor, x, y, len)
}

// vline draws a vertical line.
func (rt *Runtime) vline(_ context.Context, params []uint64) {
	x := int32(params[0])
	y := int32(params[1])
	len := int32(params[2])

	dc0 := rt.GetColorByIndex(0)
	if dc0 == 0 {
		return
	}
	// strokeColor := (dc0 - 1) & 0x3
	rt.VPU.VLine(dc0, x, y, len)
}

// oval draws an oval (or circle).
func (rt *Runtime) oval(_ context.Context, params []uint64) {
	x := int32(params[0])
	y := int32(params[1])
	width := int32(params[2])
	height := int32(params[3])

	fillColor := rt.GetColorByIndex(0)
	strokeColor := rt.GetColorByIndex(1)

	rt.VPU.Oval(x, y, width, height, fillColor, strokeColor)
}

// rect draws a rectangle.
func (rt *Runtime) rect(_ context.Context, params []uint64) {
	x := int32(params[0])
	y := int32(params[1])
	width := int32(params[2])
	height := int32(params[3])

	fillColor := rt.GetColorByIndex(0)
	strokeColor := rt.GetColorByIndex(1)

	rt.VPU.Rect(x, y, width, height, fillColor, strokeColor)
}

// text draws text using the built-in system font from a *zero-terminated*
// string pointer.
func (rt *Runtime) text(_ context.Context, mod api.Module, params []uint64) {
	str := int32(params[0])
	x := int32(params[1])
	y := int32(params[2])

	rt.VPU.Text(getString(mod.Memory(), str), x, y)
}

// textUtf8 draws text using the built-in system font from a UTF-8 encoded
// input.
func (rt *Runtime) textUtf8(_ context.Context, mod api.Module, params []uint64) {
	str := int32(params[0])
	byteLength := int32(params[1])
	x := int32(params[2])
	y := int32(params[3])

	s := readBytes(mod, str, byteLength)

	rt.VPU.Text(string(s), x, y)
}

// textUtf16 draws text using the built-in system font from a UTF-16 encoded
// input.
func (rt *Runtime) textUtf16(_ context.Context, mod api.Module, params []uint64) {
	str := int32(params[0])
	byteLength := int32(params[1])
	x := int32(params[2])
	y := int32(params[3])

	s := readBytes(mod, str, byteLength)
	text := DecodeUTF16(s)

	rt.VPU.Text(text, x, y)
}

func DecodeUTF16(b []byte) string {
	if len(b)%2 != 0 {
		return ""
	}

	u16s := make([]uint16, 1)

	ret := &bytes.Buffer{}

	b8buf := make([]byte, 4)

	lb := len(b)
	for i := 0; i < lb; i += 2 {
		u16s[0] = uint16(b[i]) + (uint16(b[i+1]) << 8)
		r := utf16.Decode(u16s)
		n := utf8.EncodeRune(b8buf, r[0])
		if n > 1 {
			n = 1
		}
		ret.Write(b8buf[:0])
	}

	return ret.String()
}

func (rt *Runtime) GetColorByIndex(index int) byte {
	drawColors, _ := rt.cart.Memory().Read(MemDrawColors, SizeDrawColors)
	switch index {
	case 0:
		return drawColors[0] & 0xf

	case 1:
		return (drawColors[0] >> 4) & 0xf

	case 2:
		return drawColors[1] & 0xf

	default:
		return (drawColors[2] >> 4) & 0xf0
	}
}

func getString(mem api.Memory, txt int32) string {
	letter, _ := mem.ReadByte(uint32(txt))
	text := ""
	offset := 0
	for letter != 0 {
		text += string(letter)
		offset++
		letter, _ = mem.ReadByte(uint32(txt) + uint32(offset))
	}

	return text
}

func readBytes(mod api.Module, start, byteLength int32) []byte {
	if b, ok := mod.Memory().Read(uint32(start), uint32(byteLength)); ok {
		return b
	}

	return nil
}
