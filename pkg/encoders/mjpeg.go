package encoders

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/christopher-kleine/mjpeg"
)

type MJPEG struct {
	running bool
	file    mjpeg.AviWriter
	Quality int
}

func NewMJPEG(quality int) Encoder {
	return &MJPEG{
		running: false,
		Quality: quality,
	}
}

func (encoder *MJPEG) Encode(img image.Image) {
	buf := &bytes.Buffer{}
	err := jpeg.Encode(buf, img, &jpeg.Options{
		Quality: encoder.Quality,
	})

	if err == nil {
		encoder.file.AddFrame(buf.Bytes())
	}
}

func (encoder *MJPEG) IsRunning() bool {
	return encoder.running
}

func (encoder *MJPEG) Start(name string) {
	udir, err := os.UserHomeDir()
	if err != nil {
		log.Println(err)
		return
	}

	fname := fmt.Sprintf("%s_%v.avi", name, time.Now().Format("2006-01-02_15-04-05"))
	fname = filepath.Join(udir, fname)
	f, err := mjpeg.New(fname, 160, 160, 60)
	if err != nil {
		log.Println(err)
		return
	}

	encoder.file = f
	encoder.running = true
}

func (encoder *MJPEG) Stop() {
	encoder.file.Close()
	encoder.running = false
}
