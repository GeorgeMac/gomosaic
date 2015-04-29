package mosaic

import (
	"image"
	"io"
	"log"
	"math"
	"sync"

	"image/color"
	"image/color/palette"
	"image/draw"
	_ "image/png"
)

// type convert palette.WebSafe in to color.Palette type locally
var WebSafe color.Palette = color.Palette(palette.WebSafe)

type Decoder struct {
	io.Reader
	Width, Height int
}

func NewDecoder(r io.Reader, width, height int) *Decoder {
	return &Decoder{
		Reader: r,
		Width:  width,
		Height: height,
	}
}

func (d *Decoder) Decode() (image.Image, error) {
	im, _, err := image.Decode(d)
	if err != nil {
		return nil, err
	}

	// tiles to process channel
	proc := make(chan image.Rectangle, 100)
	// tiles to compose
	comp := make(chan tile)

	// initial image bounds
	bounds := im.Bounds()
	// begin tiling routing
	go func() {
		x, y := bounds.Min.X, bounds.Min.Y
		dx := int(math.Ceil(float64(bounds.Max.X / d.Width)))
		dy := int(math.Ceil(float64(bounds.Max.Y / d.Height)))

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

				log.Println(x1, y1, x2, y2)
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
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				for rec := range proc {
					comp <- tile{
						Rect:  rec,
						Image: uniformForRectangleInImage(im, rec),
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

// tile represents a window (tile) from a source image
type tile struct {
	Rect  image.Rectangle
	Image image.Image
}

func uniformForRectangleInImage(m image.Image, r image.Rectangle) image.Image {
	bins := map[[4]uint32]int{}

	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			r, g, b, a := WebSafe.Convert(m.At(x, y)).RGBA()
			key := [4]uint32{r, g, b, a}
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
		if c == nil {
			c = RGBA(k)
			continue
		}
		if v > max {
			c = RGBA(k)
			max = v
		}
	}

	if c == nil {
		return image.NewUniform(color.Black)
	}
	return image.NewUniform(c)
}

func RGBA(v [4]uint32) color.Color {
	return color.RGBA{uint8(v[0]), uint8(v[1]), uint8(v[2]), uint8(v[3])}
}
