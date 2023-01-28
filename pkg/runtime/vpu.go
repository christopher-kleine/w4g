package runtime

import (
	"bytes"
	"image/color"
	"math"

	"github.com/christopher-kleine/w4g/pkg/tools"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/tetratelabs/wazero/api"
)

const (
	WIDTH  = 160
	HEIGHT = 160
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

type VPU struct {
	Memory func() api.Memory
}

// This file implements direct access to the framebuffer.
// Other Drawing functions may use them.

func (vpu *VPU) Init() {
	colors, _ := vpu.Memory().Read(MemPalette, 16)
	if bytes.Equal(colors, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}) {
		vpu.Memory().Write(MemPalette, []byte{
			0xcf, 0xf8, 0xe0, 0xff,
			0x6c, 0xc0, 0x86, 0xff,
			0x50, 0x68, 0x30, 0xff,
			0x21, 0x18, 0x07, 0xff,
		})
	}
}

func (vpu *VPU) Clear() {
	for pos := MemFramebuffer; pos < MemFramebuffer+SizeFramebuffer; pos++ {
		vpu.Memory().WriteByte(pos, 0)
	}
}

func (vpu *VPU) RenderFB(screen *ebiten.Image) {
	palette, _ := vpu.Memory().Read(MemPalette, SizePalette)
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

	framebuffer, _ := vpu.Memory().Read(MemFramebuffer, SizeFramebuffer)
	for offY, pixel := range framebuffer {
		colorIndex := []byte{
			pixel & 3,
			(pixel >> 2) & 3,
			(pixel >> 4) & 3,
			(pixel >> 6) & 3,
		}

		x := (offY * 4) % WIDTH
		y := (offY * 4) / WIDTH

		for offX, index := range colorIndex {
			screen.Set(int(x)+offX, int(y), colors[index])
		}
	}
}

func (vpu *VPU) BlitFB(sprite []byte, dstX, dstY, w, h, srcX, srcY, stride int32, bpp2, flipX, flipY, rotate bool) {
	drawColors, _ := vpu.Memory().Read(MemDrawColors, SizeDrawColors)

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
		clipXMin = tools.Max(0, dstY) - dstY
		clipYMin = tools.Max(0, dstX) - dstX
		clipXMax = tools.Min(w, WIDTH-dstY)
		clipYMax = tools.Min(h, WIDTH-dstX)
	} else {
		clipXMin = tools.Max(0, dstX) - dstX
		clipYMin = tools.Max(0, dstY) - dstY
		clipXMax = tools.Min(w, WIDTH-dstX)
		clipYMax = tools.Min(h, WIDTH-dstY)
	}

	for y := clipYMin; y < clipYMax; y++ {
		for x := clipXMin; x < clipXMax; x++ {
			tx = dstX + tools.Ternary(rotate, y, x)
			ty = dstY + tools.Ternary(rotate, x, y)
			sx = srcX + tools.Ternary(flipX, w-x-1, x)
			sy = srcY + tools.Ternary(flipY, h-y-1, y)

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
				vpu.PointUnclippedFB(dc-1, tx, ty)
			}
		}
	}
}

