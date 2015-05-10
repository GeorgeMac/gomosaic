package main

import (
	"flag"
	"image"
	"image/draw"
	"image/png"
	"io"
	"log"
	"os"

	"github.com/GeorgeMac/gomosaic/mosaic"
	"github.com/GeorgeMac/gomosaic/mosaic/tile"
)

func main() {
	var width, height, alpha, t int
	var outp, dirp string
	flag.IntVar(&width, "w", 50, "Width in number of tiles")
	flag.IntVar(&height, "h", 50, "Height in number of tiles")
	flag.IntVar(&t, "t", 100, "Tile size in t/t px")
	flag.IntVar(&alpha, "a", 255, "Alpha for masking tiles (0 to 255)")
	flag.StringVar(&outp, "o", "", "Destination path to write file to (otherwise STDOUT)")
	flag.StringVar(&dirp, "d", "", "Location of images to use as tiles")
	flag.Parse()

	path := flag.Args()[0]

	fi, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	im, _, err := image.Decode(fi)
	if err != nil {
		log.Fatal("Decoding Error: ", err)
	}

	bounds := im.Bounds()
	var w, x, y int
	x, y = bounds.Dx(), bounds.Dy()
	w = x
	if y < x {
		w = y
	}

	rect := image.Rect(0, 0, x, y)
	rgbaim := image.NewRGBA(rect)
	draw.Draw(rgbaim, rect, im, image.ZP, draw.Src)

	im, err = tile.Tile(rgbaim, w, w)
	if err != nil {
		log.Fatal("Tiling Error: ", err)
	}

	palette := mosaic.CommonPaletteGenerator(mosaic.NewUniformWebColorPalette)
	if dirp != "" {
		palette = mosaic.PaletteGeneratorFunc(func(size int) (*mosaic.TilePalette, error) {
			return mosaic.NewImageTilePalette(dirp, size)
		})
	}
	decoder := mosaic.NewConverter(im,
		mosaic.WithWidth(width),
		mosaic.WithHeight(height),
		mosaic.WithSize(t),
		mosaic.WithPaletteGenerator(palette),
		mosaic.WithAlpha(uint8(alpha)))
	im, err = decoder.Decode()
	if err != nil {
		log.Fatal(err)
	}

	var out io.Writer = os.Stdout

	if outp != "" {
		var err error
		fi, err := os.Create(outp)
		if err != nil {
			log.Fatal(err)
		}
		defer fi.Close()
		out = fi
	}

	if err := png.Encode(out, im); err != nil {
		log.Fatal(err)
	}
}
