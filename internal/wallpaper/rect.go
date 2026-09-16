package wallpaper

import (
	"slices"
	"sort"
)

// Rect is a rectangle of cells: X and Y are the top-left cell, W and H the
// size. A W or H of zero or less is empty.
type Rect struct {
	X, Y, W, H int
}

// Empty reports whether the rectangle covers no cell.
func (r Rect) Empty() bool { return r.W <= 0 || r.H <= 0 }

// Intersect is the overlap of r and o, empty when they do not touch.
func (r Rect) Intersect(o Rect) Rect {
	x0, y0 := max(r.X, o.X), max(r.Y, o.Y)
	x1, y1 := min(r.X+r.W, o.X+o.W), min(r.Y+r.H, o.Y+o.H)
	if x1 <= x0 || y1 <= y0 {
		return Rect{}
	}
	return Rect{X: x0, Y: y0, W: x1 - x0, H: y1 - y0}
}

// FreeRects is region minus the union of covered, as horizontal strips.
//
// The strips are maximal in both directions: within a band between two
// window edges each uncovered run is one strip, and strips in consecutive
// bands with the same columns are merged. The order is by row then column,
// and it is the same for the same input, so a caller can compare one frame's
// answer with the last frame's and only redraw what moved.
func FreeRects(region Rect, covered []Rect) []Rect {
	if region.Empty() {
		return nil
	}
	clipped := make([]Rect, 0, len(covered))
	for _, c := range covered {
		if in := c.Intersect(region); !in.Empty() {
			clipped = append(clipped, in)
		}
	}

	edges := []int{region.Y, region.Y + region.H}
	for _, c := range clipped {
		edges = append(edges, c.Y, c.Y+c.H)
	}
	sort.Ints(edges)
	edges = slices.Compact(edges)

	var out []Rect
	var open []Rect
	for i := 0; i+1 < len(edges); i++ {
		y0, y1 := edges[i], edges[i+1]
		runs := freeRuns(region, clipped, y0, y1)
		var next []Rect
		for _, run := range runs {
			strip := Rect{X: run[0], Y: y0, W: run[1] - run[0], H: y1 - y0}
			if j := slices.IndexFunc(open, func(o Rect) bool {
				return o.X == strip.X && o.W == strip.W && o.Y+o.H == strip.Y
			}); j >= 0 {
				open[j].H += strip.H
				next = append(next, open[j])
				open = slices.Delete(open, j, j+1)
				continue
			}
			next = append(next, strip)
		}
		out = append(out, open...)
		open = next
	}
	out = append(out, open...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Y != out[j].Y {
			return out[i].Y < out[j].Y
		}
		return out[i].X < out[j].X
	})
	return out
}

// freeRuns is the uncovered column runs [x0, x1) of the band between rows y0
// and y1. Every covered rectangle either spans the whole band or none of it,
// because the band's edges are the rectangles' own.
func freeRuns(region Rect, covered []Rect, y0, y1 int) [][2]int {
	var spans [][2]int
	for _, c := range covered {
		if c.Y <= y0 && c.Y+c.H >= y1 {
			spans = append(spans, [2]int{c.X, c.X + c.W})
		}
	}
	sort.Slice(spans, func(i, j int) bool { return spans[i][0] < spans[j][0] })

	var runs [][2]int
	x := region.X
	end := region.X + region.W
	for _, s := range spans {
		if s[0] > x {
			runs = append(runs, [2]int{x, s[0]})
		}
		x = max(x, s[1])
		if x >= end {
			break
		}
	}
	if x < end {
		runs = append(runs, [2]int{x, end})
	}
	return runs
}
