//go:build unix

package app

import (
	"bytes"
	"fmt"
	"image"
	"strings"
	"testing"

	"github.com/Gaurav-Gosain/tuios/internal/config"
	"github.com/Gaurav-Gosain/tuios/internal/terminal"
	"github.com/Gaurav-Gosain/tuios/internal/wallpaper"
)

func wallStrip(x, y, w, h int) wallpaperPlacement {
	return wallpaperPlacement{src: image.Rect(x, y, x+w, y+h), dst: wallpaper.Rect{X: x, Y: y, W: w, H: h}}
}

func countEscapes(buf []byte, key string) int {
	return bytes.Count(buf, []byte(key))
}

func TestWallpaperKittyFrameDiffs(t *testing.T) {
	pixels := image.NewRGBA(image.Rect(0, 0, 4, 4))
	key := wallpaperKey{path: "a.png"}
	var k wallpaperKitty

	first := k.frame(key, pixels, []wallpaperPlacement{wallStrip(0, 0, 10, 2), wallStrip(0, 5, 10, 2)}, false, 80, 24)
	if countEscapes(first, "a=t,f=100") != 1 {
		t.Errorf("first frame uploaded %d times", countEscapes(first, "a=t,f=100"))
	}
	if countEscapes(first, "a=p,") != 2 {
		t.Errorf("first frame placed %d strips, want 2", countEscapes(first, "a=p,"))
	}
	if !strings.Contains(string(first), fmt.Sprintf("z=%d", wallpaperZ)) {
		t.Error("placements are not under painted backgrounds")
	}

	same := k.frame(key, pixels, []wallpaperPlacement{wallStrip(0, 0, 10, 2), wallStrip(0, 5, 10, 2)}, false, 80, 24)
	if len(same) != 0 {
		t.Errorf("an unchanged layout emitted %q", same)
	}

	moved := k.frame(key, pixels, []wallpaperPlacement{wallStrip(0, 0, 10, 2)}, false, 80, 24)
	if countEscapes(moved, "a=d,d=i,") != 1 || countEscapes(moved, "a=p,") != 0 {
		t.Errorf("dropping one strip emitted %q", moved)
	}

	hidden := k.frame(key, pixels, []wallpaperPlacement{wallStrip(0, 0, 10, 2)}, true, 80, 24)
	if countEscapes(hidden, "a=d,d=i,") != 1 || countEscapes(hidden, "a=t,") != 0 {
		t.Errorf("hiding emitted %q", hidden)
	}
	back := k.frame(key, pixels, []wallpaperPlacement{wallStrip(0, 0, 10, 2)}, false, 80, 24)
	if countEscapes(back, "a=p,") != 1 || countEscapes(back, "a=t,") != 0 {
		t.Errorf("unhiding emitted %q", back)
	}
}

func TestWallpaperKittyResizeUploadsAgain(t *testing.T) {
	pixels := image.NewRGBA(image.Rect(0, 0, 4, 4))
	key := wallpaperKey{path: "a.png"}
	var k wallpaperKitty
	k.frame(key, pixels, []wallpaperPlacement{wallStrip(0, 0, 10, 2)}, false, 80, 24)
	resized := k.frame(key, pixels, []wallpaperPlacement{wallStrip(0, 0, 12, 2)}, false, 100, 30)
	if countEscapes(resized, "a=t,f=100") != 1 {
		t.Errorf("a resize did not upload again: %q", resized)
	}
	if countEscapes(resized, "a=p,") != 1 {
		t.Errorf("a resize placed %d strips", countEscapes(resized, "a=p,"))
	}
}

func TestWallpaperKittyNewPictureReplacesTheOld(t *testing.T) {
	pixels := image.NewRGBA(image.Rect(0, 0, 4, 4))
	var k wallpaperKitty
	k.frame(wallpaperKey{path: "a.png"}, pixels, []wallpaperPlacement{wallStrip(0, 0, 10, 2)}, false, 80, 24)
	next := k.frame(wallpaperKey{path: "b.png"}, pixels, []wallpaperPlacement{wallStrip(0, 0, 10, 2)}, false, 80, 24)
	if countEscapes(next, "a=d,d=I,") != 1 || countEscapes(next, "a=t,f=100") != 1 || countEscapes(next, "a=p,") != 1 {
		t.Errorf("a new picture emitted %q", next)
	}
	if cleared := k.clear(); countEscapes(cleared, "a=d,d=I,") != 1 {
		t.Errorf("clear emitted %q", cleared)
	}
	if k.clear() != nil {
		t.Error("a second clear emitted something")
	}
}

func TestWallpaperPlacementsAvoidPanes(t *testing.T) {
	m := wallpaperOS(t, config.WallpaperRendererKitty)
	m.PostRenderWriter = NewPostRenderWriter(nil)
	if !m.wallpaperUsesKitty() {
		t.Fatal("renderer kitty did not use kitty")
	}
	region := m.wallpaperRegion()
	all := m.wallpaperPlacements(m.wallpaperPixels())
	if len(all) != 1 || all[0].dst != region {
		t.Fatalf("an empty desktop is %v, want one strip over %v", all, region)
	}

	w := &terminal.Window{X: region.X + 5, Y: region.Y + 2, Width: 10, Height: 4, Workspace: m.CurrentWorkspace}
	m.Windows = append(m.Windows, w)
	strips := m.wallpaperPlacements(m.wallpaperPixels())
	for _, s := range strips {
		if !s.dst.Intersect(windowRect(w)).Empty() {
			t.Errorf("strip %v is under the pane", s.dst)
		}
	}
	if len(strips) != 4 {
		t.Errorf("%d strips around one pane, want 4", len(strips))
	}

	w.Minimized = true
	if got := m.wallpaperPlacements(m.wallpaperPixels()); len(got) != 1 {
		t.Errorf("a minimized pane still covers: %v", got)
	}
}

func TestWallpaperAutoRendererNeedsKittyCaps(t *testing.T) {
	m := wallpaperOS(t, config.WallpaperRendererAuto)
	m.PostRenderWriter = NewPostRenderWriter(nil)
	m.Caps = &HostCapabilities{}
	if m.wallpaperUsesKitty() {
		t.Error("auto used kitty on a host without graphics")
	}
	m.Caps = &HostCapabilities{KittyGraphics: true, CellWidth: 10, CellHeight: 20}
	if !m.wallpaperUsesKitty() {
		t.Error("auto did not use kitty on a host with graphics")
	}
	m.BrowserClient = true
	if m.wallpaperUsesKitty() {
		t.Error("a browser client was handed raw kitty escapes")
	}
}
