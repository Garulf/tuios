//go:build unix

package app

import (
	"fmt"
	"image"
	"image/png"

	"github.com/Gaurav-Gosain/tuios/internal/config"
	"github.com/Gaurav-Gosain/tuios/internal/terminal"
	"github.com/Gaurav-Gosain/tuios/internal/wallpaper"
)

// The kitty path draws the wallpaper as a real picture. The image is uploaded
// once and then placed as one crop per uncovered strip of the desktop, so a
// pane's transparent cells never show it: the picture is simply not under
// them. The placements sit below the threshold at which the protocol draws an
// image under cells with a background of their own, so the dock, a popup or a
// notification that lands on a free strip covers it too.
//
// The placements are diffed against what is on screen. A frame that moved
// nothing emits nothing, which is what keeps an idle desktop idle.

const (
	// wallpaperImageID is the kitty image id the picture lives under. It is
	// above the launcher's icon range, which counts up from 0xF000_0000, and
	// far above anything a pane's own images reach.
	wallpaperImageID uint32 = 0xF800_0000
	// wallpaperZ is one below the protocol's INT32_MIN/2 line: images there are
	// drawn under any cell that has a background colour of its own.
	wallpaperZ = -1073741825
)

// wallpaperPlacement is one strip of the picture on screen: the pixels shown
// and the cells they fill.
type wallpaperPlacement struct {
	src image.Rectangle
	dst wallpaper.Rect
}

// wallpaperKitty is what the host holds: which picture, whether it survived
// the last resize, and where each strip is.
type wallpaperKitty struct {
	uploaded     wallpaperKey
	sent         bool
	encoded      []byte
	hostW, hostH int
	shown        map[uint32]wallpaperPlacement
}

// wallpaperUsesKitty reports whether this frame's wallpaper goes through the
// graphics protocol rather than the cell renderer.
func (m *OS) wallpaperUsesKitty() bool {
	if m.PostRenderWriter == nil || m.BrowserClient {
		return false
	}
	switch m.wallpaperConfig().RendererName() {
	case config.WallpaperRendererCells:
		return false
	case config.WallpaperRendererKitty:
		return true
	}
	caps := m.hostCaps()
	return caps != nil && caps.KittyGraphics && caps.CellWidth > 0 && caps.CellHeight > 0
}

// wallpaperCoveredRects is the panes on screen this frame, as the rectangles
// the picture must stay out from under. It follows the rules GetCanvas draws
// by: this workspace, not minimized, and only the zoomed pane and popups when
// one is zoomed.
func (m *OS) wallpaperCoveredRects() []wallpaper.Rect {
	zoomed := m.zoomedWindow()
	var covered []wallpaper.Rect
	for _, w := range m.Windows {
		if w == nil || w.Workspace != m.CurrentWorkspace || w.Minimized {
			continue
		}
		if zoomed != nil && w != zoomed && !w.IsPopup {
			continue
		}
		covered = append(covered, windowRect(w))
	}
	return covered
}

func windowRect(w *terminal.Window) wallpaper.Rect {
	return wallpaper.Rect{X: w.X, Y: w.Y, W: w.Width, H: w.Height}
}

// wallpaperPlacements is the strips to draw this frame.
func (m *OS) wallpaperPlacements(pixels *image.RGBA) []wallpaperPlacement {
	region := m.wallpaperRegion()
	if region.Empty() {
		return nil
	}
	layout := m.wallpaperLayout(pixels, region)
	var out []wallpaperPlacement
	for _, strip := range wallpaper.FreeRects(region, m.wallpaperCoveredRects()) {
		src, dst, ok := wallpaper.Crop(layout, strip)
		if !ok {
			continue
		}
		out = append(out, wallpaperPlacement{src: src, dst: dst})
	}
	return out
}

