package runtime

import (
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/exp/constraints"
)

const (
	WIDTH  = 160
	HEIGHT = 160
)

func min[T constraints.Integer](a, b T) T {
	if a <= b {
		return a
	}

	return b
}

func max[T constraints.Integer](a, b T) T {
	if a >= b {
		return a
	}

	return b
}

func ternary[T any](eq bool, a, b T) T {
	if eq {
		return a
	}

	return b
}

func decompressBuffer(buffer []byte, bpp2 bool) []byte {
	result := []byte{}

	if bpp2 {
		for _, pixel := range buffer {
			pixels := []byte{
				pixel & 3,
				(pixel >> 2) & 3,
				(pixel >> 4) & 3,
				(pixel >> 6) & 3,
			}
			result = append(result, pixels...)
		}
	} else {
		for _, pixel := range buffer {
			pixels := []byte{
				pixel & 1,
				(pixel >> 1) & 1,
				(pixel >> 2) & 1,
				(pixel >> 3) & 1,
				(pixel >> 4) & 1,
				(pixel >> 5) & 1,
				(pixel >> 6) & 1,
				(pixel >> 7) & 1,
			}
			result = append(result, pixels...)
		}
	}

	return result
}

func compressBuffer(buffer []byte, bpp2 bool) []byte {
	result := []byte{}

	if bpp2 {
		for index := 0; index < len(buffer); index = index + 4 {
			p0 := buffer[index]
			p1 := buffer[index+1] << 2
			p2 := buffer[index+2] << 4
			p3 := buffer[index+3] << 6
			pixel := p0 | p1 | p2 | p3
			result = append(result, pixel)
		}
	} else {
		for index := 0; index < len(buffer); index = index + 8 {
			p0 := buffer[index]
			p1 := buffer[index+1] << 1
			p2 := buffer[index+2] << 2
			p3 := buffer[index+3] << 3
			p4 := buffer[index+4] << 4
			p5 := buffer[index+5] << 5
			p6 := buffer[index+6] << 6
			p7 := buffer[index+7] << 7
			pixel := p0 | p1 | p2 | p3 | p4 | p5 | p6 | p7
			result = append(result, pixel)
		}
	}

	return result
}

// This file implements direct access to the framebuffer.
// Other Drawing functions may use them.

func (rt *Runtime) ClearFB() {
	for pos := MemFramebuffer; pos < MemFramebuffer+SizeFramebuffer; pos++ {
		rt.cart.Memory().WriteByte(pos, 0)
	}
}

func (rt *Runtime) RenderFB(screen *ebiten.Image) {
	palette, _ := rt.cart.Memory().Read(MemPalette, SizePalette)
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

	framebuffer, _ := rt.cart.Memory().Read(MemFramebuffer, SizeFramebuffer)
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

func (rt *Runtime) BlitFB(sprite []byte, dstX, dstY, w, h, srcX, srcY, stride int32, bpp2, flipX, flipY, rotate bool) {
	drawColors, _ := rt.cart.Memory().Read(MemDrawColors, SizeDrawColors)

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
		clipXMin = max(0, dstY) - dstY
		clipYMin = max(0, dstX) - dstX
		clipXMax = min(w, WIDTH-dstY)
		clipYMax = min(h, WIDTH-dstX)
	} else {
		clipXMin = max(0, dstX) - dstX
		clipYMin = max(0, dstY) - dstY
		clipXMax = min(w, WIDTH-dstX)
		clipYMax = min(h, WIDTH-dstY)
	}

	for y := clipYMin; y < clipYMax; y++ {
		for x := clipXMin; x < clipXMax; x++ {
			tx = dstX + ternary(rotate, y, x)
			ty = dstY + ternary(rotate, x, y)
			sx = srcX + ternary(flipX, w-x-1, x)
			sy = srcY + ternary(flipY, h-y-1, y)

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
				rt.PointUnclippedFB(dc-1, tx, ty)
			}
		}
	}
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

