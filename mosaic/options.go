package mosaic

import "github.com/GeorgeMac/gomosaic/mosaic/palette"

type option func(d *Converter)

func WithWidth(w int) option {
	return func(d *Converter) {
		d.width = w
	}
}

func WithHeight(h int) option {
	return func(d *Converter) {
		d.height = h
	}
}

func WithSize(s int) option {
	return func(d *Converter) {
		d.size = s
	}
}

func WithAlpha(a uint8) option {
	return func(d *Converter) {
		d.alpha = a
	}
}

func WithPaletteGenerator(g palette.Generator) option {
	return func(d *Converter) {
		d.generator = g
	}
}
