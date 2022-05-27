package encoders

import (
	"image"
)

type Encoder interface {
	Encode(img image.Image)
}
