package app

import (
	"image/color"

	"charm.land/lipgloss/v2"
	"github.com/Gaurav-Gosain/tuios/internal/config"
	"github.com/Gaurav-Gosain/tuios/internal/theme"
)

// A pane's cells reach the compositor with no background wherever the guest
// set none, so the host terminal's own background shows through them. That is
// what a terminal does by default and what most people expect, until they put
// a picture behind the terminal and find it showing through every pane.
// appearance.pane_background lets a pane be a solid block instead: the cells
// that came without a background are given one as the pane's layer is parsed,
// which is one pass over the layer on the frames that reparse it and nothing
// on the frames that reuse it.

// paneBackgroundFill is the colour to paint under a pane's unset cells, or nil
// to leave them transparent.
func (m *OS) paneBackgroundFill() color.Color {
	switch m.Settings.PaneBackgroundResolved() {
	case config.PaneBackgroundTheme:
		return toRGBA(theme.TerminalBg())
	case config.PaneBackgroundTransparent:
		return nil
	}
	hex, _ := m.Settings.PaneBackgroundHex()
	return toRGBA(lipgloss.Color(hex))
}

// isWindowLayer reports whether a layer id is one of the panes on this frame.
// Pane layers carry the window's id; every other layer names what it is.
func (m *OS) isWindowLayer(id string) bool {
	if id == "" {
		return false
	}
	for _, w := range m.Windows {
		if w != nil && w.ID == id {
			return true
		}
	}
	return false
}

// paneBackgroundColor is the colour the setting is producing, for the
// settings page's swatch. Transparent shows the ground the swatch sits on.
func paneBackgroundColor(ground color.Color, s *config.Settings) color.Color {
	if hex, ok := s.PaneBackgroundHex(); ok {
		return lipgloss.Color(hex)
	}
	return paneBackgroundKeywordColor(s.PaneBackgroundResolved(), ground)
}

func paneBackgroundKeywordColor(keyword string, ground color.Color) color.Color {
	if keyword == config.PaneBackgroundTheme {
		return theme.TerminalBg()
	}
	return ground
}
