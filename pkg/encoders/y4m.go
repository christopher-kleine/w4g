package encoders

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"os"
	"path/filepath"
	"time"
)

type Y4M struct {
	running bool
	file    *os.File
}

func NewY4M() Encoder {
	return &Y4M{
		running: false,
	}
}

func (y4m *Y4M) Encode(img image.Image) {
	//log.Println("Encoding")
	y4m.file.Write([]byte("FRAME\n"))
	YSlice := []byte{}
	CbSlice := []byte{}
	CrSlice := []byte{}
	for y := 0; y < 160; y++ {
		for x := 0; x < 160; x++ {
			c := img.At(x, y)
			r, g, b, _ := c.RGBA()
			Y, Cb, Cr := color.RGBToYCbCr(byte(r), byte(g), byte(b))

			YSlice = append(YSlice, Y)
			CbSlice = append(CbSlice, Cb)
			CrSlice = append(CrSlice, Cr)
		}
	}
	y4m.file.Write(YSlice)
	y4m.file.Write(CbSlice)
	y4m.file.Write(CrSlice)
}

func (y4m *Y4M) IsRunning() bool {
	return y4m.running
}

func (y4m *Y4M) Start(name string) {
	udir, err := os.UserHomeDir()
	if err != nil {
		log.Println(err)
		return
	}

	fname := fmt.Sprintf("%s_%v.y4m", name, time.Now().Format("2006-01-02_15-04-05"))
	fname = filepath.Join(udir, fname)
	f, err := os.Create(fname)
	if err != nil {
		log.Println(err)
		return
	}

	f.Write([]byte("YUV4MPEG2 W160 H160 F60:1 Ip A1:1 C444\n"))

	y4m.file = f
	y4m.running = true
}

func (y4m *Y4M) Stop() {
	y4m.file.Close()
	y4m.running = false
}
