package wallpaper

import (
	"reflect"
	"testing"
)

func TestFreeRects(t *testing.T) {
	region := Rect{X: 2, Y: 1, W: 20, H: 10}
	for _, tc := range []struct {
		name    string
		covered []Rect
		want    []Rect
	}{
		{"nothing covered", nil, []Rect{region}},
		{"one window in the middle",
			[]Rect{{X: 6, Y: 3, W: 5, H: 4}},
			[]Rect{
				{X: 2, Y: 1, W: 20, H: 2},
				{X: 2, Y: 3, W: 4, H: 4},
				{X: 11, Y: 3, W: 11, H: 4},
				{X: 2, Y: 7, W: 20, H: 4},
			}},
		{"two overlapping windows",
			[]Rect{{X: 2, Y: 1, W: 10, H: 5}, {X: 8, Y: 3, W: 14, H: 8}},
			[]Rect{
				{X: 12, Y: 1, W: 10, H: 2},
				{X: 2, Y: 6, W: 6, H: 5},
			}},
		{"a window past the edge is clipped",
			[]Rect{{X: -5, Y: -5, W: 12, H: 8}},
			[]Rect{
				{X: 7, Y: 1, W: 15, H: 2},
				{X: 2, Y: 3, W: 20, H: 8},
			}},
		{"everything covered", []Rect{{X: 0, Y: 0, W: 40, H: 40}}, nil},
		{"strips with the same columns merge across bands",
			[]Rect{{X: 2, Y: 1, W: 10, H: 3}, {X: 2, Y: 4, W: 10, H: 7}},
			[]Rect{{X: 12, Y: 1, W: 10, H: 10}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := FreeRects(region, tc.covered)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got %v\nwant %v", got, tc.want)
			}
		})
	}
}

func TestFreeRectsIsDeterministic(t *testing.T) {
	region := Rect{W: 30, H: 12}
	a := []Rect{{X: 1, Y: 1, W: 5, H: 5}, {X: 10, Y: 2, W: 6, H: 3}}
	b := []Rect{a[1], a[0]}
	if !reflect.DeepEqual(FreeRects(region, a), FreeRects(region, b)) {
		t.Error("the order of the covered rectangles changed the answer")
	}
}

func TestIntersect(t *testing.T) {
	a := Rect{X: 0, Y: 0, W: 10, H: 10}
	if got := a.Intersect(Rect{X: 5, Y: 5, W: 10, H: 10}); got != (Rect{X: 5, Y: 5, W: 5, H: 5}) {
		t.Errorf("overlap %v", got)
	}
	if got := a.Intersect(Rect{X: 10, Y: 0, W: 3, H: 3}); !got.Empty() {
		t.Errorf("touching edges overlap: %v", got)
	}
}
