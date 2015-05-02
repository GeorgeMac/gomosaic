package mosaic

import (
	"image"
	"image/color"
	"image/color/palette"
)

var UniformPalette *TilePalette

func init() {
	tiles := make([]Tile, 0)
	for _, c := range palette.WebSafe {
		tiles = append(tiles, &UniformTile{
			Uniform: image.NewUniform(c),
		})
	}
	UniformPalette = NewTilePalette(tiles)
}

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
}

func NewTilePalette(tiles []Tile) *TilePalette {
	t := &TilePalette{
		lookup:  map[Key]Tile{},
		palette: color.Palette(make([]color.Color, 0)),
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
