package mosaic

import (
	"encoding/binary"
	"image"
	"image/color"
	"image/color/palette"
	"os"
	"path/filepath"
	"sync"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

type ColorKey [4]uint32

func NewColorKey(c color.Color) ColorKey {
	r, g, b, a := c.RGBA()
	return [4]uint32{r, g, b, a}
}

func ColorKeyFromBytes(b []byte) ColorKey {
	key := ColorKey{}
	for i := 0; i < len(key); i++ {
		n := i * 8
		v, _ := binary.Uvarint(b[n : n+8])
		key[i] = uint32(v)
	}
	return key
}

func (k ColorKey) Bytes() []byte {
	buf := make([]byte, 32)
	for i, v := range k {
		n := i * 8
		binary.PutUvarint(buf[n:n+8], uint64(v))
	}
	return buf
}

func (k ColorKey) Color() color.Color {
	return color.RGBA{uint8(k[0]), uint8(k[1]), uint8(k[2]), uint8(k[3])}
}

type PaletteGenerator interface {
	Palette(size int) (*TilePalette, error)
}

type PaletteGeneratorFunc func(size int) (*TilePalette, error)

func (p PaletteGeneratorFunc) Palette(size int) (*TilePalette, error) {
	return p(size)
}

func (p PaletteGeneratorFunc) Begin(size int) <-chan *TilePalette {
	ch := make(chan *TilePalette)
	go func() {
		pt, err := p(size)
		if err == nil {
			ch <- pt
		}
		close(ch)
	}()
	return ch
}

func NewUniformWebColorPalette(size int) (*TilePalette, error) {
	tiles := make([]Tile, 0)
	for _, c := range palette.WebSafe {
		tiles = append(tiles, &UniformTile{
			Uniform: image.NewUniform(c),
		})
	}
	return NewTilePalette(tiles, size), nil
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

type TilePalette struct {
	lookup  map[ColorKey][]Tile
	palette color.Palette
	mu      sync.Mutex
	Size    int
}

func NewTilePalette(tiles []Tile, size int) *TilePalette {
	t := &TilePalette{
		lookup:  map[ColorKey][]Tile{},
		palette: color.Palette(make([]color.Color, 0)),
		Size:    size,
	}

	for _, tile := range tiles {
		t.palette = append(t.palette, tile)
		key := NewColorKey(tile)
		if l, ok := t.lookup[key]; ok {
			t.lookup[key] = append(l, tile)
			continue
		}
		t.lookup[NewColorKey(tile)] = []Tile{tile}
	}
	return t
}

func (t *TilePalette) Convert(k ColorKey) Tile {
	t.mu.Lock()
	defer t.mu.Unlock()
	// normalise the color in to palette colors
	c := t.palette.Convert(k.Color())

	// lookup using key from color palette
	key := NewColorKey(c)
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
