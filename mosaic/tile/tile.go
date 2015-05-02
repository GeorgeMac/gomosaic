package tile

import (
	"image"
	"image/draw"
	"math"

	"github.com/bamiaux/rez"
)

func Tile(m image.Image, width, height int) (image.Image, error) {
	bounds := m.Bounds()
	x, y := bounds.Dx(), bounds.Dy()
	if x < width || y < height {
		return nil, ImageNotSuitable{}
	}

	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	if x != y {
		sp := image.ZP
		var w int
		if x < y {
			w = x
			sp = image.Pt(0, int(math.Floor(float64((y-x)/2))))
		} else {
			w = y
			sp = image.Pt(int(math.Floor(float64((x-y)/2))), 0)
		}
		n := image.NewRGBA(image.Rect(0, 0, w, w))
		draw.Draw(n, n.Bounds(), m, sp, draw.Src)
		m = n
	}
	return dst, rez.Convert(dst, m, rez.NewBilinearFilter())
}

// errors

type ImageNotSuitable struct{}

func (i ImageNotSuitable) Error() string { return "image not suitable for tiling" }
