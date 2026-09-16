package app

import (
	"image"
	"image/color"

	tea "charm.land/bubbletea/v2"
	"github.com/Gaurav-Gosain/tuios/internal/config"
	"github.com/Gaurav-Gosain/tuios/internal/wallpaper"
	uv "github.com/charmbracelet/ultraviolet"
)

// The wallpaper is a picture on the empty desktop. Two things draw it. Where
// the host speaks the kitty graphics protocol the picture is uploaded once and
// placed, in pieces, over whatever part of the desktop no pane covers; see
// wallpaper_kitty.go. Everywhere else it is drawn as half-block cells straight
// into the frame, under the panes, which is what this file does.
//
// Both paths draw the same prepared image, so a dim of 40 looks the same as
// pixels and as cells, and both leave the panes alone: a pane's background is
// transparent by design (see internal/terminal), so the picture is only ever
// put where no pane is.

// wallpaperKey is what the decoded pixels were made from. A change in either
// is a new decode.
type wallpaperKey struct {
	path string
	dim  int
}

// wallpaperLoadedMsg carries a decoded wallpaper back to the Update goroutine.
type wallpaperLoadedMsg struct {
	key    wallpaperKey
	pixels *image.RGBA
	err    error
}

// wallpaperState is everything the wallpaper keeps between frames. It is only
// touched on the Update goroutine; the decode runs elsewhere and hands its
// result over as a message.
type wallpaperState struct {
	// asked is the key of the last decode requested, so a reload that changes
	// nothing does not decode again and a stale result can be told apart.
	asked wallpaperKey
	// key and pixels are the decoded picture. pixels is nil until a decode
	// lands, and after one fails.
	key    wallpaperKey
	pixels *image.RGBA
	// reported is the key whose failure has been shown, so a broken path is
	// said once rather than on every reload.
	reported wallpaperKey

	cells wallpaperCells
	kitty wallpaperKitty
}

// wallpaperCells is the rasterized picture for one region size, kept so a
// frame costs a row copy per line rather than a resample.
type wallpaperCells struct {
	key   wallpaperKey
	mode  string
	ascii bool
	w, h  int
	lines []uv.Line
}

// wallpaperConfig is the [wallpaper] table in force, or an empty one when no
// config has been loaded.
func (m *OS) wallpaperConfig() config.WallpaperConfig {
	if m.UserConfig == nil {
		return config.WallpaperConfig{}
	}
	return m.UserConfig.Wallpaper
}

func wallpaperKeyOf(cfg config.WallpaperConfig) wallpaperKey {
	return wallpaperKey{path: cfg.ExpandedPath(), dim: cfg.DimPercent()}
}

// wallpaperDecodeCmd asks for the configured image to be decoded, or returns
// nil when there is nothing new to decode. Reading and decoding a file is not
// for the Update goroutine, so it is a command and the result is a message.
//
// A config with no wallpaper drops the picture here, on this goroutine, so the
// next frame draws none.
func (m *OS) wallpaperDecodeCmd() tea.Cmd {
	cfg := m.wallpaperConfig()
	if !cfg.Enabled() {
		m.wallpaper.asked = wallpaperKey{}
		m.wallpaper.key = wallpaperKey{}
		m.wallpaper.pixels = nil
		return nil
	}
	key := wallpaperKeyOf(cfg)
	if key == m.wallpaper.asked {
		return nil
	}
	m.wallpaper.asked = key
	return func() tea.Msg {
		pixels, err := wallpaper.Load(key.path, key.dim)
		return wallpaperLoadedMsg{key: key, pixels: pixels, err: err}
	}
}

// applyWallpaper files a decoded picture away. A result for a key that is no
// longer wanted is dropped: the config moved on while the decode ran.
func (m *OS) applyWallpaper(msg wallpaperLoadedMsg) {
	if msg.key != m.wallpaper.asked {
		return
	}
	m.wallpaper.key = msg.key
	m.wallpaper.pixels = msg.pixels
	if msg.err != nil {
		m.wallpaper.pixels = nil
		if m.wallpaper.reported != msg.key {
			m.wallpaper.reported = msg.key
			m.ShowNotification("Wallpaper: "+msg.err.Error(), "warning", 0)
		}
	}
	m.MarkAllDirty()
}

