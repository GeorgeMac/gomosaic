package mosaic

import (
	"image"
	"image/color"
	plt "image/color/palette"

	"github.com/GeorgeMac/gomosaic/mosaic/palette"
)

// type convert palette.WebSafe in to color.Palette type locally
var WebSafe color.Palette = color.Palette(plt.WebSafe)

type ImageTile struct {
	image.Image
	Color color.Color
}

func NewImageTile(i image.Image) *ImageTile {
	tile := &ImageTile{
		Image: i,
	}

	if i, ok := i.(*image.Uniform); ok {
		tile.Color = i.C
		return tile
	}

	bounds := i.Bounds()
	tile.Color = tile.ColorAt(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	return tile
}

func (t *ImageTile) RGBA() (r, g, b, a uint32) {
	return t.Color.RGBA()
}

func (t *ImageTile) ColorAt(r image.Rectangle) color.Color {
	bins := map[palette.ColorKey]int{}

	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			key := palette.NewColorKey(WebSafe.Convert(t.Image.At(x, y)))
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
