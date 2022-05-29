package runtime

func (rt *Runtime) Blit(spr, x, y, w, h, f int32) {
	rt.BlitSub(spr, x, y, w, h, 0, 0, w, f)
}

func (rt *Runtime) BlitSub(spr, x, y, w, h, srcX, srcY, stride, f int32) {
	sprite, _ := rt.env.Memory().Read(rt.ctx, uint32(spr), uint32(w*h))

	bpp2 := f&1 == 1
	flipX := f&2 == 2
	flipY := f&4 == 4
	rotate := f&8 == 8

	rt.BlitFB(sprite, x, y, w, h, srcX, srcY, stride, bpp2, flipX, flipY, rotate)
}

func (rt *Runtime) Line(x1, y1, x2, y2 int32) {
	rt.LineFB(x1, y1, x2, y2)
}

func (rt *Runtime) HLine(x, y, l int32) {
	rt.HLineFB(x, y, l)
}

func (rt *Runtime) VLine(x, y, l int32) {
	rt.VLineFB(x, y, l)
}

func (rt *Runtime) Oval(x, y, w, h int32) {
}

func (rt *Runtime) Rect(x, y, w, h int32) {
}

func (rt *Runtime) Text(txt, x, y int32) {
	currX := x
	currY := y
	for _, letter := range rt.getString(txt) {
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

func (rt *Runtime) TextUTF8(textPtr, byteLength, x, y int32) {
	// const text = new Uint8Array(this.memory.buffer, textPtr, byteLength);
	// this.framebuffer.drawText(text, x, y);
}

func (rt *Runtime) getString(txt int32) string {
	letter, _ := rt.env.Memory().ReadByte(rt.ctx, uint32(txt))
	text := ""
	offset := 0
	for letter != 0 {
		text += string(letter)
		offset++
		letter, _ = rt.env.Memory().ReadByte(rt.ctx, uint32(txt)+uint32(offset))
	}

	return text
}
