//go:build !darwin && !windows

package main

func initialWindowSize() (int, int) {
	return 1200, 800
}