// flushWallpaperForFrame puts the frame's strips on the host, after the frame
// itself so the cells are already there to be drawn under. hidden takes every
// strip down and keeps the upload, for a resize drag or an overlay that has
// the screen.
func (m *OS) flushWallpaperForFrame(hidden bool) {
	k := &m.wallpaper.kitty
	pixels := m.wallpaperPixels()
	var buf []byte
	if pixels == nil || !m.wallpaperUsesKitty() {
		buf = k.clear()
	} else {
		want := m.wallpaperPlacements(pixels)
		buf = k.frame(m.wallpaper.key, pixels, want, hidden, m.Width, m.Height)
	}
	if len(buf) == 0 || m.PostRenderWriter == nil {
		return
	}
	m.PostRenderWriter.QueuePostRender(wrapSync(buf))
}

// frame is the escapes that take the host from what it shows to want.
//
// A resize is treated as the host having lost the picture. The screenshot
// preview found that a kitty resized under a placed image did not keep what
// was under the id, so the picture is uploaded again afterwards. One upload
// per resize, at a human's pace, against a desktop that is otherwise wrong
// until the next config change.
func (k *wallpaperKitty) frame(key wallpaperKey, pixels *image.RGBA, want []wallpaperPlacement, hidden bool, hostW, hostH int) []byte {
	var buf []byte
	if k.sent && (k.hostW != hostW || k.hostH != hostH) {
		k.sent = false
		k.shown = nil
	}
	if !k.sent || k.uploaded != key {
		if k.sent {
			buf = appendKittyDeleteImage(buf, wallpaperImageID)
		}
		if k.uploaded != key || k.encoded == nil {
			k.encoded = encodePNG(pixels)
			k.uploaded = key
		}
		buf = appendKittyTransmitPNG(buf, wallpaperImageID, k.encoded)
		k.sent = true
		k.hostW, k.hostH = hostW, hostH
		k.shown = nil
	}
	if hidden {
		want = nil
	}
	next := make(map[uint32]wallpaperPlacement, len(want))
	for i, p := range want {
		id := uint32(i + 1)
		next[id] = p
		if was, drawn := k.shown[id]; drawn {
			if was == p {
				continue
			}
			buf = appendKittyUnplace(buf, wallpaperImageID, id)
		}
		buf = appendKittyPlaceCrop(buf, wallpaperImageID, id, p)
	}
	for id := range k.shown {
		if _, kept := next[id]; !kept {
			buf = appendKittyUnplace(buf, wallpaperImageID, id)
		}
	}
	k.shown = next
	return buf
}

// clear takes the picture off the host altogether, for a wallpaper that was
// turned off or handed to the cell renderer.
func (k *wallpaperKitty) clear() []byte {
	if !k.sent {
		return nil
	}
	buf := appendKittyDeleteImage(nil, wallpaperImageID)
	k.sent = false
	k.shown = nil
	return buf
}

// appendKittyPlaceCrop places a crop of a resident image at the strip's cells,
// scaled to fill them, under the text and under painted backgrounds.
func appendKittyPlaceCrop(buf []byte, id, placement uint32, p wallpaperPlacement) []byte {
	buf = append(buf, "\x1b7"...)
	buf = append(buf, fmt.Sprintf("\x1b[%d;%dH", p.dst.Y+1, p.dst.X+1)...)
	buf = append(buf, fmt.Sprintf("\x1b_Ga=p,i=%d,p=%d,x=%d,y=%d,w=%d,h=%d,c=%d,r=%d,z=%d,q=2,C=1;\x1b\\",
		id, placement, p.src.Min.X, p.src.Min.Y, p.src.Dx(), p.src.Dy(), p.dst.W, p.dst.H, wallpaperZ)...)
	buf = append(buf, "\x1b8"...)
	return buf
}

// appendKittyDeleteImage frees an image and every placement of it.
func appendKittyDeleteImage(buf []byte, id uint32) []byte {
	return append(buf, fmt.Sprintf("\x1b_Ga=d,d=I,i=%d,q=2;\x1b\\", id)...)
}

func encodePNG(img *image.RGBA) []byte {
	w := &byteWriter{}
	if png.Encode(w, img) != nil {
		return nil
	}
	return w.b
}
