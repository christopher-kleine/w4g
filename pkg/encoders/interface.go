package encoders

import (
	"image"
)

type Encoder interface {
	Encode(img image.Image)
	IsRunning() bool
	Start(name string)
	Stop()
}
