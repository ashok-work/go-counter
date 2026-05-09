# Fyne Browser

Desktop webview app written in Go using `github.com/webview/webview_go`.

## Structure

- `main.go`: entrypoint
- `webview_app.go`: webview startup and JS/Go bridge
- `print.go`: shared order, PDF, and receipt logic
- `print_spooler_unix.go`: macOS/Linux printer integration
- `print_spooler_windows.go`: Windows printer integration
- `window_size_darwin.go`: macOS window sizing
- `window_size_windows.go`: Windows window sizing
- `window_size_default.go`: fallback window sizing

## Important

- Run the package with `go run .` or build it with `go build .`
- Do not use `go run main.go`
  `main.go` is only the entrypoint, and the app depends on the other Go files in the package

## Requirements

- Go `1.23.2` or newer
- CGO enabled
- A working C/C++ compiler on the target platform
- WebView runtime support on the target platform

## macOS

### Requirements

- Xcode Command Line Tools
- WebKit is provided by macOS
- `lp` / `lpstat` available for printing

### Install build tools

```bash
xcode-select --install
```

### Run

```bash
go run .
```

### Build

```bash
go build -o fyne-browser .
```

## Linux

### Requirements

- GCC / G++
- `pkg-config`
- GTK 3 development headers
- WebKit2GTK development headers
- CUPS commands `lp` / `lpstat` for printing

Example package names:

- Ubuntu/Debian:
  `build-essential pkg-config libgtk-3-dev libwebkit2gtk-4.0-dev cups-client`
- Fedora:
  `gcc gcc-c++ pkgconf-pkg-config gtk3-devel webkit2gtk4.0-devel cups-client`

### Run

```bash
go run .
```

### Build

```bash
go build -o fyne-browser .
```

## Windows

### Requirements

- Go
- WebView2 Runtime
- GCC / G++ available on `PATH`
- PowerShell

Recommended compiler setup: `MSYS2 + MinGW-w64`

### Install GCC with MSYS2

Install MSYS2, then from PowerShell:

```powershell
C:\msys64\usr\bin\bash.exe -lc "pacman -Syu --noconfirm"
C:\msys64\usr\bin\bash.exe -lc "pacman -S --needed --noconfirm mingw-w64-ucrt-x86_64-gcc mingw-w64-ucrt-x86_64-pkgconf"
```

Add this to `PATH`:

```text
C:\msys64\ucrt64\bin
```

Verify:

```powershell
gcc --version
g++ --version
```

Enable CGO for Go:

```powershell
go env -w CGO_ENABLED=1
go env -w CC=gcc
go env -w CXX=g++
```

### Run

```powershell
go run .
```

### Build

```powershell
go build -ldflags="-H windowsgui" -o fyne-browser.exe .
```

`-H windowsgui` removes the extra console window.

## Cross-Build To Windows

Cross-building to Windows from macOS or Linux requires a Windows CGO toolchain such as `mingw-w64`.

Example:

```bash
env CC=x86_64-w64-mingw32-gcc \
    CXX=x86_64-w64-mingw32-g++ \
    GOOS=windows GOARCH=amd64 CGO_ENABLED=1 \
    go build -ldflags="-H windowsgui" -o fyne-browser.exe .
```

If `gcc` is missing or `CGO_ENABLED=0`, `webview_go` will fail to build.

## Printing Notes

- macOS/Linux printing uses `lp` and printer discovery uses `lpstat`
- Windows printing uses PowerShell printer discovery and PDF print shell verbs
- The app generates PDF receipts in Go before sending them to the platform printer command

## Common Commands

Format:

```bash
gofmt -w *.go
```

Build:

```bash
go build .
```

Run:

```bash
go run .
```
