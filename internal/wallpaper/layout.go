package wallpaper

import "image"

// Mode is how the image is fitted to the region.
type Mode int

const (
	// Fill scales the image to cover the whole region and crops what hangs
	// over the edges.
	Fill Mode = iota
	// Fit scales the image to show all of it, centred, with the region's own
	// colour around it.
	Fit
)

// ParseMode maps a config name to a Mode, with fill for anything unknown.
func ParseMode(name string) Mode {
	if name == "fit" {
		return Fit
	}
	return Fill
}

// Layout is where the image lands: Src is the part of the image that is
// shown, in pixels, and Dst the cells it is drawn into.
type Layout struct {
	Src image.Rectangle
	Dst Rect
}

// Frame works out the layout of an image in a region of cols by rows cells,
// each cellW by cellH pixels. The cell size is what makes the picture keep
// its shape: a cell is about twice as tall as it is wide, so a region of
// equal columns and rows is a portrait, not a square.
//
// A host that has not said how big its cells are gets a guess of 1 by 2, the
// usual proportion.
func Frame(img image.Rectangle, cols, rows, cellW, cellH int, mode Mode) Layout {
	if cellW <= 0 || cellH <= 0 {
		cellW, cellH = 1, 2
	}
	iw, ih := img.Dx(), img.Dy()
	if iw <= 0 || ih <= 0 || cols <= 0 || rows <= 0 {
		return Layout{}
	}
	regionW, regionH := cols*cellW, rows*cellH

	if mode == Fit {
		dstCols, dstRows := cols, rows
		if iw*regionH > ih*regionW {
			dstRows = max(1, (cols*cellW*ih+iw*cellH/2)/(iw*cellH))
		} else {
			dstCols = max(1, (rows*cellH*iw+ih*cellW/2)/(ih*cellW))
		}
		return Layout{
			Src: img,
			Dst: Rect{X: (cols - dstCols) / 2, Y: (rows - dstRows) / 2, W: dstCols, H: dstRows},
		}
	}

	srcW, srcH := iw, ih
	if iw*regionH > ih*regionW {
		srcW = max(1, ih*regionW/regionH)
	} else {
		srcH = max(1, iw*regionH/regionW)
	}
	x0 := img.Min.X + (iw-srcW)/2
	y0 := img.Min.Y + (ih-srcH)/2
	return Layout{
		Src: image.Rect(x0, y0, x0+srcW, y0+srcH),
		Dst: Rect{W: cols, H: rows},
	}
}

// Crop is the pixels of the source that a run of cells shows. The cells are
// clipped to the layout's destination, and ok is false when none of them
// are on the picture.
func Crop(l Layout, cells Rect) (src image.Rectangle, dst Rect, ok bool) {
	dst = cells.Intersect(l.Dst)
	if dst.Empty() || l.Dst.Empty() {
		return image.Rectangle{}, Rect{}, false
	}
	sw, sh := l.Src.Dx(), l.Src.Dy()
	x0 := l.Src.Min.X + (dst.X-l.Dst.X)*sw/l.Dst.W
	x1 := l.Src.Min.X + (dst.X+dst.W-l.Dst.X)*sw/l.Dst.W
	y0 := l.Src.Min.Y + (dst.Y-l.Dst.Y)*sh/l.Dst.H
	y1 := l.Src.Min.Y + (dst.Y+dst.H-l.Dst.Y)*sh/l.Dst.H
	if x1 <= x0 {
		x1 = x0 + 1
	}
	if y1 <= y0 {
		y1 = y0 + 1
	}
	return image.Rect(x0, y0, x1, y1), dst, true
}
