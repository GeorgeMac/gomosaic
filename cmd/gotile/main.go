package main

import (
	"flag"
	"image"
	"image/draw"
	"image/png"
	"log"
	"os"

	"github.com/GeorgeMac/gomosaic/mosaic"

	_ "image/gif"
	_ "image/jpeg"
)

func main() {
	var size int
	flag.IntVar(&size, "s", 50, "Tile Output Size")
	flag.Parse()
	src, dst := flag.Arg(0), flag.Arg(1)

	srcfi, err := os.Open(src)
	if err != nil {
		log.Fatal(err)
	}
	defer srcfi.Close()

	dstfi, err := os.Create(dst)
	if err != nil {
		log.Fatal(err)
	}
	defer dstfi.Close()

	srcim, _, err := image.Decode(srcfi)
	if err != nil {
		log.Fatal(err)
	}

	bounds := srcim.Bounds()
	rect := image.Rect(0, 0, bounds.Dx(), bounds.Dx())
	copyim := image.NewRGBA(rect)
	draw.Draw(copyim, rect, srcim, image.ZP, draw.Src)

	dstim, err := mosaic.Resize(copyim, size, size)
	if err != nil {
		log.Fatal(err)
	}

	if err := png.Encode(dstfi, dstim); err != nil {
		log.Fatal(err)
	}
}
