package app

import (
	"image/color"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/Gaurav-Gosain/tuios/internal/config"
	"github.com/Gaurav-Gosain/tuios/internal/terminal"
	uv "github.com/charmbracelet/ultraviolet"
)

// composePane draws one pane layer and one overlay layer onto a canvas under
// the given pane background setting, and returns the canvas.
func composePane(t *testing.T, setting string) *frameCanvas {
	t.Helper()
	m := newNarrowOS(t, 20, 4)
	m.Settings.PaneBackground = setting
	m.Windows = append(m.Windows, &terminal.Window{ID: "pane-1", Workspace: m.CurrentWorkspace})
	styled := lipgloss.NewStyle().Background(lipgloss.Color("#ff0000")).Render("R")
	pane := lipgloss.NewLayer("ab\n" + styled).X(0).Y(0).ID("pane-1")
	overlay := lipgloss.NewLayer("zz").X(10).Y(0).ID("welcome")
	canvas := &frameCanvas{Buffer: *uv.NewBuffer(20, 4)}
	canvas.Clear()
	m.composeLayers(canvas, []*lipgloss.Layer{pane, overlay})
	return canvas
}

func TestPaneBackgroundTransparentLeavesCellsBare(t *testing.T) {
	canvas := composePane(t, config.PaneBackgroundTransparent)
	if bg := canvas.Lines[0][0].Style.Bg; bg != nil {
		t.Errorf("a transparent pane painted %v under its text", bg)
	}
}

func TestPaneBackgroundFillsPaneCellsOnly(t *testing.T) {
	canvas := composePane(t, "#112233")
	want := color.RGBA{R: 0x11, G: 0x22, B: 0x33, A: 255}
	if got := canvas.Lines[0][0].Style.Bg; got != want {
		t.Errorf("text cell background %v, want %v", got, want)
	}
	if got := canvas.Lines[1][1].Style.Bg; got != want {
		t.Errorf("padding past the line end got %v, want %v", got, want)
	}
	if got := canvas.Lines[1][0].Style.Bg; got == nil || toRGBA(got) != (color.RGBA{R: 255, A: 255}) {
		t.Errorf("a cell with its own background was overwritten: %v", got)
	}
	if got := canvas.Lines[0][10].Style.Bg; got != nil {
		t.Errorf("the overlay layer was filled: %v", got)
	}
}

func TestPaneBackgroundChangeReparsesACachedLayer(t *testing.T) {
	m := newNarrowOS(t, 20, 4)
	m.Windows = append(m.Windows, &terminal.Window{ID: "pane-1", Workspace: m.CurrentWorkspace})
	pane := lipgloss.NewLayer("ab").X(0).Y(0).ID("pane-1")
	canvas := &frameCanvas{Buffer: *uv.NewBuffer(20, 4)}

	m.Settings.PaneBackground = config.PaneBackgroundTransparent
	canvas.Clear()
	m.composeLayers(canvas, []*lipgloss.Layer{pane})
	if canvas.Lines[0][0].Style.Bg != nil {
		t.Fatal("transparent pane has a background")
	}

	m.Settings.PaneBackground = config.PaneBackgroundTheme
	canvas.Clear()
	m.composeLayers(canvas, []*lipgloss.Layer{pane})
	if canvas.Lines[0][0].Style.Bg == nil {
		t.Error("switching to theme did not repaint a layer whose string had not changed")
	}

	m.Settings.PaneBackground = config.PaneBackgroundTransparent
	canvas.Clear()
	m.composeLayers(canvas, []*lipgloss.Layer{pane})
	if canvas.Lines[0][0].Style.Bg != nil {
		t.Error("switching back to transparent kept the fill")
	}
}
