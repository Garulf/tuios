//go:build !unix

package app

// The kitty helpers the picture path is built on live in the unix build with
// the launcher's icons, so off unix the wallpaper is always drawn as cells.

type wallpaperKitty struct{}

func (m *OS) wallpaperUsesKitty() bool { return false }

func (m *OS) flushWallpaperForFrame(bool) {}
