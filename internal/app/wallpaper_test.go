package app

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/Gaurav-Gosain/tuios/internal/config"
	uv "github.com/charmbracelet/ultraviolet"
)

func wallpaperPNG(t *testing.T) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := range 4 {
		for x := range 4 {
			img.SetRGBA(x, y, color.RGBA{R: uint8(60 * x), G: uint8(60 * y), B: 128, A: 255})
		}
	}
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

// wallpaperOS is a 40x12 desktop with a wallpaper configured and decoded, the
// way it is after the decode command's message has landed.
func wallpaperOS(t *testing.T, renderer string) *OS {
	t.Helper()
	m := newNarrowOS(t, 40, 12)
	m.UserConfig.Wallpaper = config.WallpaperConfig{
		Path: wallpaperPNG(t), Mode: config.WallpaperModeFill, Renderer: renderer,
	}
	cmd := m.wallpaperDecodeCmd()
	if cmd == nil {
		t.Fatal("a configured wallpaper asked for no decode")
	}
	msg, ok := cmd().(wallpaperLoadedMsg)
	if !ok {
		t.Fatal("the decode did not answer with a wallpaperLoadedMsg")
	}
	if msg.err != nil {
		t.Fatal(msg.err)
	}
	m.applyWallpaper(msg)
	return m
}

func TestWallpaperDecodeCmdIsNilWhenOff(t *testing.T) {
	m := newNarrowOS(t, 40, 12)
	if m.wallpaperDecodeCmd() != nil {
		t.Error("an unset [wallpaper] section asked for a decode")
	}
}

func TestWallpaperDecodeCmdAsksOncePerKey(t *testing.T) {
	m := newNarrowOS(t, 40, 12)
	m.UserConfig.Wallpaper.Path = wallpaperPNG(t)
	if m.wallpaperDecodeCmd() == nil {
		t.Fatal("first ask returned nil")
	}
	if m.wallpaperDecodeCmd() != nil {
		t.Error("the same path was asked for twice")
	}
	m.UserConfig.Wallpaper.Dim = 20
	if m.wallpaperDecodeCmd() == nil {
		t.Error("a new dim did not ask for a decode")
	}
}

func TestWallpaperStaleResultIsDropped(t *testing.T) {
	m := wallpaperOS(t, config.WallpaperRendererCells)
	stale := wallpaperLoadedMsg{key: wallpaperKey{path: "/gone.png"}, pixels: image.NewRGBA(image.Rect(0, 0, 1, 1))}
	m.applyWallpaper(stale)
	if m.wallpaperPixels() == nil || m.wallpaperPixels().Bounds().Dx() != 4 {
		t.Error("a result for a path no longer configured replaced the picture")
	}
}

func TestWallpaperTurnedOffDropsThePicture(t *testing.T) {
	m := wallpaperOS(t, config.WallpaperRendererCells)
	m.UserConfig.Wallpaper.Path = ""
	if m.wallpaperDecodeCmd() != nil {
		t.Error("turning the wallpaper off asked for a decode")
	}
	if m.wallpaperPixels() != nil {
		t.Error("the picture survived being turned off")
	}
}

func TestWallpaperCellsFillTheDesktopOnly(t *testing.T) {
	m := wallpaperOS(t, config.WallpaperRendererCells)
	if m.wallpaperUsesKitty() {
		t.Fatal("renderer cells went through kitty")
	}
	canvas := &frameCanvas{Buffer: *uv.NewBuffer(m.GetRenderWidth(), m.GetRenderHeight())}
	m.blitWallpaperCells(canvas)

	region := m.wallpaperRegion()
	if region.Empty() {
		t.Fatal("no desktop region")
	}
	inside := canvas.Lines[region.Y][region.X]
	if inside.Content != "▀" || inside.Style.Fg == nil || inside.Style.Bg == nil {
		t.Errorf("desktop cell was %q with style %+v", inside.Content, inside.Style)
	}
	for y := range canvas.Lines {
		if y >= region.Y && y < region.Y+region.H {
			continue
		}
		for x, c := range canvas.Lines[y] {
			if c.Content != " " || c.Style.Bg != nil {
				t.Fatalf("a chrome row was painted at %d,%d: %q", x, y, c.Content)
			}
		}
	}
}

func TestWallpaperCellsCacheFollowsTheRegion(t *testing.T) {
	m := wallpaperOS(t, config.WallpaperRendererCells)
	canvas := &frameCanvas{Buffer: *uv.NewBuffer(m.GetRenderWidth(), m.GetRenderHeight())}
	m.blitWallpaperCells(canvas)
	first := m.wallpaper.cells.lines
	m.blitWallpaperCells(canvas)
	if &m.wallpaper.cells.lines[0][0] != &first[0][0] {
		t.Error("an unchanged frame rasterized again")
	}
	m.Width, m.EffectiveWidth = 60, 60
	canvas = &frameCanvas{Buffer: *uv.NewBuffer(60, 12)}
	m.blitWallpaperCells(canvas)
	if len(m.wallpaper.cells.lines[0]) == len(first[0]) {
		t.Error("a wider desktop reused the narrow raster")
	}
}

func TestWallpaperASCIIOnlyUsesSpaces(t *testing.T) {
	m := wallpaperOS(t, config.WallpaperRendererCells)
	m.Settings.UseASCIIOnly = true
	canvas := &frameCanvas{Buffer: *uv.NewBuffer(m.GetRenderWidth(), m.GetRenderHeight())}
	m.blitWallpaperCells(canvas)
	region := m.wallpaperRegion()
	c := canvas.Lines[region.Y][region.X]
	if c.Content != " " || c.Style.Bg == nil {
		t.Errorf("ascii cell was %q with style %+v", c.Content, c.Style)
	}
}

func TestWallpaperFitLeavesLetterboxBlank(t *testing.T) {
	m := wallpaperOS(t, config.WallpaperRendererCells)
	m.UserConfig.Wallpaper.Mode = config.WallpaperModeFit
	canvas := &frameCanvas{Buffer: *uv.NewBuffer(m.GetRenderWidth(), m.GetRenderHeight())}
	m.blitWallpaperCells(canvas)
	region := m.wallpaperRegion()
	layout := m.wallpaperLayout(m.wallpaperPixels(), region)
	if layout.Dst.W >= region.W {
		t.Skip("a square picture on this desktop is not pillarboxed")
	}
	edge := canvas.Lines[region.Y][region.X]
	if edge.Content != " " {
		t.Errorf("the pillarbox was painted: %q", edge.Content)
	}
	middle := canvas.Lines[region.Y+layout.Dst.Y][region.X+layout.Dst.X]
	if middle.Content != "▀" {
		t.Errorf("the picture was not painted: %q", middle.Content)
	}
}
