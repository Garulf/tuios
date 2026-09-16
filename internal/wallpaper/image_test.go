package wallpaper

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func writePNG(t *testing.T, img image.Image) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "wall.png")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadDims(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	src.Set(0, 0, color.NRGBA{R: 200, G: 100, B: 50, A: 255})
	src.Set(1, 0, color.NRGBA{R: 255, G: 255, B: 255, A: 0})
	path := writePNG(t, src)

	img, err := Load(path, 50)
	if err != nil {
		t.Fatal(err)
	}
	if got := img.RGBAAt(0, 0); got != (color.RGBA{R: 100, G: 50, B: 25, A: 255}) {
		t.Errorf("dimmed pixel %v", got)
	}
	if got := img.RGBAAt(1, 0); got != (color.RGBA{A: 255}) {
		t.Errorf("transparent pixel should flatten to black, got %v", got)
	}
}

func TestLoadMissingFile(t *testing.T) {
	if _, err := Load(filepath.Join(t.TempDir(), "none.png"), 0); err == nil {
		t.Error("no error for a missing file")
	}
}

func TestLoadNotAnImage(t *testing.T) {
	path := filepath.Join(t.TempDir(), "wall.png")
	if err := os.WriteFile(path, []byte("not a picture"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path, 0); err == nil {
		t.Error("no error for a file that does not decode")
	}
}
