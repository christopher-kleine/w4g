package runtime

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// This file implements direct access to the framebuffer.
// Other Drawing functions may use them.

func (rt *Runtime) ClearFB() {
	for pos := MemFramebuffer; pos < MemFramebuffer+SizeFramebuffer; pos++ {
		rt.env.Memory().WriteByte(rt.ctx, pos, 0)
	}
}

func (rt *Runtime) RenderFB(screen *ebiten.Image) {
	palette, _ := rt.env.Memory().Read(rt.ctx, MemPalette, SizePalette)
	colors := []color.Color{
		color.RGBA{
			A: 0xff,
			R: palette[2],
			G: palette[1],
			B: palette[0],
		},
		color.RGBA{
			A: 0xff,
			R: palette[2+4],
			G: palette[1+4],
			B: palette[0+4],
		},
		color.RGBA{
			A: 0xff,
			R: palette[2+8],
			G: palette[1+8],
			B: palette[0+8],
		},
		color.RGBA{
			A: 0xff,
			R: palette[2+12],
			G: palette[1+12],
			B: palette[0+12],
		},
	}

	framebuffer, _ := rt.env.Memory().Read(rt.ctx, MemFramebuffer, SizeFramebuffer)
	for offY, pixel := range framebuffer {
		colorIndex := []byte{
			pixel & 3,
			(pixel >> 2) & 3,
			(pixel >> 4) & 3,
			(pixel >> 6) & 3,
		}

		x := (offY * 4) % 160
		y := (offY * 4) / 160

		for offX, index := range colorIndex {
			screen.Set(int(x)+offX, int(y), colors[index])
		}
	}
}

func (rt *Runtime) BlitFB(sprite []byte, dstX, dstY, w, h, srcX, srcY, stride int32, bpp2, flipX, flipY, rotate bool) {
	drawColors, _ := rt.env.Memory().Read(rt.ctx, MemDrawColors, SizeDrawColors)

	var (
		colors             uint16 = uint16(drawColors[0]) | (uint16(drawColors[1]) << 8)
		clipXMin, clipYMin int32
		clipXMax, clipYMax int32
		tx, ty             int32
		sx, sy             int32
		colorIdx           int32
		bitIndex           int32
		bite               uint8
		shift              int32
		dc                 uint8
	)

	if rotate {
		flipX = !flipX

		clipXMin = 0
		if dstY > 0 {
			clipXMin = dstY
		}
		clipXMin -= dstY

		clipYMin = 0
		if dstX > 0 {
			clipYMin = dstX
		}
		clipYMin -= dstX

		clipXMax = w
		if min := 160 - dstX; min < clipXMax {
			clipXMax = min
		}

		clipYMax = h
		if min := 160 - dstY; min < clipYMax {
			clipYMax = min
		}
	} else {
		clipXMin = 0
		if dstX > 0 {
			clipXMin = dstX
		}
		clipXMin -= dstX

		clipYMin = 0
		if dstY > 0 {
			clipYMin = dstY
		}
		clipYMin -= dstY

		clipXMax = w
		if min := 160 - dstX; min < clipXMax {
			clipXMax = min
		}

		clipYMax = h
		if min := 160 - dstY; min < clipYMax {
			clipYMax = min
		}
	}

	for y := clipYMin; y < clipYMax; y++ {
		for x := clipXMin; x < clipXMax; x++ {
			tx = dstX + x
			ty = dstY + y
			if rotate {
				tx = dstX + y
				ty = dstY + x
			}

			sx = srcX + x
			if flipX {
				sx = srcX + (w - x - 1)
			}

			sy = srcY + y
			if flipY {
				sy = srcY + (h - y - 1)
			}

			bitIndex = sy*stride + sx
			if bpp2 {
				bite = sprite[bitIndex>>2]
				shift = 6 - ((bitIndex & 0x03) << 1)
				colorIdx = int32((bite >> shift) & 0x3)
			} else {
				bite = sprite[bitIndex>>3]
				shift = 7 - (bitIndex & 0x07)
				colorIdx = int32((bite >> shift) & 0x1)
			}

			dc = uint8((colors >> (colorIdx << 2)) & 0x0f)
			if dc != 0 {
				rt.PointFB(dc-1, tx, ty)
			}
		}
	}
}

func (rt *Runtime) LineFB(x1, y1, x2, y2 int32) {
	drawColors, _ := rt.env.Memory().Read(rt.ctx, MemDrawColors, SizeDrawColors)
	var dc0 uint8 = drawColors[0] & 0xf
	if dc0 == 0 {
		return
	}
	var strokeColor uint8 = (dc0 - 1) & 0x3

	if y1 > y2 {
		x1, x2 = x2, x1
		y1, y2 = y2, y1
	}

	var dx int32 = int32(math.Abs(float64(x2 - x1)))
	var sx int32 = -1
	if x1 < x2 {
		sx = 1
	}
	var dy int32 = y2 - y1
	var err int32 = -dy
	if dx > dy {
		err = dx
	}
	err = err / 2

	for {
		rt.PointUnclippedFB(strokeColor, x1, y1)
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := err
		if e2 > -dx {
			err -= dy
			x1 += sx
		}
		if e2 < dy {
			err += dx
			y1++
		}
	}
}

func (rt *Runtime) HLineFB(x, y, endX int32) {
	rt.LineFB(x, y, x+endX, y)
}

func (rt *Runtime) VLineFB(x, y, endY int32) {
	rt.LineFB(x, y, x, y+endY)
}

func (rt *Runtime) PointFB(color byte, x, y int32) {
	var (
		idx             int32 = (y*160 + x) >> 2
		currentValue, _       = rt.env.Memory().ReadByte(rt.ctx, uint32(idx)+MemFramebuffer)
		shift           int32 = (x & 0x3) << 1
		mask            int32 = 0x3 << shift
		value           uint8 = uint8((int32(color) << shift) | (int32(currentValue) & ^mask))
	)
	rt.env.Memory().WriteByte(rt.ctx, uint32(idx)+MemFramebuffer, value)
}

func (rt *Runtime) PointUnclippedFB(colorIndex byte, x, y int32) {
	if x >= 0 && x < 160 && y >= 0 && y < 160 {
		rt.PointFB(colorIndex, x, y)
	}
}

func (rt *Runtime) TextFB(txt string, x, y int32) {
	currX := x
	currY := y
	for _, letter := range txt {
		switch letter {
		case 0:
			return

		case '\n':
			currX = x
			currY += 8

		default:
			rt.BlitFB(font, currX, currY, 8, 8, 0, (letter-32)<<3, 8, false, false, false, false)
			currX += 8
		}
	}
}
