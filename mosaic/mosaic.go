package mosaic

import (
	"image"
	"io"
	"math"
	"sync"

	"image/color"
	"image/color/palette"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

// type convert palette.WebSafe in to color.Palette type locally
var WebSafe color.Palette = color.Palette(palette.WebSafe)

type option func(d *Decoder)

func WithWidth(w int) option {
	return func(d *Decoder) {
		d.width = w
	}
}

func WithHeight(h int) option {
	return func(d *Decoder) {
		d.height = h
	}
}

func WithPalette(p *TilePalette) option {
	return func(d *Decoder) {
		d.palette = p
	}
}

type Decoder struct {
	io.Reader
	palette       *TilePalette
	width, height int
}

func NewDecoder(r io.Reader, opts ...option) *Decoder {
	d := &Decoder{
		Reader:  r,
		width:   100,
		height:  100,
		palette: UniformPalette,
	}

	for _, opt := range opts {
		opt(d)
	}

	return d
}

func (d *Decoder) Decode() (image.Image, error) {
	im, _, err := image.Decode(d)
	if err != nil {
		return nil, err
	}

	imtile := NewImageTile(im)

	// tiles to process channel
	proc := make(chan image.Rectangle, 100)
	// tiles to compose
	comp := make(chan window)

	// initial image bounds
	bounds := im.Bounds()
	// begin tiling routing
	go func() {
		x, y := bounds.Min.X, bounds.Min.Y
		dx := int(math.Ceil(float64(bounds.Max.X / d.width)))
		dy := int(math.Ceil(float64(bounds.Max.Y / d.height)))

		x1 := x
		x2 := x + dx
		for {
			// don't let x1 exceed max X
			if x2 > bounds.Max.X {
				x2 = bounds.Max.X
			}

			y1 := y
			y2 := y + dx
			for {
				// don't let y1 exceed max Y
				if y2 > bounds.Max.Y {
					y2 = bounds.Max.Y
				}

				// create rectangle view
				proc <- image.Rect(x1, y1, x2, y2)

				// break now because we have reached the max boundary
				if y1+dy >= bounds.Max.Y {
					break
				}
				// increase y by dy
				y1 = y2
				y2 += dy
			}
			// break now because we have reached the max boundary
			if x1+dx >= bounds.Max.X {
				break
			}
			// increase x by dx
			x1 = x2
			x2 += dx
		}
		close(proc)
	}()

	// average calculation go routines
	var wg sync.WaitGroup
	go func() {
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				for rec := range proc {
					c := imtile.ColorAt(rec)
					comp <- window{
						Rect:  rec,
						Image: d.palette.Convert(c),
					}
				}
				wg.Done()
			}()
		}
		wg.Wait()
		close(comp)
	}()

	dst := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	// tile composition routine
	for tile := range comp {
		// draw tile in to destination
		draw.Draw(dst, tile.Rect, tile.Image, tile.Rect.Min, draw.Src)
	}

	return dst, err
}

// window contains an image to render + a target rectangle
// view to render it in to.
type window struct {
	Rect  image.Rectangle
	Image image.Image
}
