package app

import (
	"image/color"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

// TestTrueColorFastPathMatchesStyleDiff pins the fast path to the slow one:
// for every transition it accepts, the bytes are exactly what StyleDiff
// builds, and the transitions it declines are the ones with something the
// fast path does not spell.
func TestTrueColorFastPathMatchesStyleDiff(t *testing.T) {
	red := color.RGBA{R: 200, G: 10, B: 10, A: 255}
	blue := color.RGBA{R: 10, G: 10, B: 200, A: 255}
	grey := color.RGBA{R: 100, G: 100, B: 100, A: 255}
	for _, tc := range []struct {
		name     string
		from, to uv.Style
		fast     bool
	}{
		{"fg and bg from nothing", uv.Style{}, uv.Style{Fg: red, Bg: blue}, true},
		{"fg only changes", uv.Style{Fg: red, Bg: blue}, uv.Style{Fg: grey, Bg: blue}, true},
		{"bg only changes", uv.Style{Fg: red, Bg: blue}, uv.Style{Fg: red, Bg: grey}, true},
		{"both change", uv.Style{Fg: red, Bg: blue}, uv.Style{Fg: blue, Bg: red}, true},
		{"attributes differ", uv.Style{Fg: red}, uv.Style{Fg: blue, Attrs: uv.AttrBold}, false},
		{"a basic colour", uv.Style{}, uv.Style{Fg: ansi.Red}, false},
		{"fg dropped to nil", uv.Style{Fg: red, Bg: blue}, uv.Style{Bg: blue}, false},
		{"bg dropped to nil while fg changes", uv.Style{Fg: red, Bg: blue}, uv.Style{Fg: grey}, false},
		{"same rgb but from a basic bg", uv.Style{Bg: ansi.Blue}, uv.Style{Fg: red, Bg: blue}, false},
		{"from a basic colour", uv.Style{Fg: ansi.Red}, uv.Style{Fg: red}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := appendTrueColorDiff(nil, &tc.from, &tc.to)
			if ok != tc.fast {
				t.Fatalf("fast path taken = %v, want %v", ok, tc.fast)
			}
			if !ok {
				return
			}
			if want := uv.StyleDiff(&tc.from, &tc.to); string(got) != want {
				t.Errorf("fast path wrote %q, StyleDiff builds %q", got, want)
			}
		})
	}
}