func (vpu *VPU) GetColorByIndex(index int) byte {
	drawColors, _ := vpu.Memory().Read(MemDrawColors, SizeDrawColors)
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

func (vpu *VPU) LineFB(color byte, x1, y1, x2, y2 int32) {
	if color == 0 {
		return
	}

	color--

	if y1 > y2 {
		x1, x2 = x2, x1
		y1, y2 = y2, y1
	}

	var dx int32 = int32(math.Abs(float64(x2 - x1)))
	var sx int32 = tools.Ternary[int32](x1 < x2, 1, -1)
	var dy int32 = y2 - y1
	var err int32 = -dy
	if dx > dy {
		err = dx
	}
	err = err / 2
	var e2 int32

	for {
		vpu.PointUnclippedFB(color, x1, y1)
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 = err
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

func (vpu *VPU) HLineFB(color byte, startX, y, len int32) {
	// endX := startX + len

	if color == 0 {
		return
	}

	vpu.HLineUnclippedFB(color-1, startX, y, startX+len)

	// // Make sure it's from left to right
	// if startX > endX {
	// 	startX, endX = endX, startX
	// }

	// // Is the line even visible?
	// if y >= HEIGHT || y < 0 {
	// 	return
	// }
	// if endX < 0 || startX >= WIDTH {
	// 	return
	// }

	// // Stay in bound
	// if endX > 159 {
	// 	endX = 159
	// }
	// if endX < 0 {
	// 	endX = 0
	// }
	// if startX > 159 {
	// 	startX = 159
	// }
	// if startX < 0 {
	// 	startX = 0
	// }

	// for x := startX; x < endX; x++ {
	// 	vpu.PointFB(color, x, y)
	// }
}

func (vpu *VPU) HLineUnclippedFB(color byte, startX, y, endX int32) {
	if y >= 0 && y < HEIGHT {
		if startX < 0 {
			startX = 0
		}

		if endX > WIDTH {
			endX = WIDTH
		}

		if startX < endX {
			for x := startX; x < endX; x++ {
				vpu.PointUnclippedFB(color, x, y)
			}
		}
	}
}

func (vpu *VPU) VLineFB(color byte, x, y, len int32) {
	if y+len <= 0 || x < 0 || x >= WIDTH || color == 0 {
		return
	}

	startY := tools.Max(0, y)
	endY := tools.Min(HEIGHT, y+len)
	for yy := startY; yy < endY; yy++ {
		vpu.PointFB(color, x, yy)
	}
}

func (vpu *VPU) PointFB(color byte, x, y int32) {
	var (
		idx             int32 = (y*HEIGHT + x) >> 2
		currentValue, _       = vpu.Memory().ReadByte(uint32(idx) + MemFramebuffer)
		shift           int32 = (x & 0x3) << 1
		mask            int32 = 0x3 << shift
		value           uint8 = uint8((int32(color) << shift) | (int32(currentValue) & ^mask))
	)
	vpu.Memory().WriteByte(uint32(idx)+MemFramebuffer, value)
}

func (vpu *VPU) PointUnclippedFB(colorIndex byte, x, y int32) {
	if x >= 0 && x < WIDTH && y >= 0 && y < HEIGHT {
		vpu.PointFB(colorIndex, x, y)
	}
}

func (vpu *VPU) TextFB(txt string, x, y int32) {
	currX := x
	currY := y
	for _, letter := range []byte(txt) {
		// for _, l := range txt {
		// letter := l
		l := int32(letter)
		//log.Printf("letter: %v / %c / %v / %c", letter, letter, l, l)
		switch letter {
		case 0:
			return

		case '\n':
			currX = x
			currY += 8

		default:
			vpu.BlitFB(font, currX, currY, 8, 8, 0, (l-32)<<3, 8, false, false, false, false)
			currX += 8
		}
	}
}

func (vpu *VPU) OvalFB(x, y, width, height int32) {
	dc0 := vpu.GetColorByIndex(0)
	dc1 := vpu.GetColorByIndex(1)

	if dc1 == 0xf {
		return
	}

	strokeColor := (dc1 - 1) & 0x3
	fillColor := (dc0 - 1) & 0x3

	a := width - 1
	b := height - 1
	b1 := b % 2

	// north := y + int32(math.Floor(float64(height)/2))
	north := y + height/2
	west := x
	east := x + width - 1
	south := north - b1

	dx := 4 * (1 - a) * b * b
	dy := 4 * (b1 + 1) * a * a

	err := dx + dy + b1*a*a

	a *= 8 * a
	b1 = 8 * b * b

	// fmt.Print("OvalFB: ", x, ", ", y, ", ", width, ", ", height, ", ", strokeColor, ", ", fillColor, ", ", north, ", ", west, ", ", east, ", ", south, ", ", dx, ", ", dy, ", ", err, ", ", a, ", ", b1, "\n")

	var err2 int32

	for {
		vpu.PointUnclippedFB(strokeColor, east, north)
		vpu.PointUnclippedFB(strokeColor, west, north)
		vpu.PointUnclippedFB(strokeColor, west, south)
		vpu.PointUnclippedFB(strokeColor, east, south)

		start := west + 1
		len := east - start

		if dc0 != 0 && len > 0 {
			vpu.HLineUnclippedFB(fillColor, start, north, east)
			vpu.HLineUnclippedFB(fillColor, start, south, east)
		}
		err2 = 2 * err
		if err2 <= dy {
			// Move vertical scan
			north += 1
			south -= 1
			dy += a
			err += dy
		}
		if err2 >= dx || 2*err > dy {
			// Move horizontal scan
			west += 1
			east -= 1
			dx += b1
			err += dx
		}

		if west > east {
			break
		}
	}

	// Make sure north and south have moved the entire way so top/bottom aren't missing
	for north-south < height {
		vpu.PointUnclippedFB(strokeColor, west-1, north) /*   II. Quadrant    */
		vpu.PointUnclippedFB(strokeColor, east+1, north) /*   I. Quadrant     */
		north += 1
		vpu.PointUnclippedFB(strokeColor, west-1, south) /*   III. Quadrant   */
		vpu.PointUnclippedFB(strokeColor, east+1, south) /*   IV. Quadrant    */
		south -= 1
	}
}

func (vpu *VPU) RectFB(x, y, width, height int32) {
	startX := tools.Max(0, x)
	startY := tools.Max(0, y)
	endXUnclamp := x + width
	endYUnclamp := y + height
	endX := tools.Min(WIDTH, endXUnclamp)
	endY := tools.Min(HEIGHT, endYUnclamp)

	dc0 := vpu.GetColorByIndex(0)
	dc1 := vpu.GetColorByIndex(1)

	if dc0 != 0 {
		dc0 = (dc0 - 1) & 0x3
		for yy := startY; yy < endY; yy++ {
			for xx := startX; xx < endX; xx++ {
				vpu.PointFB(dc0, xx, yy)
			}
		}
	}

	if dc1 != 0 {
		dc1 = (dc1 - 1) & 0x3

		// Left edge
		if x >= 0 && x < WIDTH {
			for yy := startY; yy < endY; yy++ {
				vpu.PointFB(dc1, x, yy)
			}
		}

		// Right edge
		if endXUnclamp > 0 && endXUnclamp <= WIDTH {
			for yy := startY; yy < endY; yy++ {
				vpu.PointFB(dc1, endXUnclamp-1, yy)
			}
		}

		// Top edge
		if y >= 0 && y < HEIGHT {
			for xx := startX; xx < endX; xx++ {
				vpu.PointFB(dc1, xx, y)
			}
		}

		// Bottom edge
		if endYUnclamp > 0 && endYUnclamp <= HEIGHT {
			for xx := startX; xx < endX; xx++ {
				vpu.PointFB(dc1, xx, endYUnclamp-1)
			}
		}
	}
}
