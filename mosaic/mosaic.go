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

func WithSize(s int) option {
	return func(d *Decoder) {
		d.size = s
	}
}

func WithAlpha(a uint8) option {
	return func(d *Decoder) {
		d.alpha = a
	}
}

func WithPalette(p func(int) (*TilePalette, error)) option {
	return func(d *Decoder) {
		d.palette = p
	}
}

type Decoder struct {
	im                  image.Image
	palette             PaletteGeneratorFunc
	width, height, size int
	alpha               uint8
}

func NewDecoder(im image.Image, opts ...option) *Decoder {
	d := &Decoder{
		im:      im,
		width:   100,
		height:  100,
		size:    100,
		alpha:   255,
		palette: CommonPaletteGenerator(NewUniformWebColorPalette),
	}

	for _, opt := range opts {
		opt(d)
	}

	return d
}

func (d *Decoder) Decode() (image.Image, error) {
	// tiles to process channel
	proc := make(chan image.Rectangle, 2)
	// tiles to compose
	comp := make(chan source)
	// resized image promise
	scaled := make(chan draw.Image)
	// error channel for failed palette generation
	errc := make(chan error)

	bounds := d.im.Bounds()
	// initial image bounds
	nx := d.width * d.size
	ny := d.height * d.size
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

	// Begin calculating tiles to sample/scale
	log.Println("[mosaic] Calculating tiles")
	go d.bounds(proc, bounds)

	// average calculation go routines
	log.Println("[mosaic] Fetching color information")
	go d.process(proc, comp, errc, sx, sy)

	// tile composition routine
	log.Println("[mosaic] Composing image")
	mask := image.NewUniform(color.Alpha{A: d.alpha})
	// get scaled source image
	dst := <-scaled
	if err := func() error {
		for {
			select {
			case tile, ok := <-comp:
				if !ok {
					return nil
				}
				draw.DrawMask(dst, tile.Rect, tile.Image, image.ZP, mask, image.ZP, draw.Over)
			case err := <-errc:
				if err != nil {
					return err
				}
			}
		}
	}(); err != nil {
		return nil, err
	}

	return dst, nil
}

func (d *Decoder) bounds(proc chan<- image.Rectangle, bounds image.Rectangle) {
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
}

func (d *Decoder) process(proc <-chan image.Rectangle, comp chan<- source, errc chan<- error, sx, sy float64) {
	var wg sync.WaitGroup
	imtile := NewImageTile(d.im)
	palette, err := d.palette(d.size)
	if err != nil {
		log.Println("[mosaic] Error creating palette")
		close(comp)
		errc <- err
		close(errc)
		return
	}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			for rect := range proc {
				c := imtile.ColorAt(rect)
				min, max := rect.Min, rect.Max
				comp <- source{
					Image: palette.Convert(c),
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
	close(errc)
}

// window contains an image to render + a target rectangle
// view to render it in to.
type source struct {
	Image image.Image
	Rect  image.Rectangle
}
