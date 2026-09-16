package wallpaper

import (
	"image"
	"image/color"
)

// Cell is one half-block character: the upper half in Top, the lower in
// Bottom. Set is false for a cell outside the picture, which a fit layout
// leaves around the edges.
type Cell struct {
	Top, Bottom color.RGBA
	Set         bool
}

// Rasterize draws the layout's picture as cols by rows half-block cells,
// row-major. Each cell samples two rows of pixels, one per half, so the
// vertical resolution is twice the row count; each sample is the mean of
// the source pixels it stands for, which is what keeps a large image from
// turning to noise when it is shrunk to a few hundred cells.
func Rasterize(img *image.RGBA, l Layout, cols, rows int) []Cell {
	if cols <= 0 || rows <= 0 {
		return nil
	}
	cells := make([]Cell, cols*rows)
	if img == nil || l.Dst.Empty() || l.Src.Empty() {
		return cells
	}
	sw, sh := l.Src.Dx(), l.Src.Dy()
	dw, dh := l.Dst.W, l.Dst.H*2
	for row := range rows {
		for col := range cols {
			cx, cy := col-l.Dst.X, row-l.Dst.Y
			if cx < 0 || cx >= l.Dst.W || cy < 0 || cy >= l.Dst.H {
				continue
			}
			x0 := l.Src.Min.X + cx*sw/dw
			x1 := max(l.Src.Min.X+(cx+1)*sw/dw, x0+1)
			top := sample(img, x0, x1, l.Src.Min.Y, sh, dh, cy*2)
			bottom := sample(img, x0, x1, l.Src.Min.Y, sh, dh, cy*2+1)
			cells[row*cols+col] = Cell{Top: top, Bottom: bottom, Set: true}
		}
	}
	return cells
}

func sample(img *image.RGBA, x0, x1, srcY, sh, dh, half int) color.RGBA {
	y0 := srcY + half*sh/dh
	y1 := max(srcY+(half+1)*sh/dh, y0+1)
	return average(img, image.Rect(x0, y0, x1, y1))
}
