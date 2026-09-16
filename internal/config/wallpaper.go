package config

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// WallpaperConfig is the [wallpaper] section: a picture drawn on the empty
// desktop, behind every pane and under the chrome.
//
// It is client-local appearance, like the theme: the picture is drawn by the
// client on its own host terminal, so two people attached to one session can
// each see a different one, or none.
type WallpaperConfig struct {
	Path     string `toml:"path"`     // image file, ~ expanded; empty means no wallpaper
	Mode     string `toml:"mode"`     // fill crops to cover the desktop, fit letterboxes (default: fill)
	Dim      int    `toml:"dim"`      // percent darkening, 0 to 100 (default: 0)
	Renderer string `toml:"renderer"` // auto, kitty or cells (default: auto)
}

// Wallpaper modes.
const (
	WallpaperModeFill = "fill"
	WallpaperModeFit  = "fit"
)

// Wallpaper renderers. Auto draws kitty graphics on a host that has them and
// half-block cells anywhere else; the other two force one path, which is how
// the cell rendering is checked on a terminal that would otherwise never show it.
const (
	WallpaperRendererAuto  = "auto"
	WallpaperRendererKitty = "kitty"
	WallpaperRendererCells = "cells"
)

// Wallpaper dim range, in percent.
const (
	WallpaperMinDim = 0
	WallpaperMaxDim = 100
)

// WallpaperModes is what the mode option accepts.
var WallpaperModes = []string{WallpaperModeFill, WallpaperModeFit}

// WallpaperRenderers is what the renderer option accepts.
var WallpaperRenderers = []string{WallpaperRendererAuto, WallpaperRendererKitty, WallpaperRendererCells}

func defaultWallpaperConfig() WallpaperConfig {
	return WallpaperConfig{
		Mode:     WallpaperModeFill,
		Renderer: WallpaperRendererAuto,
	}
}

func fillMissingWallpaper(cfg, defaultCfg *UserConfig) {
	w, d := &cfg.Wallpaper, &defaultCfg.Wallpaper
	if w.Mode == "" {
		w.Mode = d.Mode
	}
	if w.Renderer == "" {
		w.Renderer = d.Renderer
	}
}

// Enabled reports whether there is a wallpaper to draw at all.
func (w WallpaperConfig) Enabled() bool { return strings.TrimSpace(w.Path) != "" }

// ExpandedPath is the image path with ~ resolved against the process's home
// and made absolute.
func (w WallpaperConfig) ExpandedPath() string {
	path := strings.TrimSpace(w.Path)
	if path == "" {
		return ""
	}
	if strings.HasPrefix(path, "~") {
		if home, err := os.UserHomeDir(); err == nil {
			path = filepath.Join(home, strings.TrimPrefix(path, "~"))
		}
	}
	if abs, err := filepath.Abs(path); err == nil {
		return abs
	}
	return path
}

// ModeName is the mode to draw in, falling back to fill for a name this
// version does not know.
func (w WallpaperConfig) ModeName() string {
	if slices.Contains(WallpaperModes, w.Mode) {
		return w.Mode
	}
	return WallpaperModeFill
}

// DimPercent is how much the picture is darkened, clamped to the range.
func (w WallpaperConfig) DimPercent() int {
	return clampRange(w.Dim, WallpaperMinDim, WallpaperMaxDim)
}

// RendererName is the renderer to use, falling back to auto for a name this
// version does not know.
func (w WallpaperConfig) RendererName() string {
	if slices.Contains(WallpaperRenderers, w.Renderer) {
		return w.Renderer
	}
	return WallpaperRendererAuto
}
