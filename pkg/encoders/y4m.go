package encoders

import "image"

type Y4M struct{}

func NewY4M(filename string) Encoder {
	return &Y4M{}
}

func (yam *Y4M) Encode(img image.Image) {
}
