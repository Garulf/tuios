package wallpaper

import (
	"image"
	"testing"
)

func TestFrame(t *testing.T) {
	for _, tc := range []struct {
		name       string
		img        image.Rectangle
		cols, rows int
		mode       Mode
		want       Layout
	}{
		{"fill crops the sides of a wide image in a tall region",
			image.Rect(0, 0, 400, 100), 10, 10, Fill,
			Layout{Src: image.Rect(175, 0, 225, 100), Dst: Rect{W: 10, H: 10}}},
		{"fill crops the top and bottom of a tall image in a wide region",
			image.Rect(0, 0, 100, 400), 40, 10, Fill,
			Layout{Src: image.Rect(0, 175, 100, 225), Dst: Rect{W: 40, H: 10}}},
		{"fill of a matching aspect shows everything",
			image.Rect(0, 0, 200, 200), 20, 10, Fill,
			Layout{Src: image.Rect(0, 0, 200, 200), Dst: Rect{W: 20, H: 10}}},
		{"fit letterboxes a wide image",
			image.Rect(0, 0, 400, 100), 20, 10, Fit,
			Layout{Src: image.Rect(0, 0, 400, 100), Dst: Rect{X: 0, Y: 3, W: 20, H: 3}}},
		{"fit pillarboxes a tall image",
			image.Rect(0, 0, 100, 400), 40, 10, Fit,
			Layout{Src: image.Rect(0, 0, 100, 400), Dst: Rect{X: 17, Y: 0, W: 5, H: 10}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := Frame(tc.img, tc.cols, tc.rows, 10, 20, tc.mode)
			if got != tc.want {
				t.Errorf("got %+v\nwant %+v", got, tc.want)
			}
		})
	}
}

func TestFrameWithoutCellSizeAssumesTallCells(t *testing.T) {
	got := Frame(image.Rect(0, 0, 200, 200), 20, 10, 0, 0, Fill)
	if got.Src != image.Rect(0, 0, 200, 200) {
		t.Errorf("a 20x10 region of 1x2 cells is square, but the crop was %v", got.Src)
	}
}

func TestCrop(t *testing.T) {
	l := Layout{Src: image.Rect(100, 50, 300, 150), Dst: Rect{X: 5, Y: 2, W: 20, H: 10}}
	for _, tc := range []struct {
		name  string
		cells Rect
		src   image.Rectangle
		dst   Rect
		ok    bool
	}{
		{"the whole destination", Rect{X: 5, Y: 2, W: 20, H: 10}, image.Rect(100, 50, 300, 150), Rect{X: 5, Y: 2, W: 20, H: 10}, true},
		{"a strip half outside is clipped", Rect{X: 0, Y: 2, W: 15, H: 5}, image.Rect(100, 50, 200, 100), Rect{X: 5, Y: 2, W: 10, H: 5}, true},
		{"a strip off the picture", Rect{X: 30, Y: 0, W: 5, H: 5}, image.Rectangle{}, Rect{}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src, dst, ok := Crop(l, tc.cells)
			if ok != tc.ok || src != tc.src || dst != tc.dst {
				t.Errorf("got %v %v %v, want %v %v %v", src, dst, ok, tc.src, tc.dst, tc.ok)
			}
		})
	}
}
