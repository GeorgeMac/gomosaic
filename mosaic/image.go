package mosaic

import (
	"image"
	"image/color"
	"image/color/palette"
)

// type convert palette.WebSafe in to color.Palette type locally
var WebSafe color.Palette = color.Palette(palette.WebSafe)

type Tile interface {
	image.Image
	color.Color
	ColorAt(image.Rectangle) color.Color
}

type ImageTile struct {
	image.Image
	c color.Color
}

func NewImageTile(i image.Image) *ImageTile {
	tile := &ImageTile{
		Image: i,
	}
	bounds := i.Bounds()
	tile.c = tile.ColorAt(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	return tile
}

func (t *ImageTile) RGBA() (r, g, b, a uint32) {
	return t.c.RGBA()
}

func (t *ImageTile) ColorAt(r image.Rectangle) color.Color {
	bins := map[Key]int{}

	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			key := NewKey(WebSafe.Convert(t.Image.At(x, y)))
			if v, ok := bins[key]; ok {
				bins[key] = v + 1
				continue
			}
			bins[key] = 1
		}
	}

	var (
		c   color.Color
		max int
	)

	for k, v := range bins {
		if c == nil || v > max {
			c = k.Color()
			max = v
		}
	}

	return c
}

// Uniform tile implements Tile interface
// It is a tile which is just an infinitely bound
// single color. It wraps an image.Uniform struct type.
type UniformTile struct {
	*image.Uniform
}

func (u *UniformTile) ColorAt(_ image.Rectangle) color.Color {
	return u.C
}
