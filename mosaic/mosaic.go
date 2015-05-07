package mosaic

import (
	"fmt"
	"image"
	"log"
	"math"
	"sync"

	"github.com/bamiaux/rez"

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
	im            image.Image
	palette       *TilePalette
	width, height int
}

func NewDecoder(im image.Image, opts ...option) *Decoder {
	d := &Decoder{
		im:      im,
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
	// tiles to process channel
	proc := make(chan image.Rectangle, 10)
	// tiles to compose
	comp := make(chan source)
	// resized image promise
	scaled := make(chan draw.Image)

	bounds := d.im.Bounds()
	// initial image bounds
	nx := d.width * d.palette.Size
	ny := d.height * d.palette.Size
	sx, sy := float64(nx)/float64(bounds.Dx()), float64(ny)/float64(bounds.Dy())

	fmt.Printf("[mosaic] Original [%d, %d] New [%d, %d] Scale [%d, %d]\n", bounds.Dx(), bounds.Dy(), nx, ny, sx, sy)

	log.Println("[mosaic] Begin resizing")
	go func() {
		dst := image.NewRGBA(image.Rect(0, 0, nx, ny))
		if err := rez.Convert(dst, d.im, rez.NewBilinearFilter()); err != nil {
			log.Fatal(err)
		}
		scaled <- dst
		close(scaled)
	}()

	imtile := NewImageTile(d.im)
	log.Println("[mosaic] Calculating tiles")
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

	log.Println("[mosaic] Fetching color information")
	// average calculation go routines
	var wg sync.WaitGroup
	go func() {
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				for rect := range proc {
					c := imtile.ColorAt(rect)
					min, max := rect.Min, rect.Max
					comp <- source{
						Image: d.palette.Convert(c),
						Rect: image.Rectangle{
							Min: image.Point{
								X: int(math.Floor(float64(min.X) * sx)),
								Y: int(math.Floor(float64(min.Y) * sy)),
							},
							Max: image.Point{
								X: int(math.Floor(float64(max.X) * sx)),
								Y: int(math.Floor(float64(max.Y) * sy)),
							},
						},
					}
				}
				wg.Done()
			}()
		}
		wg.Wait()
		close(comp)
	}()

	mask := image.NewUniform(color.Alpha{A: uint8(200)})
	// tile composition routine
	log.Println("[mosaic] Composing image")
	dst := <-scaled
	for tile := range comp {
		// draw tile in to destination
		draw.DrawMask(dst, tile.Rect, tile.Image, image.ZP, mask, image.ZP, draw.Over)
	}

	return dst, nil
}

// window contains an image to render + a target rectangle
// view to render it in to.
type source struct {
	Image image.Image
	Rect  image.Rectangle
}
