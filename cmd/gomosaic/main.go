package main

import (
	"flag"
	"image/png"
	"io"
	"log"
	"os"

	"github.com/GeorgeMac/gomosaic/mosiac"
)

func main() {
	var width, height int
	var outp string
	flag.IntVar(&width, "w", 50, "Width in number of tiles")
	flag.IntVar(&height, "h", 50, "Height in number of tiles")
	flag.StringVar(&outp, "o", "", "Destination path to write file to (otherwise STDOUT)")
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
