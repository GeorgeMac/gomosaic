package palette

import (
	"image/color"
	"math"
	"testing"
)

var errmsg string = "Expected %v, Got %v/n"

func TestNewColorKey(t *testing.T) {
	key := NewColorKey(color.Black)

	expected := [4]uint32{0, 0, 0, math.MaxUint16}
	if key != expected {
		t.Errorf(errmsg, expected, key)
	}
}

func TestColorKey_Color(t *testing.T) {
	red := color.RGBA{
		R: 255,
		G: 0,
		B: 0,
		A: 255,
	}
	key := NewColorKey(red)

	if key.Color() != red {
		t.Errorf(errmsg, red, key)
	}
}

func TestColorKey_Bytes(t *testing.T) {
	colors := []color.Color{
		color.RGBA{0, 0, 0, 255},
		color.RGBA{255, 255, 255, 255},
		color.RGBA{R: 255, B: 0, G: 0, A: 255},
		color.RGBA{R: 0, B: 255, G: 0, A: 255},
		color.RGBA{R: 0, B: 0, G: 255, A: 255},
		color.RGBA{R: 255, B: 255, G: 255, A: 125},
	}
	for _, col := range colors {
		key := NewColorKey(col)

		buf := key.Bytes()
		if len(buf) != 32 {
			t.Errorf(errmsg, 32, len(buf))
		}

		newKey := ColorKeyFromBytes(buf)
		if newKey != key {
			t.Errorf(errmsg, key, newKey)
		}

		newCol := newKey.Color()
		if newCol != col {
			t.Errorf(errmsg, col, newCol)
		}
	}
}
