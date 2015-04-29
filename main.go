package main

import (
	"flag"
	"image/png"
	"log"
	"os"

	"github.com/GeorgeMac/gomosaic/mosiac"
)

func main() {
	var width, height int
	flag.IntVar(&width, "w", 50, "Width in number of tiles")
	flag.IntVar(&height, "h", 50, "Height in number of tiles")
	flag.Parse()

	path := flag.Args()[0]

	fi, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	im, err := mosaic.NewDecoder(fi, width, height).Decode()
	if err != nil {
		log.Fatal(err)
	}

	dst, err := os.Create("output.png")
	if err != nil {
		log.Fatal(err)
	}
	defer dst.Close()

	if err := png.Encode(dst, im); err != nil {
		log.Fatal(err)
	}
}
