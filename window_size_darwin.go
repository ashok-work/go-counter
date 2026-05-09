//go:build darwin

package main

/*
#cgo darwin LDFLAGS: -framework CoreGraphics
#include <CoreGraphics/CoreGraphics.h>

static int initialWidth(void) {
	int width = (int)CGDisplayPixelsWide(CGMainDisplayID()) * 9 / 10;
	return width < 1200 ? 1200 : width;
}

static int initialHeight(void) {
	int height = (int)CGDisplayPixelsHigh(CGMainDisplayID()) * 17 / 20;
	return height < 800 ? 800 : height;
}
*/
import "C"

func initialWindowSize() (int, int) {
	return int(C.initialWidth()), int(C.initialHeight())
}
