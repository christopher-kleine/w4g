package runtime

import (
	"bytes"
	"context"
	"fmt"
	"log"
	uutf16 "unicode/utf16"
	uutf8 "unicode/utf8"

	"github.com/tetratelabs/wazero/api"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
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

	rt.BlitFB(sprite, x, y, width, height, srcX, srcY, stride, bpp2, flipX, flipY, rotate)
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
	var strokeColor uint8 = (dc0 - 1) & 0x3
	rt.LineFB(strokeColor, x1, y1, x2, y2)
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
	rt.HLineFB(strokeColor, x, y, len)
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
	strokeColor := (dc0 - 1) & 0x3
	rt.VLineFB(strokeColor, x, y, len)
}

// oval draws an oval (or circle).
func (rt *Runtime) oval(_ context.Context, params []uint64) {
	x := int32(params[0])
	y := int32(params[1])
	width := int32(params[2])
	height := int32(params[3])

	rt.OvalFB(x, y, width, height)
}

// rect draws a rectangle.
func (rt *Runtime) rect(_ context.Context, params []uint64) {
	x := int32(params[0])
	y := int32(params[1])
	width := int32(params[2])
	height := int32(params[3])

	rt.RectFB(x, y, width, height)
}

// text draws text using the built-in system font from a *zero-terminated*
// string pointer.
func (rt *Runtime) text(_ context.Context, mod api.Module, params []uint64) {
	str := int32(params[0])
	x := int32(params[1])
	y := int32(params[2])

	rt.TextFB(getString(mod.Memory(), str), x, y)
}

// textUtf8 draws text using the built-in system font from a UTF-8 encoded
// input.
func (rt *Runtime) textUtf8(_ context.Context, mod api.Module, params []uint64) {
	str := int32(params[0])
	byteLength := int32(params[1])
	x := int32(params[2])
	y := int32(params[3])

	log.Printf("textUtf8: str=%d, byteLength=%d, x=%d, y=%d", str, byteLength, x, y)

	s := mustDecode(mod, utf8, str, byteLength, "str")

	rt.TextFB(s, x, y)
}

// textUtf16 draws text using the built-in system font from a UTF-16 encoded
// input.
func (rt *Runtime) textUtf16(_ context.Context, mod api.Module, params []uint64) {
	str := int32(params[0])
	byteLength := int32(params[1])
	x := int32(params[2])
	y := int32(params[3])

	log.Printf("textUtf16: str=%d, byteLength=%d, x=%d, y=%d", str, byteLength, x, y)

	s := mustDecode(mod, utf16, str, byteLength, "str")
	// s, _ = DecodeUTF16([]byte(s))

	rt.TextFB(s, x, y)
}

func DecodeUTF16(b []byte) (string, error) {

	if len(b)%2 != 0 {
		return "", fmt.Errorf("Must have even length byte slice")
	}

	u16s := make([]uint16, 1)

	ret := &bytes.Buffer{}

	b8buf := make([]byte, 4)

	lb := len(b)
	for i := 0; i < lb; i += 2 {
		u16s[0] = uint16(b[i]) + (uint16(b[i+1]) << 8)
		r := uutf16.Decode(u16s)
		n := uutf8.EncodeRune(b8buf, r[0])
		ret.Write(b8buf[:n])
	}

	return ret.String(), nil
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

var (
	utf8  = unicode.UTF8
	utf16 = unicode.UTF16(unicode.BigEndian, unicode.UseBOM)
)

func mustDecode(mod api.Module, encoding encoding.Encoding, str, byteLength int32, field string) (s string) {
	var err error
	log.Printf("mustDecode: str=%d, byteLength=%d, field=%s", str, byteLength, field)
	if b, ok := mod.Memory().Read(uint32(str), uint32(byteLength)); !ok {
		log.Printf("out of memory reading %s", field)
	} else if encoding == utf8 {
		return string(b)
	} else if s, err = encoding.NewDecoder().String(string(b)); err != nil {
		log.Printf("error reading %s: %v", field, err)
	}
	log.Printf("mustDecode: returning %q", s)
	return
}
