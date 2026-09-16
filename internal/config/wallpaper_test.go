package config

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

// TestWallpaperDefaultsAreTheOnesTheRegistryPublishes pins the accessors, the
// table DefaultConfig builds and the registry to one answer, the same way the
// spotlight test does.
func TestWallpaperDefaultsAreTheOnesTheRegistryPublishes(t *testing.T) {
	var unset WallpaperConfig
	if unset.Enabled() {
		t.Error("an unset [wallpaper] section is enabled")
	}
	built := DefaultConfig().Wallpaper
	for _, tc := range []struct{ path, accessor, table string }{
		{"wallpaper.path", unset.Path, built.Path},
		{"wallpaper.mode", unset.ModeName(), built.ModeName()},
		{"wallpaper.dim", strconv.Itoa(unset.DimPercent()), strconv.Itoa(built.DimPercent())},
		{"wallpaper.renderer", unset.RendererName(), built.RendererName()},
	} {
		opt, ok := LookupOption(tc.path)
		if !ok {
			t.Fatalf("%s has no registry entry", tc.path)
		}
		if tc.accessor != opt.Default {
			t.Errorf("%s resolves to %q, the registry publishes %q", tc.path, tc.accessor, opt.Default)
		}
		if tc.table != opt.Default {
			t.Errorf("%s is %q in DefaultConfig, the registry publishes %q", tc.path, tc.table, opt.Default)
		}
	}
}

func TestWallpaperAccessorsClampAndDefault(t *testing.T) {
	for _, tc := range []struct {
		name     string
		cfg      WallpaperConfig
		mode     string
		dim      int
		renderer string
	}{
		{"unknown names fall back", WallpaperConfig{Mode: "stretch", Renderer: "sixel"}, WallpaperModeFill, 0, WallpaperRendererAuto},
		{"dim clamps high", WallpaperConfig{Dim: 150}, WallpaperModeFill, 100, WallpaperRendererAuto},
		{"dim clamps low", WallpaperConfig{Dim: -5}, WallpaperModeFill, 0, WallpaperRendererAuto},
		{"valid values pass", WallpaperConfig{Mode: "fit", Dim: 30, Renderer: "cells"}, WallpaperModeFit, 30, WallpaperRendererCells},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.cfg.ModeName(); got != tc.mode {
				t.Errorf("mode %q, want %q", got, tc.mode)
			}
			if got := tc.cfg.DimPercent(); got != tc.dim {
				t.Errorf("dim %d, want %d", got, tc.dim)
			}
			if got := tc.cfg.RendererName(); got != tc.renderer {
				t.Errorf("renderer %q, want %q", got, tc.renderer)
			}
		})
	}
}

func TestWallpaperExpandedPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home directory")
	}
	cfg := WallpaperConfig{Path: "~/Pictures/wall.png"}
	if !cfg.Enabled() {
		t.Fatal("a section with a path is not enabled")
	}
	want := filepath.Join(home, "Pictures", "wall.png")
	if got := cfg.ExpandedPath(); got != want {
		t.Errorf("expanded to %q, want %q", got, want)
	}
}

func TestParseUserConfigFillsWallpaper(t *testing.T) {
	cfg, err := ParseUserConfig([]byte("[wallpaper]\npath = \"/tmp/w.png\"\n"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Wallpaper.Mode != WallpaperModeFill || cfg.Wallpaper.Renderer != WallpaperRendererAuto {
		t.Errorf("unset keys were not filled: %+v", cfg.Wallpaper)
	}
}
