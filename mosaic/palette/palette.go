package palette

import (
	"encoding/binary"
	"image"
	"image/color"
)

type Palette interface {
	Convert(ColorKey) Tile
}

type Tile interface {
	image.Image
	color.Color
	ColorAt(image.Rectangle) color.Color
}

type Generator interface {
	Palette(term string, size int) (Palette, error)
}

type GeneratorFunc func(term string, size int) (Palette, error)

func (p GeneratorFunc) Palette(term string, size int) (Palette, error) {
	return p(term, size)
}

type ColorKey [4]uint32

func NewColorKey(c color.Color) ColorKey {
	r, g, b, a := c.RGBA()
	return [4]uint32{r, g, b, a}
}

func ColorKeyFromBytes(b []byte) ColorKey {
	key := ColorKey{}
	for i := 0; i < len(key); i++ {
		n := i * 8
		v, _ := binary.Uvarint(b[n : n+8])
		key[i] = uint32(v)
	}
	return key
}

func (k ColorKey) Bytes() []byte {
	buf := make([]byte, 32)
	for i, v := range k {
		n := i * 8
		binary.PutUvarint(buf[n:n+8], uint64(v))
	}
	return buf
}

func (k ColorKey) Color() color.Color {
	return color.RGBA{uint8(k[0]), uint8(k[1]), uint8(k[2]), uint8(k[3])}
}
