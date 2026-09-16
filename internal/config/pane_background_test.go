package config

import "testing"

func TestPaneBackgroundResolves(t *testing.T) {
	for _, tc := range []struct{ set, want string }{
		{"", PaneBackgroundTransparent},
		{"transparent", PaneBackgroundTransparent},
		{"theme", PaneBackgroundTheme},
		{"#1a1b26", "#1a1b26"},
		{"opaque", PaneBackgroundTransparent},
	} {
		s := Settings{PaneBackground: tc.set}
		if got := s.PaneBackgroundResolved(); got != tc.want {
			t.Errorf("%q resolves to %q, want %q", tc.set, got, tc.want)
		}
	}
	s := Settings{PaneBackground: "#1a1b26"}
	if hex, ok := s.PaneBackgroundHex(); !ok || hex != "#1a1b26" {
		t.Errorf("hex literal not reported: %q %v", hex, ok)
	}
}

func TestPaneBackgroundDefaultMatchesRegistry(t *testing.T) {
	opt, ok := LookupOption("appearance.pane_background")
	if !ok {
		t.Fatal("appearance.pane_background has no registry entry")
	}
	if got := DefaultConfig().Appearance.PaneBackground; got != opt.Default {
		t.Errorf("DefaultConfig holds %q, registry publishes %q", got, opt.Default)
	}
	if got := DefaultSettings().PaneBackground; got != opt.Default {
		t.Errorf("DefaultSettings holds %q, registry publishes %q", got, opt.Default)
	}
}

func TestPaneBackgroundValidation(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Appearance.PaneBackground = "opaque"
	result := ValidateConfig(cfg)
	found := false
	for _, w := range result.Warnings {
		if w.Key == "pane_background" {
			found = true
		}
	}
	if !found {
		t.Error("a misspelled pane_background produced no warning")
	}
	cfg.Appearance.PaneBackground = "#abcdef"
	for _, w := range ValidateConfig(cfg).Warnings {
		if w.Key == "pane_background" {
			t.Error("a colour literal was warned about")
		}
	}
}
