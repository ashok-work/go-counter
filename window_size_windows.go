//go:build windows

package main

import "syscall"

const (
	smCxScreen = 0
	smCyScreen = 1
)

var (
	user32Proc           = syscall.NewLazyDLL("user32.dll")
	getSystemMetricsProc = user32Proc.NewProc("GetSystemMetrics")
)

func initialWindowSize() (int, int) {
	width := getSystemMetrics(smCxScreen) * 9 / 10
	height := getSystemMetrics(smCyScreen) * 17 / 20

	if width < 1200 {
		width = 1200
	}
	if height < 800 {
		height = 800
	}

	return width, height
}

func getSystemMetrics(index int) int {
	value, _, _ := getSystemMetricsProc.Call(uintptr(index))
	return int(value)
}
