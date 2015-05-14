package mosaic

import (
	"image"
	"image/color"
	plt "image/color/palette"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"sync"

	"github.com/GeorgeMac/gomosaic/mosaic/palette"
)

func NewUniformWebColorPalette(_ string, size int) (palette.Palette, error) {
	tiles := make([]palette.Tile, 0)
	for _, c := range plt.WebSafe {
		tiles = append(tiles, &UniformTile{
			Uniform: image.NewUniform(c),
		})
	}
	return NewTilePalette(tiles, size), nil
}

func NewImageTilePalette(dir string, size int) (palette.Palette, error) {
	tile := make([]palette.Tile, 0)
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

type TilePalette struct {
	lookup  map[palette.ColorKey][]palette.Tile
	palette color.Palette
	mu      sync.Mutex
	Size    int
}

func NewTilePalette(tiles []palette.Tile, size int) *TilePalette {
	t := &TilePalette{
		lookup:  map[palette.ColorKey][]palette.Tile{},
		palette: color.Palette(make([]color.Color, 0)),
		Size:    size,
	}

	for _, tile := range tiles {
		t.palette = append(t.palette, tile)
		key := palette.NewColorKey(tile)
		if l, ok := t.lookup[key]; ok {
			t.lookup[key] = append(l, tile)
			continue
		}
		t.lookup[palette.NewColorKey(tile)] = []palette.Tile{tile}
	}
	return t
}

func (t *TilePalette) Convert(k palette.ColorKey) palette.Tile {
	t.mu.Lock()
	defer t.mu.Unlock()
	// normalise the color in to palette colors
	c := t.palette.Convert(k.Color())

	// lookup using key from color palette
	key := palette.NewColorKey(c)
	if tiles, ok := t.lookup[key]; ok {
		if len(tiles) == 1 {
			return tiles[0]
		}
		tile := tiles[0]
		t.lookup[key] = append(tiles[1:], tile)
		return tile
	}
	return nil
}
