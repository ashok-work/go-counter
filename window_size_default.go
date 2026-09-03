//go:build !darwin && !windows

package main

// initialWindowSize returns the default starting width and height for the application window.
// This is used as a fallback for operating systems other than macOS (darwin) and Windows.
func initialWindowSize() (int, int) {
	return 1200, 800
}
