package app

import (
	"image"
	"image/color"
	"testing"

	"github.com/Gaurav-Gosain/tuios/internal/config"
)

func benchWallpaperOS(b *testing.B, on bool) *OS {
	m := benchOS(b, 1)
	m.UserConfig = config.DefaultConfig()
	// A small pane in the corner, so the desktop is mostly exposed.
	m.Windows[0].Width, m.Windows[0].Height = 40, 10
	if !on {
		return m
	}
	img := image.NewRGBA(image.Rect(0, 0, 3840, 2160))
	for y := 0; y < 2160; y++ {
		for x := 0; x < 3840; x++ {
			img.SetRGBA(x, y, color.RGBA{R: uint8(x / 15), G: uint8(y / 9), B: 128, A: 255})
		}
	}
	m.UserConfig.Wallpaper = config.WallpaperConfig{Path: "/bench.png", Mode: "fill", Renderer: "cells"}
	key := wallpaperKeyOf(m.UserConfig.Wallpaper)
	m.wallpaper.asked = key
	m.applyWallpaper(wallpaperLoadedMsg{key: key, pixels: img})
	return m
}

func BenchmarkWallpaperFrame(b *testing.B) {
	for _, tc := range []struct {
		name string
		on   bool
	}{{"blank-desktop", false}, {"cells-wallpaper", true}} {
		b.Run(tc.name, func(b *testing.B) {
			m := benchWallpaperOS(b, tc.on)
			var bytes int
			b.ReportAllocs()
			for b.Loop() {
				m.Windows[0].MarkContentDirty()
				canvas := m.GetCanvas(false)
				bytes = len(canvas.Render())
			}
			b.ReportMetric(float64(bytes), "bytes/frame")
		})
	}
}
