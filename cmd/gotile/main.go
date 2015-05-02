package main

import (
	"image"
	"image/png"
	"log"
	"os"

	"github.com/GeorgeMac/gomosaic/mosaic/tile"

	_ "image/gif"
	_ "image/jpeg"
)

func main() {
	src, dst := os.Args[1], os.Args[2]

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

	dstim, err := tile.Tile(srcim, 100, 100)
	if err != nil {
		log.Fatal(err)
	}

	if err := png.Encode(dstfi, dstim); err != nil {
		log.Fatal(err)
	}
}
