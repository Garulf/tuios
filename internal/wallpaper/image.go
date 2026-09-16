package wallpaper

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"  // decoded so a gif wallpaper is a picture rather than an error
	_ "image/jpeg" // same
	_ "image/png"  // same
	"os"
)

// Load decodes the image at path and prepares it for drawing: any
// transparency is flattened over black, and the picture is darkened by
// dimPercent. The result is what both renderers draw, so a dim of 40 looks
// the same as pixels and as cells.
func Load(path string, dimPercent int) (*image.RGBA, error) {
	f, err := os.Open(path) // #nosec G304 - the path is the user's own config
	if err != nil {
		return nil, err
	}
	defer f.Close()
	src, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	b := src.Bounds()
	if b.Dx() <= 0 || b.Dy() <= 0 {
		return nil, fmt.Errorf("%s: empty image", path)
	}
	return Prepare(src, dimPercent), nil
}

// Prepare is Load for an image already decoded.
func Prepare(src image.Image, dimPercent int) *image.RGBA {
	b := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(dst, dst.Bounds(), image.Black, image.Point{}, draw.Src)
	draw.Draw(dst, dst.Bounds(), src, b.Min, draw.Over)
	keep := 100 - min(max(dimPercent, 0), 100)
	if keep == 100 {
		return dst
	}
	for i := 0; i+3 < len(dst.Pix); i += 4 {
		dst.Pix[i] = uint8(int(dst.Pix[i]) * keep / 100)
		dst.Pix[i+1] = uint8(int(dst.Pix[i+1]) * keep / 100)
		dst.Pix[i+2] = uint8(int(dst.Pix[i+2]) * keep / 100)
	}
	return dst
}

// average is the mean colour of a pixel rectangle of an opaque RGBA image.
func average(img *image.RGBA, r image.Rectangle) color.RGBA {
	r = r.Intersect(img.Bounds())
	if r.Empty() {
		return color.RGBA{A: 255}
	}
	var sr, sg, sb, n uint64
	for y := r.Min.Y; y < r.Max.Y; y++ {
		i := img.PixOffset(r.Min.X, y)
		for x := r.Min.X; x < r.Max.X; x++ {
			sr += uint64(img.Pix[i])
			sg += uint64(img.Pix[i+1])
			sb += uint64(img.Pix[i+2])
			i += 4
			n++
		}
	}
	return color.RGBA{R: uint8(sr / n), G: uint8(sg / n), B: uint8(sb / n), A: 255}
}
