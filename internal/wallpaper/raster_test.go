package wallpaper

import (
	"image"
	"image/color"
	"testing"
)

// gradient is 2 wide and 4 tall, each pixel a distinct grey, so a cell's two
// halves can be told apart.
func gradient() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 2, 4))
	for y := range 4 {
		for x := range 2 {
			v := uint8(10*y + x)
			img.SetRGBA(x, y, color.RGBA{R: v, G: v, B: v, A: 255})
		}
	}
	return img
}

func grey(v uint8) color.RGBA { return color.RGBA{R: v, G: v, B: v, A: 255} }

func TestRasterizeOneToOne(t *testing.T) {
	img := gradient()
	l := Layout{Src: img.Bounds(), Dst: Rect{W: 2, H: 2}}
	cells := Rasterize(img, l, 2, 2)
	want := []Cell{
		{Top: grey(0), Bottom: grey(10), Set: true},
		{Top: grey(1), Bottom: grey(11), Set: true},
		{Top: grey(20), Bottom: grey(30), Set: true},
		{Top: grey(21), Bottom: grey(31), Set: true},
	}
	for i := range want {
		if cells[i] != want[i] {
			t.Errorf("cell %d: got %+v, want %+v", i, cells[i], want[i])
		}
	}
}

func TestRasterizeAveragesWhenShrinking(t *testing.T) {
	img := gradient()
	l := Layout{Src: img.Bounds(), Dst: Rect{W: 1, H: 1}}
	cells := Rasterize(img, l, 1, 1)
	want := Cell{Top: grey(5), Bottom: grey(25), Set: true}
	if cells[0] != want {
		t.Errorf("got %+v, want %+v", cells[0], want)
	}
}

func TestRasterizeLeavesLetterboxUnset(t *testing.T) {
	img := gradient()
	l := Layout{Src: img.Bounds(), Dst: Rect{X: 1, Y: 0, W: 2, H: 2}}
	cells := Rasterize(img, l, 4, 2)
	if cells[0].Set || cells[3].Set {
		t.Error("cells outside the destination were painted")
	}
	if !cells[1].Set || !cells[2].Set {
		t.Error("cells inside the destination were left blank")
	}
	if len(cells) != 8 {
		t.Errorf("%d cells for a 4x2 region", len(cells))
	}
}
