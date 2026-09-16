// Package wallpaper is the arithmetic behind a picture on the desktop: where
// an image lands in a region of cells, which parts of that region are not
// under a window, which pixels of the image map to a run of cells, and how to
// draw those pixels as half-block characters where the host cannot draw
// pixels at all.
//
// It knows nothing about the app. Everything here is a value in and a value
// out, so it can be tested without a terminal and reused by both the kitty
// and the cell renderer, which is what keeps the two drawing the same picture.
package wallpaper