// wallpaperPixels is the picture to draw this frame, or nil when there is
// none: no wallpaper configured, not decoded yet, or decoded for a config
// that has since changed.
func (m *OS) wallpaperPixels() *image.RGBA {
	cfg := m.wallpaperConfig()
	if !cfg.Enabled() || m.wallpaper.pixels == nil || m.wallpaper.key != wallpaperKeyOf(cfg) {
		return nil
	}
	return m.wallpaper.pixels
}

// wallpaperRegion is the desktop in screen cells: the render area less the
// sidebar and dock bands. The chrome paints over its own bands, so nothing
// outside this rectangle is the wallpaper's to draw on.
func (m *OS) wallpaperRegion() wallpaper.Rect {
	return wallpaper.Rect{
		X: m.GetLeftMargin(),
		Y: m.GetTopMargin(),
		W: m.GetContentWidth(),
		H: m.GetUsableHeight(),
	}
}

// wallpaperCellSize is the host's cell size in pixels, or zeros when it has
// not said. The layout guesses tall cells for zeros.
func (m *OS) wallpaperCellSize() (w, h int) {
	caps := m.hostCaps()
	if caps == nil {
		return 0, 0
	}
	return caps.CellWidth, caps.CellHeight
}

// wallpaperLayout is where the picture lands in the region this frame.
func (m *OS) wallpaperLayout(pixels *image.RGBA, region wallpaper.Rect) wallpaper.Layout {
	cw, ch := m.wallpaperCellSize()
	mode := wallpaper.ParseMode(m.wallpaperConfig().ModeName())
	return wallpaper.Frame(pixels.Bounds(), region.W, region.H, cw, ch, mode)
}

// blitWallpaperCells draws the picture as cells into the cleared canvas, before
// the panes are composed over it. It does nothing when the kitty path is
// drawing instead, and forgets its cache so the two never both hold a copy.
func (m *OS) blitWallpaperCells(canvas *frameCanvas) {
	pixels := m.wallpaperPixels()
	if pixels == nil || m.wallpaperUsesKitty() {
		m.wallpaper.cells = wallpaperCells{}
		return
	}
	region := m.wallpaperRegion()
	if region.Empty() {
		return
	}
	cache := &m.wallpaper.cells
	mode := m.wallpaperConfig().ModeName()
	ascii := m.Settings.UseASCIIOnly
	if cache.lines == nil || cache.key != m.wallpaper.key || cache.mode != mode ||
		cache.ascii != ascii || cache.w != region.W || cache.h != region.H {
		layout := m.wallpaperLayout(pixels, region)
		cells := wallpaper.Rasterize(pixels, layout, region.W, region.H)
		*cache = wallpaperCells{
			key: m.wallpaper.key, mode: mode, ascii: ascii, w: region.W, h: region.H,
			lines: wallpaperLines(cells, region.W, region.H, ascii),
		}
	}
	for row, line := range cache.lines {
		y := region.Y + row
		if y < 0 || y >= len(canvas.Lines) {
			continue
		}
		dst := canvas.Lines[y]
		if region.X >= len(dst) {
			continue
		}
		copy(dst[region.X:], line)
	}
}

// wallpaperLines turns rasterized cells into the rows the canvas copies. A
// half block with the top half in the foreground and the bottom in the
// background is two pixels per cell; an ASCII-only client gets a plain space
// in the mean of the two, which is one.
func wallpaperLines(cells []wallpaper.Cell, w, h int, ascii bool) []uv.Line {
	lines := make([]uv.Line, h)
	for row := range h {
		line := make(uv.Line, w)
		for col := range w {
			c := cells[row*w+col]
			if !c.Set {
				line[col] = uv.EmptyCell
				continue
			}
			if ascii {
				line[col] = uv.Cell{Content: " ", Width: 1, Style: uv.Style{Bg: mixRGBA(c.Top, c.Bottom)}}
				continue
			}
			line[col] = uv.Cell{Content: "▀", Width: 1, Style: uv.Style{Fg: c.Top, Bg: c.Bottom}}
		}
		lines[row] = line
	}
	return lines
}

func mixRGBA(a, b color.RGBA) color.RGBA {
	return color.RGBA{
		R: uint8((int(a.R) + int(b.R)) / 2),
		G: uint8((int(a.G) + int(b.G)) / 2),
		B: uint8((int(a.B) + int(b.B)) / 2),
		A: 255,
	}
}