func (rt *Runtime) LineFB(color byte, x1, y1, x2, y2 int32) {
	if color == 0 {
		return
	}

	if y1 > y2 {
		x1, x2 = x2, x1
		y1, y2 = y2, y1
	}

	var dx int32 = int32(math.Abs(float64(x2 - x1)))
	var sx int32 = ternary[int32](x1 < x2, 1, -1)
	var dy int32 = y2 - y1
	var err int32 = -dy
	if dx > dy {
		err = dx
	}
	err = err / 2
	var e2 int32

	for {
		rt.PointUnclippedFB(color, x1, y1)
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

func (rt *Runtime) HLineFB(color byte, startX, y, len int32) {
	endX := startX + len

	// Make sure it's from left to right
	if startX > endX {
		startX, endX = endX, startX
	}

	// Is the line even visible?
	if y >= HEIGHT || y < 0 {
		return
	}
	if endX < 0 || startX >= WIDTH {
		return
	}

	// Stay in bound
	if endX > 159 {
		endX = 159
	}
	if endX < 0 {
		endX = 0
	}
	if startX > 159 {
		startX = 159
	}
	if startX < 0 {
		startX = 0
	}

	for x := startX; x < endX; x++ {
		rt.PointFB(color, x, y)
	}
}

func (rt *Runtime) HLineUnclippedFB(color byte, startX, y, endX int32) {
	if y >= 0 && y < HEIGHT {
		if startX < 0 {
			startX = 0
		}

		if endX > WIDTH {
			endX = WIDTH
		}

		if startX < endX {
			for x := startX; x < endX; x++ {
				rt.PointUnclippedFB(color, x, y)
			}
		}
	}
}

func (rt *Runtime) VLineFB(color byte, x, y, len int32) {
	if y+len <= 0 || x < 0 || x >= WIDTH || color == 0 {
		return
	}

	startY := max(0, y)
	endY := min(HEIGHT, y+len)
	for yy := startY; yy < endY; yy++ {
		rt.PointFB(color, x, yy)
	}
}

func (rt *Runtime) PointFB(color byte, x, y int32) {
	var (
		idx             int32 = (y*HEIGHT + x) >> 2
		currentValue, _       = rt.cart.Memory().ReadByte(uint32(idx) + MemFramebuffer)
		shift           int32 = (x & 0x3) << 1
		mask            int32 = 0x3 << shift
		value           uint8 = uint8((int32(color) << shift) | (int32(currentValue) & ^mask))
	)
	rt.cart.Memory().WriteByte(uint32(idx)+MemFramebuffer, value)
}

func (rt *Runtime) PointUnclippedFB(colorIndex byte, x, y int32) {
	if x >= 0 && x < WIDTH && y >= 0 && y < HEIGHT {
		rt.PointFB(colorIndex, x, y)
	}
}

func (rt *Runtime) TextFB(txt string, x, y int32) {
	currX := x
	currY := y
	for _, letter := range []byte(txt) {
		// for _, l := range txt {
		// letter := l
		l := int32(letter)
		log.Printf("letter: %v / %c / %v / %c", letter, letter, l, l)
		switch letter {
		case 0:
			return

		case '\n':
			currX = x
			currY += 8

		default:
			rt.BlitFB(font, currX, currY, 8, 8, 0, (l-32)<<3, 8, false, false, false, false)
			currX += 8
		}
	}
}

func (rt *Runtime) OvalFB(x, y, width, height int32) {
	dc0 := rt.GetColorByIndex(0)
	dc1 := rt.GetColorByIndex(1)

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
		rt.PointUnclippedFB(strokeColor, east, north)
		rt.PointUnclippedFB(strokeColor, west, north)
		rt.PointUnclippedFB(strokeColor, west, south)
		rt.PointUnclippedFB(strokeColor, east, south)

		start := west + 1
		len := east - start

		if dc0 != 0 && len > 0 {
			rt.HLineUnclippedFB(fillColor, start, north, east)
			rt.HLineUnclippedFB(fillColor, start, south, east)
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
		rt.PointUnclippedFB(strokeColor, west-1, north) /*   II. Quadrant    */
		rt.PointUnclippedFB(strokeColor, east+1, north) /*   I. Quadrant     */
		north += 1
		rt.PointUnclippedFB(strokeColor, west-1, south) /*   III. Quadrant   */
		rt.PointUnclippedFB(strokeColor, east+1, south) /*   IV. Quadrant    */
		south -= 1
	}
}

func (rt *Runtime) RectFB(x, y, width, height int32) {
	startX := max(0, x)
	startY := max(0, y)
	endXUnclamp := x + width
	endYUnclamp := y + height
	endX := min(WIDTH, endXUnclamp)
	endY := min(HEIGHT, endYUnclamp)

	dc0 := rt.GetColorByIndex(0)
	dc1 := rt.GetColorByIndex(1)

	if dc0 != 0 {
		dc0 = (dc0 - 1) & 0x3
		for yy := startY; yy < endY; yy++ {
			for xx := startX; xx < endX; xx++ {
				rt.PointFB(dc0, xx, yy)
			}
		}
	}

	if dc1 != 0 {
		dc1 = (dc1 - 1) & 0x3

		// Left edge
		if x >= 0 && x < WIDTH {
			for yy := startY; yy < endY; yy++ {
				rt.PointFB(dc1, x, yy)
			}
		}

		// Right edge
		if endXUnclamp > 0 && endXUnclamp <= WIDTH {
			for yy := startY; yy < endY; yy++ {
				rt.PointFB(dc1, endXUnclamp-1, yy)
			}
		}

		// Top edge
		if y >= 0 && y < HEIGHT {
			for xx := startX; xx < endX; xx++ {
				rt.PointFB(dc1, xx, y)
			}
		}

		// Bottom edge
		if endYUnclamp > 0 && endYUnclamp <= HEIGHT {
			for xx := startX; xx < endX; xx++ {
				rt.PointFB(dc1, xx, endYUnclamp-1)
			}
		}
	}
}

func (rt *Runtime) Tone(frequency, duration, volume, flags int32) {
	//log.Printf("tone: %d / %d / %d / %d", frequency, duration, volume, flags)
}
