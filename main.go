package main

import (
	"image/png"
	"log"
	"os"

	"github.com/GeorgeMac/gomosaic/mosiac"
)

func main() {
	path := os.Args[1]

	fi, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	im, err := mosaic.NewDecoder(fi).Decode()
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
