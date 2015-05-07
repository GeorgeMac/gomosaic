package mosaic

import (
	"image"
	"image/color"
	"image/color/palette"
	"os"
	"path/filepath"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

type Key [4]uint32

func NewKey(c color.Color) Key {
	r, g, b, a := c.RGBA()
	return [4]uint32{r, g, b, a}
}

func (k Key) Color() color.Color {
	return color.RGBA{uint8(k[0]), uint8(k[1]), uint8(k[2]), uint8(k[3])}
}

type TilePalette struct {
	lookup  map[Key]Tile
	palette color.Palette
	Size    int
}

func NewUniformWebColorPalette(size int) *TilePalette {
	tiles := make([]Tile, 0)
	for _, c := range palette.WebSafe {
		tiles = append(tiles, &UniformTile{
			Uniform: image.NewUniform(c),
		})
	}
	return NewTilePalette(tiles, size)
}

func NewImageTilePalette(dir string, size int) (*TilePalette, error) {
	tile := make([]Tile, 0)
	walkfn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		for _, e := range []string{".gif", ".jpg", ".jpeg", ".png"} {
			if ext == e {
				fi, err := os.Open(path)
				if err != nil {
					return err
				}
				im, _, err := image.Decode(fi)
				if err != nil {
					return err
				}

				tile = append(tile, NewImageTile(im))
				return nil
			}
		}
		return nil
	}

	if err := filepath.Walk(dir, walkfn); err != nil {
		return nil, err
	}

	return NewTilePalette(tile, size), nil
}

func NewTilePalette(tiles []Tile, size int) *TilePalette {
	t := &TilePalette{
		lookup:  map[Key]Tile{},
		palette: color.Palette(make([]color.Color, 0)),
		Size:    size,
	}

	for _, tile := range tiles {
		t.palette = append(t.palette, tile)
		t.lookup[NewKey(tile)] = tile
	}
	return t
}

func (t *TilePalette) Convert(c color.Color) Tile {
	// normalise the color in to palette colors
	c = t.palette.Convert(c)

	// lookup using key from color palette
	if tile, ok := t.lookup[NewKey(c)]; ok {
		return tile
	}
	return nil
}
