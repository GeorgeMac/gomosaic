package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/GeorgeMac/gomosaic/mosaic"
)

func main() {
	var width, height int
	var outp, dirp string
	flag.IntVar(&width, "w", 50, "Width in number of tiles")
	flag.IntVar(&height, "h", 50, "Height in number of tiles")
	flag.StringVar(&outp, "o", "", "Destination path to write file to (otherwise STDOUT)")
	flag.StringVar(&dirp, "d", "", "Location of images to use as tiles")
	flag.Parse()

	path := flag.Args()[0]

	fi, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	palette := mosaic.UniformPalette
	if dirp != "" {
		tiles := []mosaic.Tile{}
		if err := filepath.Walk(dirp, func(path string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if fi.IsDir() {
				return nil
			}

			for _, ext := range []string{"png", "jpg", "jpeg", "gif"} {
				if strings.HasSuffix(path, ext) {
					fi, err := os.Open(path)
					if err != nil {
						return err
					}

					m, _, err := image.Decode(fi)
					if err != nil {
						return err
					}

					tile := mosaic.NewImageTile(m)
					fmt.Println("Tile", tile)
					tiles = append(tiles, tile)
					return nil
				}
			}
			return nil
		}); err != nil {
			log.Fatal(err)
		}

		palette = mosaic.NewTilePalette(tiles)
	}

	decoder := mosaic.NewDecoder(fi, mosaic.WithWidth(width), mosaic.WithHeight(height), mosaic.WithPalette(palette))
	im, err := decoder.Decode()
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
